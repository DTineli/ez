package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const HXRedirect = "HX-Redirect"
const MIN_LEN_PASSWD = 4

type RegisterHandler struct {
	userStore    store.UserStore
	tenantStore  store.TenantStore
	inviteStore  store.InviteStore
	contactStore store.ContactStore
	sessionStore store.SessionStore
}

func RenderErrorPage(w http.ResponseWriter, message string) {
	w.Write([]byte("<h1>" + message + "</h1>"))
}

func NewRegisterHandler(
	userStore store.UserStore,
	tenantStore store.TenantStore,
	invite store.InviteStore,
	contact store.ContactStore,
	sessionStore store.SessionStore,
) *RegisterHandler {
	return &RegisterHandler{
		userStore:    userStore,
		tenantStore:  tenantStore,
		inviteStore:  invite,
		contactStore: contact,
		sessionStore: sessionStore,
	}
}

func NewRegisterHandlerWithService() *RegisterHandler {
	return &RegisterHandler{userStore: nil}
}

func (h *RegisterHandler) GetRegisterClientPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		RenderErrorPage(w, "Token de convite Invalido")
		return
	}

	parsedToken, err := uuid.Parse(token)
	if err != nil {
		RenderErrorPage(w, "Token de convite Invalido - no parse")
		return
	}

	invite, err := h.inviteStore.FindByID(parsedToken)
	if invite == nil {
		RenderErrorPage(w, "Token de convite Invalido no find")
		return
	}

	templates.ClientRegisterPage(invite).Render(r.Context(), w)
	return
}

type ClientRegisterInput struct {
	Name      string
	Email     string
	Password  string
	Phone     string
	Document  string
	ContactID uint
	TenantID  uint
}

func parseClientInput(r *http.Request) (*ClientRegisterInput, error) {
	password := strings.TrimSpace(r.FormValue("password"))
	confirm := strings.TrimSpace(r.FormValue("password_confirmation"))

	if password != confirm {
		return nil, fmt.Errorf("Senhas precisam ser iguais")
	}

	if len(password) < MIN_LEN_PASSWD {
		return nil, fmt.Errorf("Senha deve ter no mínimo %d caracteres", MIN_LEN_PASSWD)
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return nil, fmt.Errorf("Nome é obrigatório")
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		return nil, fmt.Errorf("Email é obrigatório")
	}

	contactID, err := strconv.Atoi(r.FormValue("contact_id"))
	if err != nil || contactID <= 0 {
		return nil, fmt.Errorf("Contato inválido")
	}

	tenantID, err := strconv.Atoi(r.FormValue("tenant_id"))
	if err != nil || tenantID <= 0 {
		return nil, fmt.Errorf("Tenant inválido")
	}

	return &ClientRegisterInput{
		Name:      name,
		Email:     email,
		Password:  password,
		Phone:     r.FormValue("phone"),
		Document:  r.FormValue("document"),
		ContactID: uint(contactID),
		TenantID:  uint(tenantID),
	}, nil
}

func (h *RegisterHandler) linkContact(contactID, tenantID, userID uint) error {
	return h.contactStore.UpdateById(
		contactID,
		tenantID,
		map[string]any{
			"user_id": userID,
		},
	)
}
func mapDBError(err error) string {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "email"):
		return "Este email já está em uso."
	case strings.Contains(msg, "phone"):
		return "Este telefone já está em uso."
	default:
		return "Erro ao criar conta. Tente novamente."
	}
}

func redirect(w http.ResponseWriter, r *http.Request, path string) {
	url := fmt.Sprintf("http://%s%s", r.Host, path)
	w.Header().Set(HXRedirect, url)
	w.WriteHeader(http.StatusOK)
}

func (h *RegisterHandler) PostRegisterClient(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeRegisterError(r, w, "Dados inválidos.")
		return
	}

	parsedToken, err := uuid.Parse(r.FormValue("invite_token"))
	if err != nil {
		writeRegisterError(r, w, "Token de convite inválido.")
		return
	}

	input, err := parseClientInput(r)
	if err != nil {
		writeRegisterError(r, w, err.Error())
		return
	}

	contact, err := h.contactStore.GetOne(input.ContactID)
	if err != nil || contact == nil {
		writeRegisterError(r, w, "Contato não encontrado.")
		return
	}

	// Busca usuário existente pelo telefone (cliente já cadastrado em outro tenant)
	user, err := h.userStore.GetUserByPhone(input.Phone)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			writeRegisterError(r, w, "Erro ao verificar usuário.")
			return
		}
		// Usuário novo: criar
		user = &store.User{
			Name:       input.Name,
			Email:      input.Email,
			Password:   input.Password,
			Phone:      input.Phone,
			Document:   input.Document,
			UserAccess: store.AccessCustomer,
			TenantID:   input.TenantID,
		}
		if err := h.userStore.CreateUser(user); err != nil {
			writeRegisterError(r, w, mapDBError(err))
			return
		}
	}

	if err := h.linkContact(input.ContactID, input.TenantID, user.ID); err != nil {
		writeRegisterError(r, w, "Erro ao vincular contato.")
		return
	}

	tenant, err := h.tenantStore.GetTenantByID(input.TenantID)
	if err != nil {
		writeRegisterError(r, w, "Erro ao criar sessão.")
		return
	}

	_ = h.inviteStore.DeleteByID(parsedToken)

	if err := h.sessionStore.CreateSession(r, w, store.Session{
		Name:           store.ClientSessionName,
		UserAccessType: store.AccessCustomer,
		UserID:         user.ID,
		UserName:       user.Name,
		UserEmail:      user.Email,
		TenantID:       tenant.ID,
		TenantSlug:     tenant.Slug,
		ContactInfo: &store.ContactInfo{
			ID:         contact.ID,
			PriceTable: contact.PriceTableID,
		},
	}); err != nil {
		writeRegisterError(r, w, "Erro ao criar sessão.")
		return
	}

	w.Header().Set(HXRedirect, "/client/items")
	w.WriteHeader(http.StatusOK)
}

func (h *RegisterHandler) GetRegisterPage(w http.ResponseWriter, r *http.Request) {
	err := templates.RegisterPage().Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func (h *RegisterHandler) PostRegister(w http.ResponseWriter, r *http.Request) {
	if h.userStore == nil {
		http.Error(w, "user store not configured", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeRegisterError(r, w, "Dados inválidos.")
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	password := r.FormValue("password")
	password_confirmation := r.FormValue("password_confirmation")

	if name == "" {
		writeRegisterError(r, w, "Nome é obrigatório.")
		return
	}

	if email == "" {
		writeRegisterError(r, w, "Email é obrigatório.")
		return
	}

	if slug == "" {
		writeRegisterError(r, w, "Slug é obrigatório.")
		return
	}

	if password != password_confirmation {
		writeRegisterError(r, w, "senhas nao batem")
		return
	}

	if len(password) < MIN_LEN_PASSWD {
		writeRegisterError(r, w, fmt.Sprintf("Senha deve ter no mínimo %v caracteres.", MIN_LEN_PASSWD))
		return
	}

	tenantID, err := h.tenantStore.CreateTenant(store.Tenant{
		Slug:     slug,
		Document: "",
	})

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "Duplicate") {
			writeRegisterError(r, w, "Este Slug já está em uso.")
			return
		}
		writeRegisterError(r, w, "Erro ao criar conta. Tente novamente.")
		return
	}

	err = h.userStore.CreateUser(&store.User{
		Name:     name,
		Email:    email,
		TenantID: tenantID,
		Password: password,
	})

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "Duplicate") {
			writeRegisterError(r, w, "Este email já está em uso.")
			return
		}
		writeRegisterError(r, w, "Erro ao criar conta. Tente novamente.")
		return
	}

	url := fmt.Sprintf("http://%s/admin", r.Host)

	w.Header().Set(HXRedirect, url)
	w.WriteHeader(http.StatusOK)
}

func writeRegisterError(r *http.Request, w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.WriteHeader(http.StatusUnprocessableEntity)

	_ = templates.RegisterErrors(message).Render(r.Context(), w)
}
