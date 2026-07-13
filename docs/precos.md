# Sistema de Preços — EZ

## Visão Geral

O sistema usa **preço de custo como base imutável**. O preço de venda nunca é armazenado — sempre recalculado no momento do checkout aplicando o percentual da tabela selecionada.

---

## Modelos e Estrutura

### `PriceTable`
`internal/store/productStore.go:137`
```
┌─────────────────────────────────────────────┐
│ PriceTable                                  │
├─────────────────────────────────────────────┤
│ ID         uint        PK                   │
│ Name       string      único por tenant     │
│ Percentage float64     ex: 50 = +50%        │
│ Status     bool        ativo/inativo        │
│ TenantID   uint        FK → Tenant          │
│ Prices     []ProductPrice  (não usado)      │
└─────────────────────────────────────────────┘
```

### `Variant`
`internal/store/productStore.go:85`
```
┌─────────────────────────────────────────────┐
│ Variant                                     │
├─────────────────────────────────────────────┤
│ ID         uint        PK                   │
│ SKU        string      único por tenant     │
│ CostPrice  float64     ← BASE DO CÁLCULO    │
│ Status     bool                             │
│ IsDefault  bool                             │
│ ProductID  uint        FK → Product         │
│ TenantID   uint        FK → Tenant          │
│ Prices     []ProductPrice  (não usado)      │
└─────────────────────────────────────────────┘
```

### `ProductPrice` ⚠️ existe mas não é usado
`internal/store/productStore.go:147`
```
┌─────────────────────────────────────────────┐
│ ProductPrice                                │
├─────────────────────────────────────────────┤
│ ID           uint                           │
│ Price        float64   ← preço fixo         │
│ VariantID    uint      FK → Variant         │
│ PriceTableID uint      FK → PriceTable      │
└─────────────────────────────────────────────┘
```
> Estrutura preparada para preços fixos por variante/tabela, mas nenhum handler consulta essa tabela hoje. O cálculo usa apenas `Percentage`.

### `Contact`
`internal/store/contactStore.go:25`
```
┌─────────────────────────────────────────────┐
│ Contact                                     │
├─────────────────────────────────────────────┤
│ ID          uint       PK                   │
│ Name        string                          │
│ Phone       string                          │
│ TenantID    uint       FK → Tenant          │
│ PriceTables []PriceTable  (many2many)       │
└─────────────────────────────────────────────┘
```

---

## Relacionamentos

```
Tenant
  ├── Product (N)
  │     └── Variant (N)
  │           ├── CostPrice  ← base de cálculo
  │           └── ProductPrice (N) ← não consultado
  │
  ├── PriceTable (N)
  │     ├── Percentage  ← multiplicador
  │     └── ProductPrice (N) ← não consultado
  │
  └── Contact (N)
        └── PriceTables (many2many via contact_price_tables)
              └── quais tabelas o buyer pode usar
```

### Tabela de junção `contact_price_tables`
```
contact_price_tables
┌────────────┬───────────────┐
│ contact_id │ price_table_id│
├────────────┼───────────────┤
│     1      │       2       │  ← Contact 1 acessa tabela "Varejo"
│     1      │       5       │  ← Contact 1 acessa tabela "Atacado"
│     3      │       2       │  ← Contact 3 acessa tabela "Varejo"
└────────────┴───────────────┘
```
Um contact pode ter acesso a N tabelas. O buyer escolhe qual usar no checkout.

---

## Fórmula de Preço

```
Preço Final = CostPrice × (1 + Percentage / 100)
```

**Exemplo:**
```
CostPrice  = R$ 10,00
Percentage = 50  (tabela "Varejo")

Preço Final = 10 × (1 + 50/100)
            = 10 × 1.5
            = R$ 15,00
```

Código: `internal/handlers/client_helpers.go`
```go
func applyPrice(table store.PriceTable, variant store.Variant) float64 {
    return variant.CostPrice * (1 + table.Percentage/100)
}

func applyCheckoutPrice(costPrice float64, table *store.PriceTable) float64 {
    if table == nil {
        return costPrice
    }
    return costPrice * (1 + table.Percentage/100)
}
```

---

## Fluxo Completo

```
SELLER configura
─────────────────
[Seller] cria PriceTable ("Varejo", 50%)
[Seller] vincula PriceTable ao Contact via contact_price_tables
[Seller] cadastra Produto → Variant com CostPrice


BUYER compra
─────────────────
[1] Login
    └── Session salva: ContactInfo.ID

[2] Vitrine GET /buyer/items?price_table=<id>
    ├── Buyer seleciona tabela no dropdown
    ├── FetchItems busca PriceTable pelo ID
    ├── Para cada Variant: applyPrice(table, variant)
    └── Exibe preço calculado (não salvo)

[3] Adiciona ao carrinho POST /buyer/cart/items
    ├── Salva CostPrice no CartItem  ← custo, não venda
    └── price_table NÃO é salvo no carrinho

[4] Checkout GET /buyer/checkout?price_table=<id>
    ├── Buyer seleciona (ou mantém) tabela
    ├── RenderCheckoutContent busca PriceTable
    ├── Para cada CartItem: applyCheckoutPrice(item.CostPrice, table)
    └── Exibe subtotal e total recalculados

[5] Confirmação POST /buyer/confirmacao
    └── Preço final recalculado e salvo no Order
```

---

## Estados dos Dados

| Onde              | O que é salvo      | O que NÃO é salvo |
|-------------------|--------------------|-------------------|
| `CartItem`        | `CostPrice`        | preço de venda    |
| `Order`/`OrderItem` | preço final      | —                 |
| `ProductPrice`    | `Price` (fixo)     | nunca consultado  |

---

## Pontos de Atenção

### 1. `ProductPrice` sem uso
O modelo existe para preços fixos por variante (ex: variante X custa R$20 na tabela Atacado, independente do percentual). Hoje **nenhum handler lê essa tabela**. Se quiser migrar para preços fixos por variante, o modelo já está pronto.

### 2. Tabela não salva no carrinho
O buyer precisa re-selecionar a tabela de preço no checkout. Se trocar de tabela entre adicionar ao carrinho e confirmar, o preço muda. Isso é intencional — garante que o preço final sempre reflita a tabela selecionada no momento da confirmação.

### 3. Sem tabela selecionada no checkout
`applyCheckoutPrice` com `table == nil` retorna `CostPrice` direto — sem margem. Isso só ocorre se `price_table=0` ou param ausente na URL do checkout.
