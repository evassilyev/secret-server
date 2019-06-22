package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/evassilyev/secret-server/api/core"
	"github.com/evassilyev/secret-server/api/dev"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

// Config of the application
type Config struct {
	Addr      string
	Secret    string
	Debug     bool
	ApiPrefix string
}

// Services is a struct with implemented services
type Services struct {
	Secret core.SecretService
}

// server main application struct
type server struct {
	config   *Config
	services *Services
	handler  http.Handler
}

func (s *server) initConfig(v *viper.Viper) {
	s.config = new(Config)
	// Fill default values
	v.SetDefault("addr", ":8181")
	v.SetDefault("apiPrefix", "/api")

	s.config.Addr = v.GetString("addr")
	s.config.Debug = v.GetBool("debug")
	s.config.ApiPrefix = v.GetString("apiPrefix")
}

func (s *server) initServices(v *viper.Viper) error {
	s.services = new(Services)

	if v.Get("storage") == nil {
		return errors.New("no storage configuration found")
	}

	switch v.GetString("storage") {
	case "dev":
		s.services.Secret = dev.NewSecretService()
	case "pgdb":
		s.services.Secret = nil
		// TODO
		if v.Get("pgdb") == nil {
			return errors.New("no db configuration found")
		}

		urlKey := "db.url"
		maxIdleConnsKey := "db.maxIdleConns"
		maxOpenConnsKey := "db.maxOpenConnsKey"
		v.SetDefault(maxIdleConnsKey, 2)
		v.SetDefault(maxOpenConnsKey, 0)

		db := pgdb.NewDB(v.GetString(urlKey), v.GetInt(maxIdleConnsKey), v.GetInt(maxOpenConnsKey))

		s.services.Secret = pgdb.NewSecretService(db)
	case "redis":
		s.services.Secret = nil
		// TODO
		fmt.Println("REALIZE POSTGRES STORAGE")
	default:
		return errors.New("Wrong storage configuration")
	}
	return nil
}

func (s *server) initHandler() {
	mainRouter := mux.NewRouter()
	router := mainRouter.PathPrefix(s.config.ApiPrefix).Subrouter()
	router.StrictSlash(true)

	router.HandleFunc("/secret", s.secretSaveHandler).Methods(http.MethodPost)
	router.HandleFunc("/secret/{hash}", s.secretGetHandler).Methods(http.MethodGet)

	// Standard middleware
	recovery := negroni.NewRecovery()
	recovery.PrintStack = s.config.Debug

	handler := negroni.New(recovery, negroni.NewLogger() /*, s.CorsMiddleware()*/)
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
