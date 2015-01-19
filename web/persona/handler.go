package persona

import (
	"log"
	"net/http"
)

func New(store emailStore, audience string, users []string) PersonaHandlers {
	return PersonaHandlers{
		SignIn:  signInHandler{store, audience},
		SignOut: signOutHandler{store},
		Protect: protectFilter{store, users}.Apply,
		Switch:  switchBranch{store, users}.Apply,
	}
}

type Filter func(http.Handler) http.Handler
type Branch func(http.Handler, http.Handler) http.Handler

type PersonaHandlers struct {
	SignIn  http.Handler
	SignOut http.Handler
	Protect Filter
	Switch  Branch
}

func isSignedIn(toCheck string, users []string) bool {
	for _, user := range users {
		if user == toCheck {
			return true
		}
	}
	return false
}

type emailStore interface {
	Set(email string, w http.ResponseWriter, r *http.Request)
	Get(r *http.Request) string
}

type switchBranch struct {
	store emailStore
	users []string
}

func (b switchBranch) Apply(good, bad http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isSignedIn(b.store.Get(r), b.users) {
			bad.ServeHTTP(w, r)
			return
		}

		good.ServeHTTP(w, r)
	})
}

type protectFilter struct {
	store emailStore
	users []string
}

func Forbidden(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "403 forbidden", http.StatusForbidden)
}

func ForbiddenHandler() http.Handler {
	return http.HandlerFunc(Forbidden)
}

func (f protectFilter) Apply(handler http.Handler) http.Handler {
	return switchBranch{f.store, f.users}.Apply(handler, ForbiddenHandler())
}

type signInHandler struct {
	store    emailStore
	audience string
}

func (s signInHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assertion := r.PostFormValue("assertion")
	email, err := Assert(s.audience, assertion)

	if err != nil {
		log.Print("persona:", err)
		w.WriteHeader(403)
		return
	}

	s.store.Set(email, w, r)
	w.WriteHeader(200)
}

type signOutHandler struct {
	store emailStore
}

func (s signOutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.store.Set("-", w, r)
	http.Redirect(w, r, "/", 307)
}
