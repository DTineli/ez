# Fluxo de Status de Pedidos — Consale

> **Versão:** 1.1  
> **Última atualização:** Maio 2026  
> **Escopo:** Sistema B2B — Buyer (portal cliente) + Seller (admin panel)

---

## 1. Visão Geral

O ciclo de vida de um pedido passa por dois eixos independentes que evoluem em paralelo:

- **Status operacional** — onde o pedido está no fluxo de aprovação → separação → entrega
- **Status de pagamento** — se o pedido foi pago ou não

Um pedido só é considerado **Completo** quando **ambos** os eixos estão resolvidos.

---

## 2. Status Operacional

| Status | Código | Ator | Descrição |
|---|---|---|---|
| Pendente | `pendente` | Sistema | Pedido criado pelo buyer, aguardando revisão |
| Aprovado | `aprovado` | Seller | Seller aceitou o pedido, inicia separação |
| Em Separação | `em_separacao` | Seller | Itens sendo separados no estoque |
| Pronto | `pronto` | Seller | Separação concluída, aguardando envio/retirada |
| Em Trânsito | `em_transito` | Seller | Pedido enviado (modalidade entrega) |
| Aguardando Retirada | `aguardando_retirada` | Seller | Disponível para o buyer retirar (modalidade retirada) |
| Entregue | `entregue` | Seller | Seller confirmou entrega ou retirada pelo buyer |
| Completo | `completo` | Sistema | Pago + Entregue — arquivado |
| Cancelado | `cancelado` | Buyer / Seller | Pedido encerrado antes da conclusão |

---

## 3. Status de Pagamento

| Status | Código | Descrição |
|---|---|---|
| Pendente | `pagamento_pendente` | Ainda não pago |
| Pago | `pago` | Pagamento confirmado pelo seller |

> O controle de pagamento é **manual pelo seller** no admin panel. Não há integração com gateway neste MVP.

---

## 4. Fluxo Operacional

```
[Buyer cria pedido]
        │
        ▼
   ┌─────────┐
   │ PENDENTE │ ◄─── Visível para o seller aprovar
   └─────────┘
     │       │
     │       └──── Buyer cancela ──► CANCELADO
     │
     ▼ Seller aprova
   ┌──────────┐
   │ APROVADO │
   └──────────┘
     │       │
     │       └──── Seller cancela ──► CANCELADO
     │
     ▼ Seller inicia separação
   ┌──────────────┐
   │ EM SEPARAÇÃO │
   └──────────────┘
     │       │
     │       └──── Seller cancela ──► CANCELADO
     │
     ▼ Seller finaliza separação
   ┌────────┐
   │ PRONTO │
   └────────┘
     │              │
     ▼ Entrega      ▼ Retirada
 ┌────────────┐  ┌──────────────────────┐
 │ EM TRÂNSITO│  │ AGUARDANDO RETIRADA  │
 └────────────┘  └──────────────────────┘
     │                    │
     └─────────┬──────────┘
               ▼ Seller confirma entrega/retirada
           ┌──────────┐
           │ ENTREGUE │
           └──────────┘
               │
               ▼ (quando pagamento_status = 'pago')
           ┌──────────┐
           │ COMPLETO │ ◄─── Arquivado
           └──────────┘
```

> **Transição para COMPLETO:** ocorre automaticamente quando o pedido está `entregue` **e** o pagamento está `pago`. A ordem dos dois eventos não importa — o sistema verifica os dois critérios a cada atualização.

---

## 5. Regras de Cancelamento

| Ator | Pode cancelar quando | Observação |
|---|---|---|
| **Buyer** | Status = `pendente` | Apenas antes da aprovação do seller |
| **Seller** | Qualquer status antes de `entregue` | Seller tem controle total até a entrega |

Pedidos com status `entregue` ou `completo` **não podem ser cancelados** — devem gerar uma devolução (fora do escopo do MVP).

---

## 6. Modalidades de Entrega

O campo `modalidade_entrega` é definido pelo **seller no momento da aprovação**:

| Modalidade | Código | Fluxo após `pronto` |
|---|---|---|
| Entrega | `entrega` | `pronto` → `em_transito` → `entregue` |
| Retirada | `retirada` | `pronto` → `aguardando_retirada` → `entregue` |

> A confirmação do status `entregue` é **sempre feita pelo seller**, independente da modalidade.

---

## 7. Visibilidade por Ator

### Buyer (Portal Cliente — Next.js)
- Vê apenas **seus próprios pedidos**
- Visualiza: status operacional, itens, valor total, status de pagamento
- Ações disponíveis: cancelar pedido enquanto `pendente`
- **Não vê** pedidos de outros contatos do mesmo cliente

### Seller (Admin Panel — Go + templ + HTMX)
- Vê **todos os pedidos** de todos os clientes
- Gerencia o fluxo completo: aprovar, separar, marcar envio/retirada, confirmar entrega
- Marca pagamento como `pago`
- Pode cancelar em qualquer etapa antes de `entregue`

---

## 8. Modelo de Dados — Tabela `pedidos`

```sql
CREATE TABLE pedidos (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id          UUID NOT NULL REFERENCES clientes(id),
    contato_id          UUID NOT NULL REFERENCES contatos(id),
    tabela_preco_id     UUID REFERENCES tabelas_preco(id),

    -- Status operacional
    status              TEXT NOT NULL DEFAULT 'pendente'
                        CHECK (status IN (
                            'pendente', 'aprovado', 'em_separacao',
                            'pronto', 'em_transito', 'aguardando_retirada',
                            'entregue', 'completo', 'cancelado'
                        )),

    -- Pagamento
    pagamento_status    TEXT NOT NULL DEFAULT 'pagamento_pendente'
                        CHECK (pagamento_status IN ('pagamento_pendente', 'pago')),
    pagamento_data      TIMESTAMPTZ,                  -- quando foi marcado como pago

    -- Entrega
    modalidade_entrega  TEXT CHECK (modalidade_entrega IN ('entrega', 'retirada')),

    -- Valores
    valor_total         NUMERIC(12, 2) NOT NULL DEFAULT 0,

    -- Rastreabilidade
    criado_em           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    atualizado_em       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    arquivado_em        TIMESTAMPTZ                   -- preenchido ao ir para 'completo'
);
```

> **Nota:** O campo `tabela_preco_id` deve ser copiado do contato no momento da criação do pedido e **congelado** — mudanças futuras na tabela de preço do contato não afetam pedidos já criados.

---

## 9. Relatórios Mensais

Os relatórios cobrem o período de **D1 a último dia do mês** e são acessíveis no admin panel do seller.

### 9.1 Volume de Vendas por Cliente

| Métrica | Descrição |
|---|---|
| Total de pedidos | Contagem de pedidos `completo` + `entregue` no período |
| Valor bruto | Soma de `valor_total` dos pedidos no período |
| Ticket médio | `valor_total / contagem` |

Agrupado por `cliente_id`, ordenado por valor decrescente.

### 9.2 Pedidos por Status

Snapshot do mês: quantos pedidos terminaram em cada status.

| Status | Contagem | % do Total |
|---|---|---|
| Completo | — | — |
| Cancelado | — | — |
| Pendente (em aberto) | — | — |
| ... | — | — |

Útil para identificar gargalos operacionais (ex: muitos pedidos parados em `em_separacao`).

### 9.3 Faturamento vs Recebido (Adimplência)

| Métrica | Descrição |
|---|---|
| Faturado | Soma de `valor_total` de pedidos `entregue` + `completo` no período |
| Recebido | Soma de `valor_total` onde `pagamento_status = 'pago'` no período |
| Inadimplente | `Faturado - Recebido` |
| % Adimplência | `Recebido / Faturado * 100` |

Detalhado por cliente para facilitar a cobrança.

### 9.4 Fechamento Mensal (Cobrança)

Lista consolidada por cliente para uso na cobrança:

```
Cliente: Empresa X
─────────────────────────────────────────────────────
Pedido #001  |  Entregue 03/05  |  R$ 1.200,00  |  ✅ Pago
Pedido #007  |  Entregue 14/05  |  R$   850,00  |  ⏳ Pendente
Pedido #012  |  Entregue 28/05  |  R$ 2.100,00  |  ⏳ Pendente
─────────────────────────────────────────────────────
Total faturado:   R$ 4.150,00
Total recebido:   R$ 1.200,00
Saldo a cobrar:   R$ 2.950,00
```

> Este relatório é a base para o seller emitir cobranças no início do mês seguinte.

---

## 10. Transições de Status — Referência Rápida

```
pendente          → aprovado              (Seller)
pendente          → cancelado             (Buyer ou Seller)
aprovado          → em_separacao          (Seller)
aprovado          → cancelado             (Seller)
em_separacao      → pronto                (Seller)
em_separacao      → cancelado             (Seller)
pronto            → em_transito           (Seller — modalidade: entrega)
pronto            → aguardando_retirada   (Seller — modalidade: retirada)
em_transito       → entregue              (Seller)
aguardando_retirada → entregue            (Seller)
entregue          → completo              (Sistema — quando pago = true)

pagamento_pendente → pago                 (Seller — qualquer momento)
```

---

## 11. Lista de Separação (Expedição)

A lista de separação é gerada pelo seller no admin panel e **impressa em papel** pela expedição. Existem dois formatos, usados em conjunto.

### 11.1 Lista Individual — por Pedido

Uma folha por pedido, usada para separar e conferir os itens antes de embalar.

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
LISTA DE SEPARAÇÃO — PEDIDO #0042
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Cliente:     Empresa X
Contato:     João Silva
Modalidade:  Entrega
Data:        20/05/2026
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 [ ]  Produto A — SKU-001          Qtd: 10 cx
 [ ]  Produto B — SKU-007          Qtd:  4 un
 [ ]  Produto C — SKU-023          Qtd:  2 kg
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Separado por: ________________   Hora: ________
Conferido por: _______________   Hora: ________
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 11.2 Lista Consolidada — por Data

Uma folha única agrupando todos os pedidos aprovados do dia, ordenada por produto. Serve para a expedição saber o volume total a preparar antes de começar a separar individualmente.

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SEPARAÇÃO CONSOLIDADA — 20/05/2026
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Produto A — SKU-001
  Total: 34 cx       Pedidos: #0042, #0045, #0051

Produto B — SKU-007
  Total: 12 un       Pedidos: #0042, #0048

Produto C — SKU-023
  Total:  6 kg       Pedidos: #0042, #0044

Produto D — SKU-031
  Total: 20 un       Pedidos: #0043, #0045, #0049, #0050
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total de pedidos no lote: 6
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 11.3 Fluxo de Uso na Expedição

```
Seller aprova pedidos
        │
        ▼
Seller gera lista consolidada → imprime 1 folha
        │
        ▼
Expedição separa o volume total por produto
        │
        ▼
Seller gera listas individuais → imprime 1 folha por pedido
        │
        ▼
Expedição confere e embala cada pedido
        │
        ▼
Seller atualiza status → em_transito ou aguardando_retirada
```

### 11.4 Critério de inclusão nas listas

| Critério | Regra |
|---|---|
| Status elegível | `aprovado` ou `em_separacao` |
| Filtro padrão | Pedidos do dia atual (por `atualizado_em` da aprovação) |
| Filtro opcional | Por período (data início → data fim) |
| Ordenação consolidada | Por nome/SKU do produto |
| Ordenação individual | Por número do pedido |

### 11.5 Implementação — Endpoint de Impressão

A geração das listas é uma rota dedicada no admin Go que retorna HTML otimizado para impressão (`@media print`), sem layout de navegação:

```
GET /admin/expedicao/lista-consolidada?data=2026-05-20
GET /admin/expedicao/lista-individual/:pedido_id
GET /admin/expedicao/listas-individuais?data=2026-05-20   ← todos do dia em uma página
```

> Usar `@media print` no CSS do templ component para esconder botões e cabeçalho do admin ao imprimir (`Ctrl+P` ou botão "Imprimir" na tela). Não é necessário gerar PDF no servidor.

---

## 12. Considerações de Implementação (Go API)

- **Validação de transição:** o endpoint `PATCH /api/v1/pedidos/:id/status` deve rejeitar transições inválidas com `HTTP 422` e mensagem descritiva.
- **Trigger de arquivamento:** verificar condição `entregue + pago` tanto ao atualizar `status` quanto ao atualizar `pagamento_status`.
- **Imutabilidade de preço:** ao criar o pedido, copiar `tabela_preco_id` e calcular `valor_total` a partir da tabela vigente — nunca recalcular.
- **Soft delete:** pedidos `completo` e `cancelado` nunca são deletados do banco; apenas `arquivado_em` é preenchido.
- **Filtro de relatórios:** usar `criado_em` (ou `arquivado_em` para completos) para delimitar o período mensal.
- **Listas de separação:** queries de agregação por `itens_pedido` filtradas por status e data — sem tabela extra no banco, geradas on-the-fly.
