package orders

type Status string
type Ator string

const (
	Pendente  Status = "pendente"
	Aprovado  Status = "aprovado"
	Completo  Status = "completo"
	Cancelado Status = "cancelado"

	EmSeparacao        Status = "em_separacao"
	Entregue           Status = "entregue"
	AguardandoRetirada Status = "aguardando_retirada"
)

const (
	PagamentoPendente Status = "pagamento_pendente"
	Pago              Status = "pago"
)

const (
	StatusConfirmed Status = "confirmed"
)

const (
	AtorSeller  Ator = "seller"
	AtorBuyer   Ator = "buyer"
	AtorSistema Ator = "sistema"
)

type Transicao struct {
	De   Status
	Para Status
	Ator Ator
}

var transicoesValidas = []Transicao{
	{Pendente, Aprovado, AtorSeller},
	{Pendente, Cancelado, AtorBuyer},
	{Pendente, Cancelado, AtorSeller},

	{Aprovado, EmSeparacao, AtorSeller},
	{Aprovado, Cancelado, AtorSeller},

	{EmSeparacao, Completo, AtorSeller},
	{EmSeparacao, Cancelado, AtorSeller},
	{EmSeparacao, AguardandoRetirada, AtorSeller},

	{AguardandoRetirada, Entregue, AtorSeller},

	{Entregue, Completo, AtorSistema},
	{Aprovado, Completo, AtorSistema},
}

func PodeTransicionar(de, para Status, ator Ator) bool {
	for _, t := range transicoesValidas {
		if t.De == de && t.Para == para && t.Ator == ator {
			return true
		}
	}
	return false
}
