# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make dev            # Run with hot-reload (air)
make build          # Full production build (Tailwind + templ + binary → ./bin/ez)
make test           # Run tests: go test -race -v -timeout 30s ./...
make vet            # go vet
make staticcheck    # staticcheck linter

make templ-generate # Regenerate *_templ.go from *.templ files
make templ-watch    # Watch and regenerate templates on change
make tailwind-watch # Watch and rebuild CSS on change
```

**Important**: After editing any `.templ` file, run `make templ-generate` (or `make templ-watch` in dev). The `*_templ.go` files are generated — never edit them directly.

## Architecture

Multi-tenant e-commerce/order management app with two separate user roles: **admin** and **client** (customer), each with their own session, login page, and route group.

**Multi-tenancy** is determined by hostname slug: e.g., `company.localhost` → tenant slug `"company"`. All data (products, contacts, orders) is filtered by `TenantID`.

### Route Groups
- `/admin/*` — Admin dashboard: products, price tables, contacts, orders
- `/client/*` — Customer-facing: product listing, cart, checkout/order confirmation

### Key Packages

| Path | Role |
|------|------|
| `cmd/main.go` | Entry point, chi router setup, route definitions, graceful shutdown |
| `internal/config/` | Env config via `envconfig` (PORT, DATABASE_NAME, SESSION_COOKIE_NAME) |
| `internal/middleware/` | `SessionAuthMiddleware` (protects routes), `TextHTMLMiddleware`; session stored in context key `"info"` |
| `internal/store/store.go` | Domain models (User, Tenant, Product, PriceTable, Contact, Cart, Order) + store interfaces |
| `internal/store/db/` | GORM + SQLite initialization; auto-migrates all models on startup |
| `internal/store/dbstore/` | Concrete implementations of store interfaces |
| `internal/store/cookiesotore/` | Gorilla sessions — `CreateSession`, `GetSessionInfo`, `SetCartID` |
| `internal/handlers/` | HTTP handlers injected with store interfaces |
| `internal/templates/` | Templ components and page templates |
| `static/css/` | Tailwind CSS (`input.css` → `style.css` dev / `style.min.css` prod) |

### Data Flow Highlights
- **Admin login**: email + password → bcrypt verify → admin session → `HX-Redirect /admin/`
- **Client login**: phone + password → client session with `CartID` + `ContactInfo` (PriceTable) → `HX-Redirect /client/items`
- **Cart**: `POST /client/cart/items` adds/increments items; `GET /client/confirmacao` shows cart total; `POST /client/confirmacao` converts cart to Order with item snapshots

### Session
Two named session cookies: `ez_admin_session` and `ez_client_session`. The session secret is currently hardcoded as `"VERYSECRETKEY"` in `cookiesotore/session.go` — move to env before production.

### UI
Templates use **Templ** (type-safe Go HTML). HTMX is used for partial page updates (e.g., paginated product loading, cart updates). Responses use `HX-Redirect` headers for navigation after form POSTs.

### Environment
Defaults defined in `internal/config/config.go`:
- `PORT` → `:4000`
- `DATABASE_NAME` → `ez.db` (SQLite file in project root)
- `SESSION_COOKIE_NAME` → `session`
