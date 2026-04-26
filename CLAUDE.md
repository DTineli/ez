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
Multi-tenant e-commerce/order mgmt. Roles: **admin**, **client**. Tenant = hostname slug (`company.localhost` → `"company"`). All data filtered by `TenantID`.

Routes: `/admin/*` (dashboard) · `/client/*` (storefront)

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
- Admin login: email+pass → bcrypt → session → `HX-Redirect /admin/`
- Client login: phone+pass → session (CartID + PriceTable) → `HX-Redirect /client/items`
- Cart: `POST /client/cart/items` add/increment; `POST /client/confirmacao` → Order

## Session
Cookies: `ez_admin_session`, `ez_client_session`. Secret hardcoded `"VERYSECRETKEY"` in `cookiesotore/session.go` — move to env before prod.

## UI
Templ + HTMX (partial updates, `HX-Redirect` after POSTs). Tailwind CLI only — no npm.

## Env Defaults
`PORT=:4000` · `DATABASE_NAME=ez.db` · `SESSION_COOKIE_NAME=session`
