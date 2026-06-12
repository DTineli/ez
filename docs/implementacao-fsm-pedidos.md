# Implementação do FSM de Pedidos — Guia passo a passo

> Stack: Go · chi · templ · HTMX  
> Pacote: `internal/pedido`

---

## Como os status vão funcionar

Um pedido tem **dois status independentes** que andam em paralelo:

**Status operacional** — onde o pedido está no processo físico:

```
pendente → aprovado → em_separacao → pronto → em_transito → entregue → completo
                                            ↘ aguardando_retirada ↗
```

Qualquer etapa antes de `entregue` pode ir para `cancelado`.

**Status de pagamento** — financeiro, independente do operacional:

```
pagamento_pendente → pago
```

O pedido só vira `completo` automaticamente quando os dois critérios batem ao mesmo tempo: `status == entregue` **e** `pagamento_status == pago`. A ordem não importa — pode pagar antes de entregar ou depois.

**Por que FSM?** Sem ela, cada handler precisaria checar manualmente se a transição é válida. Com ela, existe uma fonte única de verdade: a tabela de transições. Transição inválida? Retorna erro. O handler não precisa saber das regras, só chama o service.

---

## Passo a passo

### Fase 1 — Criar a estrutura de arquivos

```bash
mkdir -p internal/pedido

touch internal/pedido/fsm.go
touch internal/pedido/service.go
touch internal/pedido/handler.go
touch internal/pedido/repository.go
```

Todos os arquivos vão começar com `package pedido`. São o mesmo pacote, arquivos separados só por organização.

---

### Fase 2 — Escrever a FSM (`fsm.go`)

Este arquivo contém apenas as regras. Sem DB, sem HTTP. Pura lógica.

**2.1 Declarar os tipos**

```go
package pedido

type Status string
type Ator  string
```

Usar tipos próprios (não `string` direto) impede passar um valor errado por acidente — o compilador pega.

**2.2 Declarar as constantes de status**

```go
const (
    Pendente           Status = "pendente"
    Aprovado           Status = "aprovado"
    EmSeparacao        Status = "em_separacao"
    Pronto             Status = "pronto"
    EmTransito         Status = "em_transito"
    AguardandoRetirada Status = "aguardando_retirada"
    Entregue           Status = "entregue"
    Completo           Status = "completo"
    Cancelado          Status = "cancelado"
)

const (
    PagamentoPendente Status = "pagamento_pendente"
    Pago              Status = "pago"
)
```

**2.3 Declarar as constantes de ator**

```go
const (
    AtorSeller  Ator = "seller"
    AtorBuyer   Ator = "buyer"
    AtorSistema Ator = "sistema"
)
```

`AtorSistema` é usado nas transições internas (ex: `entregue → completo`) que ninguém de fora pode acionar diretamente via HTTP.

**2.4 Declarar a struct de transição**

```go
type Transicao struct {
    De   Status
    Para Status
    Ator Ator
}
```

**2.5 Montar a tabela de transições válidas**

```go
var transicoesValidas = []Transicao{
    {Pendente,           Aprovado,           AtorSeller},
    {Pendente,           Cancelado,          AtorBuyer},
    {Pendente,           Cancelado,          AtorSeller},
    {Aprovado,           EmSeparacao,        AtorSeller},
    {Aprovado,           Cancelado,          AtorSeller},
    {EmSeparacao,        Pronto,             AtorSeller},
    {EmSeparacao,        Cancelado,          AtorSeller},
    {Pronto,             EmTransito,         AtorSeller},
    {Pronto,             AguardandoRetirada, AtorSeller},
    {EmTransito,         Entregue,           AtorSeller},
    {AguardandoRetirada, Entregue,           AtorSeller},
    {Entregue,           Completo,           AtorSistema}, // só o sistema aciona
}
```

Esta tabela é a documentação viva do fluxo. Qualquer mudança de regra acontece aqui.

**2.6 Implementar a função de validação**

```go
func PodeTransicionar(de, para Status, ator Ator) bool {
    for _, t := range transicoesValidas {
        if t.De == de && t.Para == para && t.Ator == ator {
            return true
        }
    }
    return false
}
```

---

### Fase 3 — Criar o modelo de dados (`repository.go`)

**3.1 Escrever a migration SQL**

Crie o arquivo `migrations/003_pedidos.sql` (ou o número seguinte no seu projeto):

```sql
CREATE TABLE pedidos (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id         UUID NOT NULL REFERENCES clientes(id),
    contato_id         UUID NOT NULL REFERENCES contatos(id),
    tabela_preco_id    UUID REFERENCES tabelas_preco(id),

    status             TEXT NOT NULL DEFAULT 'pendente'
                       CHECK (status IN (
                           'pendente','aprovado','em_separacao','pronto',
                           'em_transito','aguardando_retirada',
                           'entregue','completo','cancelado'
                       )),

    pagamento_status   TEXT NOT NULL DEFAULT 'pagamento_pendente'
                       CHECK (pagamento_status IN ('pagamento_pendente','pago')),
    pagamento_data     TIMESTAMPTZ,

    modalidade_entrega TEXT CHECK (modalidade_entrega IN ('entrega','retirada')),

    valor_total        NUMERIC(12,2) NOT NULL DEFAULT 0,

    criado_em          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    atualizado_em      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    arquivado_em       TIMESTAMPTZ
);
```

**3.2 Rodar a migration**

```bash
psql $DATABASE_URL -f migrations/003_pedidos.sql
```

**3.3 Declarar a struct Go**

```go
package pedido

import "time"

type Pedido struct {
    ID                string
    ClienteID         string
    ContatoID         string
    TabelaPrecoID     string
    Status            Status
    PagamentoStatus   Status
    PagamentoData     *time.Time
    ModalidadeEntrega string
    ValorTotal        float64
    CriadoEm         time.Time
    AtualizadoEm     time.Time
    ArquivadoEm      *time.Time
    EntregueEm       *time.Time
    CanceladoEm      *time.Time
}
```

**3.4 Implementar o repository**

```go
type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) BuscarPorID(id string) (*Pedido, error) {
    // SELECT * FROM pedidos WHERE id = $1
}

func (r *Repository) Salvar(p *Pedido) error {
    // UPDATE pedidos SET status=$1, pagamento_status=$2, ... WHERE id=$3
}
```

---

### Fase 4 — Escrever o service (`service.go`)

O service orquestra: valida via FSM, aplica side effects, persiste.

**4.1 Estrutura base**

```go
package pedido

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}
```

**4.2 AtualizarStatus — o método principal**

```go
func (s *Service) AtualizarStatus(id string, para Status, ator Ator) error {
    pedido, err := s.repo.BuscarPorID(id)
    if err != nil {
        return err
    }

    if !PodeTransicionar(pedido.Status, para, ator) {
        return fmt.Errorf("transição inválida: %s → %s", pedido.Status, para)
    }

    // aplica a transição
    pedido.Status = para

    // side effects por destino
    switch para {
    case Entregue:
        now := time.Now()
        pedido.EntregueEm = &now
    case Cancelado:
        now := time.Now()
        pedido.CanceladoEm = &now
    }

    // verifica se fecha automaticamente
    s.tentarCompletar(pedido)

    return s.repo.Salvar(pedido)
}
```

**4.3 tentarCompletar — fechamento automático**

```go
// chamado sempre que status ou pagamento mudam
func (s *Service) tentarCompletar(p *Pedido) {
    if p.Status == Entregue && p.PagamentoStatus == Pago {
        p.Status = Completo
        now := time.Now()
        p.ArquivadoEm = &now
    }
}
```

Por que separado? Porque é chamado em dois lugares: quando o seller confirma entrega e quando o seller marca como pago. A lógica fica em um único lugar.

**4.4 MarcarPago**

```go
func (s *Service) MarcarPago(id string) error {
    pedido, err := s.repo.BuscarPorID(id)
    if err != nil {
        return err
    }

    now := time.Now()
    pedido.PagamentoStatus = Pago
    pedido.PagamentoData = &now

    s.tentarCompletar(pedido) // mesma verificação

    return s.repo.Salvar(pedido)
}
```

**4.5 Escrever os testes**

Crie `internal/pedido/fsm_test.go`:

```go
func TestTransicoesValidas(t *testing.T) {
    // deve funcionar
    assert.True(t, PodeTransicionar(Pendente, Aprovado, AtorSeller))
    assert.True(t, PodeTransicionar(Pendente, Cancelado, AtorBuyer))

    // não deve funcionar
    assert.False(t, PodeTransicionar(Entregue, Pendente, AtorSeller))  // não volta atrás
    assert.False(t, PodeTransicionar(Pendente, Aprovado, AtorBuyer))   // buyer não aprova
    assert.False(t, PodeTransicionar(Entregue, Completo, AtorSeller))  // só o sistema fecha
}
```

```bash
go test ./internal/pedido/...
```

---

### Fase 5 — Handler HTTP (`handler.go`) e wiring na main

**5.1 Estrutura base do handler**

```go
package pedido

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}
```

**5.2 Handler de atualização de status**

```go
func (h *Handler) AtualizarStatus(w http.ResponseWriter, r *http.Request) {
    id   := chi.URLParam(r, "id")
    para := Status(r.FormValue("status"))

    err := h.service.AtualizarStatus(id, para, AtorSeller)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnprocessableEntity) // 422
        return
    }

    // busca o pedido atualizado e retorna o componente templ pro HTMX
    pedido, _ := h.service.BuscarPorID(id)
    views.PedidoStatusBadge(pedido).Render(r.Context(), w)
}
```

**5.3 Handler de marcar pago**

```go
func (h *Handler) MarcarPago(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    err := h.service.MarcarPago(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnprocessableEntity)
        return
    }

    pedido, _ := h.service.BuscarPorID(id)
    views.PedidoStatusBadge(pedido).Render(r.Context(), w)
}
```

**5.4 Registrar as rotas na main**

```go
// cmd/server/main.go
func main() {
    db := conectarDB()

    pedidoRepo    := pedido.NewRepository(db)
    pedidoService := pedido.NewService(pedidoRepo)
    pedidoHandler := pedido.NewHandler(pedidoService)

    r := chi.NewRouter()
    r.Patch("/admin/pedidos/{id}/status", pedidoHandler.AtualizarStatus)
    r.Post("/admin/pedidos/{id}/pago",    pedidoHandler.MarcarPago)

    http.ListenAndServe(":8080", r)
}
```

**5.5 Testar com curl**

```bash
# aprovar um pedido
curl -X PATCH http://localhost:8080/admin/pedidos/UUID_AQUI/status \
  -d "status=aprovado"

# marcar como pago
curl -X POST http://localhost:8080/admin/pedidos/UUID_AQUI/pago

# transição inválida — deve retornar 422
curl -X PATCH http://localhost:8080/admin/pedidos/UUID_AQUI/status \
  -d "status=pendente"
```

---

## Resumo do fluxo de dados

```
HTMX form POST
    │
    ▼
Handler — extrai id e status do request
    │
    ▼
Service.AtualizarStatus()
    │
    ├── FSM.PodeTransicionar() → false → retorna erro → handler devolve 422
    │
    └── true → aplica status → side effects → tentarCompletar() → repo.Salvar()
                                                                        │
                                                                        ▼
                                                              UPDATE no PostgreSQL
                                                                        │
                                                                        ▼
                                                     handler renderiza templ parcial → HTMX atualiza a UI
```
