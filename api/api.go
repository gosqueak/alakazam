package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gosqueak/alakazam/relay"
	kit "github.com/gosqueak/apikit"
	middlew "github.com/gosqueak/apikit/middleware"
	"github.com/gosqueak/jwt"
)

type Server struct {
	db          *sql.DB
	addr        string
	jwtAudience jwt.Audience
	eventRelay  *relay.Relay
}

func NewServer(addr string, db *sql.DB, aud jwt.Audience, msgRelay *relay.Relay) *Server {
	return &Server{db, addr, aud, msgRelay}
}

func (s *Server) ConfigureRoutes() {
	http.HandleFunc(
		"/ws",
		middlew.Log(middlew.CheckToken(kit.CookieNameAPIToken, s.jwtAudience, s.eventRelay.HandleUpgradeConnection)),
	)
}

func (s *Server) Run() {
	s.ConfigureRoutes()
	// start serving
	log.Fatal(http.ListenAndServe(s.addr, nil))
}
