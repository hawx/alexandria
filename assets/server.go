package assets

import (
	"net/http"
	"strings"
	"time"
)

func Server(m map[string]string) http.Handler {
	return assetServer{m, time.Now()}
}

type assetServer struct {
	assetMap  map[string]string
	createdAt time.Time
}

func (s assetServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if v := s.assetMap[r.URL.Path]; v != "" {
		http.ServeContent(w, r, r.URL.Path, s.createdAt, strings.NewReader(v))
		return
	}

	http.NotFound(w, r)
}
