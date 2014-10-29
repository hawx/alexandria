package persona

import (
	"code.google.com/p/gorilla/sessions"
	"net/http"
)

type Store struct {
	store sessions.Store
}

func NewStore(secret string) Store {
	return Store{sessions.NewCookieStore([]byte(secret))}
}

func (s Store) Get(r *http.Request) string {
	session, _ := s.store.Get(r, "session-name")

	if v, ok := session.Values["email"].(string); ok {
		return v
	}

	return ""
}

func (s Store) Set(email string, w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, "session-name")
	session.Values["email"] = email
	session.Save(r, w)
}
