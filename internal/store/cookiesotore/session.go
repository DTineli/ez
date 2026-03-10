package cookiesotore

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/DTineli/ez/internal/store"
	"github.com/gorilla/sessions"
)

const (
	AdminSessionName  = "ez_admin_session"
	ClientSessionName = "ez_client_session"
)

type SessionStore struct {
	name  string
	store *sessions.CookieStore
}

func NewSessionStore(session_name, session_key string) *SessionStore {
	return &SessionStore{
		name:  session_name,
		store: sessions.NewCookieStore([]byte(session_key)),
	}

}

func (s *SessionStore) CreateSession(r *http.Request, w http.ResponseWriter, sessValues store.Session) error {
	sess, _ := s.store.Get(r, s.name)

	sess.Values["user_id"] = sessValues.UserID
	sess.Values["user_name"] = sessValues.UserName
	sess.Values["user_email"] = sessValues.UserEmail
	sess.Values["tenant_id"] = sessValues.TenantID
	sess.Values["tenant_slug"] = sessValues.TenantSlug

	err := sess.Save(r, w)
	if err != nil {
		fmt.Println(err)

		return err
	}

	return nil
}

func (s *SessionStore) DeleteSession(r *http.Request, w http.ResponseWriter) error {
	sess, _ := s.store.Get(r, s.name)
	sess.Options.MaxAge = -1
	return sess.Save(r, w)
}

func (s *SessionStore) GetSessionInfo(r *http.Request) (*store.Session, error) {
	sess, err := s.store.Get(r, s.name)
	if err != nil {
		return nil, err
	}

	if sess.IsNew {
		return nil, errors.New("not authenticated")
	}

	userID, ok := sess.Values["user_id"].(uint)
	if !ok {
		return nil, errors.New("invalid user_id in session")
	}

	userName, ok := sess.Values["user_name"].(string)
	if !ok {
		return nil, errors.New("invalid user_name in session")
	}

	userEmail, ok := sess.Values["user_email"].(string)
	if !ok {
		return nil, errors.New("invalid user_email in session")
	}

	tenantID, ok := sess.Values["tenant_id"].(uint)
	if !ok {
		return nil, errors.New("invalid tenant_id in session")
	}

	tenantSlug, ok := sess.Values["tenant_slug"].(string)
	if !ok {
		return nil, errors.New("invalid tenant_slug in session")
	}

	return &store.Session{
		UserID:     userID,
		UserName:   userName,
		UserEmail:  userEmail,
		TenantID:   tenantID,
		TenantSlug: tenantSlug,
	}, nil
}
