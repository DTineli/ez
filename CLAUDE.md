# CLAUDE.md

## Commands
```bash
make dev            # hot-reload (air)
make build          # Tailwind + templ + binary → ./bin/ez
make test           # go test -race -v -timeout 30s ./...
make vet / staticcheck
make templ-generate # after editing .templ files
make templ-watch / tailwind-watch
```
**Never edit `*_templ.go` directly — always regenerate.**

## Architecture
Multi-tenant e-commerce/order mgmt. Roles: **seller**, **buyer**. Tenant = hostname slug (`company.localhost` → `"company"`). All data filtered by `TenantID`.

Routes: `/seller/*` (dashboard) · `/buyer/*` (storefront)

> **Note:** Terminologia seller/buyer é a nomenclatura alvo. O código ainda usa `admin`/`client` — renomear é trabalho pendente.

## Key Packages
| Path | Role |
|------|------|
| `cmd/main.go` | chi router, routes, graceful shutdown |
| `internal/config/` | envconfig: PORT, DATABASE_NAME, SESSION_COOKIE_NAME |
| `internal/middleware/` | `SessionAuthMiddleware`, `TextHTMLMiddleware`; session in ctx key `"info"` |
| `internal/store/store.go` | models + interfaces |
| `internal/store/db/` | GORM + SQLite, auto-migrate — **toda nova struct com tags gorm deve ser adicionada ao `AutoMigrate` em `db.go`** |
| `internal/store/dbstore/` | store implementations |
| `internal/store/cookiesotore/` | Gorilla sessions: `CreateSession`, `GetSessionInfo`, `SetCartID` |
| `internal/handlers/` | HTTP handlers |
| `internal/templates/` | Templ components |
| `static/css/` | Tailwind (`input.css` → `style.min.css`) |

## Data Flow
- Seller login: email+pass → bcrypt → session → `HX-Redirect /seller/`
- Buyer login: phone+pass → session (CartID + PriceTable) → `HX-Redirect /buyer/items`
- Cart: `POST /buyer/cart/items` add/increment; `POST /buyer/confirmacao` → Order

## Session
Cookies: `ez_seller_session`, `ez_buyer_session`. Secret hardcoded `"VERYSECRETKEY"` in `cookiesotore/session.go` — move to env before prod.

## UI
Templ + HTMX (partial updates, `HX-Redirect` after POSTs). Tailwind CLI only — no npm.

## Pricing Rules

### Estrutura
- `PriceTable`: tabela de preço com `Percentage` (multiplicador) e `Prices []ProductPrice`
- `ProductPrice`: preço explícito por variante por tabela (`VariantID + PriceTableID`, unique)
- `Variant.CostPrice`: preço de custo base (usado internamente, nunca exposto ao buyer)

### Fluxo
1. **Cadastro (seller):** registra `ProductPrice.Price` por variante por tabela via `/seller/products/{id}/price-tables/{tableID}/prices`
2. **Catálogo buyer:** `FindAllByUserWithFiltersAndPriceTable` — produto só aparece se tiver `ProductPrice` na tabela do buyer (JOIN em `product_prices`); `VariantData.Price = v.Prices[0].Price` (preço explícito)
3. **Cart add:** `price = variant.CostPrice` salvo no `CartItem` (base de cálculo)
4. **Cart view:** exibe `CostPrice * (1 + Percentage/100)` via `priceTableSvc.Apply`
5. **Checkout confirm (`POST /buyer/confirmacao`):** `unitPrice = ApplyPriceTable(item.CostPrice, pt)` — recalcula com multiplicador; salva no pedido

> **Regra:** preço final = `ProductPrice.Price` da tabela escolhida pelo buyer. Checkout deve usar esse valor — não o multiplicador `CostPrice * Percentage`. `ConfirmFromCart` ainda usa `ApplyPriceTable` (pendente de correção).

## Env Defaults
`PORT=:4000` · `DATABASE_NAME=ez.db` · `SESSION_COOKIE_NAME=session`

## Migrations
`SKIP_MIGRATE=true` no env dev — evita AutoMigrate a cada restart do Air.
**Tabela não reflete struct?** Roda uma vez sem o flag:
```bash
SKIP_MIGRATE=false go run ./cmd/main.go
```
Toda nova struct com tags gorm deve ser adicionada ao `MustMigrate` em `internal/store/db/db.go`.
