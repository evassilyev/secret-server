package main

import (
	"github.com/evassilyev/secret-server/api/core"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
	"net/http"
	"path"
)

// Config of the application
type Config struct {
	Addr       string
	Secret     string
	Debug      bool
	ApiPrefix  string
	StaticPath string
}

// Services is a struct with implemented services
type Services struct {
	Storage core.StorageService
}

// server main application struct
type server struct {
	config      *Config
	services    *Services
	handler     http.Handler
	apiHandlers map[string]http.HandlerFunc
}

func (s *server) initConfig(v *viper.Viper) {
	s.config = new(Config)
	// Fill default values
	v.SetDefault("addr", ":8181")
	v.SetDefault("apiPrefix", "/api")
	v.SetDefault("accessTokenTTL", 15)
	v.SetDefault("refreshTokenTTL", 30)

	s.config.Addr = v.GetString("addr")
	s.config.Secret = v.GetString("secret")
	s.config.Debug = v.GetBool("debug")
	s.config.ApiPrefix = v.GetString("apiPrefix")
	s.config.StaticPath = v.GetString("staticPath")
}

func (s *server) initServices(v *viper.Viper) error {
	s.services = new(Services)

	// TODO add storage variability

	if v.Get("db") == nil {
		return errors.New("no db configuration found")
	}

	urlKey := "db.url"
	maxIdleConnsKey := "db.maxIdleConns"
	maxOpenConnsKey := "db.maxOpenConnsKey"
	v.SetDefault(maxIdleConnsKey, 2)
	v.SetDefault(maxOpenConnsKey, 0)

	db := pgdb.NewDB(v.GetString(urlKey), v.GetInt(maxIdleConnsKey), v.GetInt(maxOpenConnsKey))

	s.services.Storage = pgdb.NewStorageService(db)

	return nil
}

func (s *server) initHandler() {
	mainRouter := mux.NewRouter()
	mainRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(s.config.StaticPath, "/index.html"))
	})
	router := mainRouter.PathPrefix(s.config.ApiPrefix).Subrouter()
	router.StrictSlash(true)

	router.HandleFunc("/secret", s.SecretSaveHandler).Methods(http.MethodPost)
	router.HandleFunc("/secret/{hash}", s.SecretGetHandler).Methods(http.MethodGet)

	// Standard middleware
	recovery := negroni.NewRecovery()
	recovery.PrintStack = s.config.Debug

	handler := negroni.New(recovery, negroni.NewLogger(), s.CorsMiddleware())
	// Serving static files if configured
	if s.config.StaticPath != "" {
		static := negroni.NewStatic(http.Dir(s.config.StaticPath))
		handler.Use(static)
	}
	handler.UseHandler(mainRouter)
	s.handler = handler
}

// NewApp is a viper based constructor for the application
func ListenAndServe(v *viper.Viper) error {
	var err error
	s := new(server)
	s.initConfig(v)
	err = s.initServices(v)
	if err != nil {
		return err
	}
	s.initHandler()
	return http.ListenAndServe(s.config.Addr, s.handler)
}
