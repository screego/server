package router

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/screego/server/auth"
	"github.com/screego/server/config"
	"github.com/screego/server/ui"
	"github.com/screego/server/ws"
)

type UIConfig struct {
	AuthMode string `json:"authMode"`
	User     string `json:"user"`
	LoggedIn bool   `json:"loggedIn"`
	Version  string `json:"version"`
}

func Router(conf config.Config, rooms *ws.Rooms, users *auth.Users, version string) *mux.Router {
	router := mux.NewRouter()
	router.Use(handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST"}), handlers.AllowedOriginValidator(conf.CheckOrigin)))
	router.HandleFunc("/stream", rooms.Upgrade)
	router.Methods("POST").Path("/login").HandlerFunc(users.Authenticate)
	router.Methods("POST").Path("/logout").HandlerFunc(users.Logout)
	router.Methods("GET").Path("/config").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, loggedIn := users.CurrentUser(r)
		_ = json.NewEncoder(w).Encode(&UIConfig{
			AuthMode: conf.AuthMode,
			LoggedIn: loggedIn,
			User:     user,
			Version:  version,
		})
	})

	ui.Register(router)

	return router
}
