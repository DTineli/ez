# Consale — Implementação: Edge Cases do Relacionamento Cliente e Loja

Documento gerado a partir das decisões de produto. Cada seção descreve o comportamento esperado, a regra de negócio e as etapas de implementação.

---

## 1. Acesso via URL

### 1.1 Acesso sem slug (`consale.com`)

**Comportamento:** Landing page com todas as lojas que o usuário tem acesso.

**Regra de negócio:**
- Se o user estiver logado, busca todos os customers vinculados a ele e exibe as lojas correspondentes.
- Se não estiver logado, exibe landing page genérica com campo de login.

**Implementação:**
- [X] Middleware de resolução de slug: ao receber requisição em subdomínio, buscar admin pelo slug
- [ ] Criar rota `/` no frontend
- [ ] Se autenticado: buscar `customers` do user logado via API → listar lojas (nome, slug, logo)
- [ ] Landing Page Publica com um bem vindo e btn para ir para login ou criar conta
- [ ] Se não autenticado: renderizar landing page pública
- [ ] Componente `LojaCard` com link para `slug.consale.com`
---

### 1.2 Slug inválido ou inexistente (`xyz.consale.com`)

**Comportamento:** Página "Loja não encontrada" com botão de redirect para `consale.com`.

**Implementação:**
- [ ] Se não encontrado: retornar página `404-loja` com mensagem amigável e botão "Ir para Consale"
- [ ] Não expor detalhes técnicos na mensagem de erro
---

### 1.3 Admin desativado

**Comportamento:** Redireciona para `consale.com`.

**Regra de negócio:** Admin com status `inactive` ou `suspended` não deve servir nenhuma página da loja.

**Implementação:**
- [ ] Middleware de resolução de slug: após encontrar o admin, verificar `status`
- [ ] Se `status !== active`: redirecionar 302 para `consale.com`
- [ ] Garantir que esse check acontece antes de qualquer renderização da loja

---

## 2. Autenticação e Sessão

### 2.1 Cliente já logado entra em slug de outro admin

**Comportamento:** Acessa direto, sessão global vale para qualquer slug.

**Regra de negócio:** O token de autenticação é global (não por loja). O vínculo com o customer é verificado após a autenticação.

**Implementação:**
- [ ] Auth token armazenado globalmente (não escopado por subdomínio/slug)
- [ ] Configurar cookies com domínio `.consale.com` para valer em todos os subdomínios
- [ ] Ao entrar num slug, verificar se o user tem customer naquele admin (ver 2.2)

---

### 2.2 Cliente logado mas sem customer no admin

**Comportamento:** Bloqueia com mensagem "aguarde liberação do administrador".

**Regra de negócio:** Acesso à loja só é permitido se existir um `customer` ativo vinculando o `user` ao `admin`.

**Implementação:**
- [ ] Após autenticação, checar `customer` onde `user_id = X AND admin_id = Y AND status = active`
- [ ] Se não encontrado: renderizar página de bloqueio com mensagem amigável
- [ ] Não expor por que o acesso foi negado (segurança)
- [ ] Não permitir navegação de produtos sem vínculo confirmado

---

### 2.3 Sessão expirada com itens no carrinho

**Comportamento:** Carrinho recuperado após login (salvo no servidor por customer).

**Regra de negócio:** O carrinho é persistido no banco vinculado ao `customer`, não à sessão.

**Implementação:**
- [ ] Tabela/entidade `cart` com FK para `customer_id`
- [ ] Itens do carrinho salvos no servidor a cada adição/remoção
- [ ] Ao fazer login e acessar a loja, carregar carrinho existente do customer
- [ ] Sem dependência de `localStorage` ou cookie para persistência do carrinho

---

## 3. Customer e Vínculo

### 3.1 Admin cadastra customer com e-mail de user já existente

**Comportamento:** Vincula silenciosamente ao user existente.

**Regra de negócio:** O sistema busca um user pelo e-mail. Se encontrado, cria o `customer` apontando para esse user. Nenhuma notificação é enviada ao admin.

**Implementação:**
- [ ] No endpoint de criação de customer: `SELECT user WHERE email = ?`
- [ ] Se encontrado: criar `customer` com `user_id` do user existente
- [ ] Se não encontrado: criar user novo (sem senha) e enviar convite por e-mail
- [ ] Constraint única: `(user_id, admin_id)` na tabela `customers`

---

### 3.2 Admin remove um customer

**Comportamento:**
- User perde acesso imediatamente
- Pedidos pendentes ficam visíveis para o admin e na página geral do user (histórico)

**Regra de negócio:** Remoção do customer revoga o acesso à loja mas preserva o histórico de pedidos.

**Implementação:**
- [ ] Soft delete no customer: campo `deleted_at` ou `status = inactive`
- [ ] Middleware de acesso à loja deve checar `status` do customer a cada requisição (não só no login)
- [ ] Pedidos com `customer_id` removido: manter registros, apenas marcar customer como inativo
- [ ] Admin continua vendo pedidos pendentes do customer removido na tela de pedidos
- [ ] User vê pedidos históricos na sua área geral, sem acesso à loja

---

### 3.3 Mesmo user com dois customers no mesmo admin

**Comportamento:** Impedido no banco com constraint única.

**Implementação:**
- [ ] `UNIQUE INDEX (user_id, admin_id)` na tabela `customers`
- [ ] Endpoint de criação de customer retorna erro claro se constraint for violada
- [ ] Frontend do admin exibe mensagem: "Este usuário já possui cadastro nesta loja"

---

## 4. Tabelas de Preço

### 4.1 Customer sem nenhuma tabela atribuída

**Comportamento:** Loja abre mas exibe mensagem "nenhum produto disponível no momento".

**Implementação:**
- [ ] Query de produtos da loja filtra por tabelas atribuídas ao customer
- [ ] Se nenhuma tabela: retornar lista vazia, não erro
- [ ] Frontend exibe estado vazio com mensagem amigável
- [ ] Não bloquear o acesso à loja, apenas não mostrar produtos

---

### 4.2 Admin remove tabela atribuída a customers ativos

**Comportamento:**
- Produtos somem imediatamente do catálogo dos customers afetados
- Pedidos já criados com aquela tabela não são afetados
- Novos pedidos não podem usar aquela tabela

**Regra de negócio:** A tabela de preço é referenciada nos itens do pedido no momento da criação. Pedidos existentes têm snapshot dos dados relevantes (preço, produto).

**Implementação:**
- [ ] Ao remover tabela: soft delete ou `status = inactive`
- [ ] Query de catálogo filtra apenas tabelas `active` atribuídas ao customer
- [ ] Itens de pedido armazenam `price_at_order` (preço no momento do pedido) — não dependem da tabela estar ativa
- [ ] Carrinho: ao finalizar pedido, validar se tabela ainda está ativa; se não, remover item e avisar o cliente
- [ ] Admin vê aviso ao remover tabela se houver customers ativos vinculados a ela

---

### 4.3 Produto em duas tabelas do mesmo customer com preços diferentes

**Comportamento:** Prevalece o preço da tabela mais recentemente atribuída ao customer.

**Regra de negócio:** A ordem de atribuição das tabelas ao customer define prioridade. A última tabela atribuída tem maior prioridade.

**Implementação:**
- [ ] Tabela `customer_price_tables` com campo `assigned_at` (timestamp de atribuição)
- [ ] Query de catálogo: ao buscar preço de um produto, se aparecer em múltiplas tabelas, ordenar por `assigned_at DESC` e usar o primeiro resultado
- [ ] Documentar essa regra no painel do admin (tooltip ou help text na tela de atribuição de tabelas)

---

## 5. Carrinho e Pedido

### 5.1 Preço alterado enquanto item está no carrinho

**Comportamento:** Preço atualiza automaticamente no carrinho.

**Regra de negócio:** O carrinho sempre reflete o preço atual da tabela. O preço é "congelado" apenas no momento em que o pedido é criado.

**Implementação:**
- [ ] Carrinho não armazena preço — armazena apenas `product_id` e `quantity`
- [ ] Ao renderizar o carrinho: buscar preço atual do produto via tabela ativa do customer
- [ ] Ao criar o pedido: fazer snapshot do preço atual e salvar em `order_items.price_at_order`
- [ ] Opcional: exibir badge "preço atualizado" se o preço mudou desde a última visualização do carrinho

---

### 5.2 Pedido sem aprovação por muito tempo

**Comportamento:**
- Pedido não expira, fica aberto indefinidamente
- Cliente pode cancelar manualmente enquanto aguarda

**Implementação:**
- [ ] Status do pedido: `pending`, `approved`, `rejected`, `cancelled`
- [ ] Endpoint `PATCH /orders/:id/cancel` disponível para o cliente enquanto status = `pending`
- [ ] Sem job de expiração automática
- [ ] Admin vê pedidos pendentes ordenados por data (mais antigos primeiro) para facilitar gestão
- [ ] Considerar notificação/badge para admin se pedido estiver pendente há mais de X dias (melhoria futura)

---

### 5.3 Admin rejeita pedido sem motivo

**Comportamento:**
- Campo de motivo é opcional
- Se preenchido, é exibido ao cliente
- Se não preenchido, cliente vê apenas "pedido rejeitado"

**Implementação:**
- [ ] Campo `rejection_reason TEXT NULL` na tabela `orders`
- [ ] Endpoint de rejeição aceita `reason` opcional no body
- [ ] Frontend do admin: textarea opcional "Motivo da rejeição (visível ao cliente)"
- [ ] Frontend do cliente: exibir motivo se `rejection_reason` não for nulo

---

### 5.4 Dois pedidos simultâneos

**Comportamento:** Permitido — múltiplos pedidos em aberto ao mesmo tempo.

**Implementação:**
- [ ] Sem validação de pedido único pendente por customer
- [ ] Listar todos os pedidos com status `pending` na área do cliente, separados por loja
- [ ] Admin vê todos os pedidos pendentes do customer, independente de quantidade

---

## Resumo — Pontos de atenção na implementação

| Prioridade | Item | Motivo |
|---|---|---|
| 🔴 Alta | Cookie com domínio `.consale.com` (2.1) | Afeta toda a autenticação multi-slug |
| 🔴 Alta | Middleware de slug com check de status do admin (1.3) | Segurança e consistência |
| 🔴 Alta | Constraint única `(user_id, admin_id)` (3.3) | Integridade do banco |
| 🔴 Alta | `price_at_order` snapshot no pedido (4.2, 5.1) | Evita inconsistência financeira |
| 🟡 Média | Soft delete em customers e tabelas (3.2, 4.2) | Preserva histórico |
| 🟡 Média | Regra de prioridade por `assigned_at` (4.3) | Comportamento previsível pro cliente |
| 🟡 Média | Carrinho sem preço armazenado (5.1) | Simplicidade e consistência |
| 🟢 Baixa | Badge "preço atualizado" no carrinho (5.1) | UX, pode vir depois |
| 🟢 Baixa | Notificação de pedido antigo pendente pro admin (5.2) | Melhoria futura |
