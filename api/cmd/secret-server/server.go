package main

import (
	"errors"
	"net/http"

	"github.com/evassilyev/secret-server/api/monitoring"

	"github.com/evassilyev/secret-server/api/core"
	"github.com/evassilyev/secret-server/api/dev"
	"github.com/evassilyev/secret-server/api/pgdb"
	"github.com/evassilyev/secret-server/api/redis"
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
	mon      *monitoring.PrometheusEndpoint
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

	var err error
	switch v.GetString("storage") {
	case "dev":
		s.services.Secret = dev.NewSecretService()
	case "pgdb":
		if v.Get("pgdb") != nil {
			s.services.Secret, err = pgdb.NewSecretService(v.GetString("pgdb.url"), v.GetInt("pgdb.maxIdleConns"), v.GetInt("pgdb.maxOpenConnsKey"))
		} else {
			return errors.New("no postgres configuration found")
		}
	case "redis":
		if v.Get("redis") != nil {
			s.services.Secret = redis.NewSecretService(v.GetString("redis.addr"), v.GetString("redis.password"), v.GetInt("redis.db"))
		} else {
			return errors.New("no redis configuration found")
		}
	default:
		return errors.New("wrong storage configuration")
	}
	return err
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

func (s *server) runMonitoring(v *viper.Viper) error {
	if v.Get("prometheus") == nil {
		return errors.New("no monitoring configuration found")
	}
	s.mon = monitoring.NewPrometheusEndpoint(
		v.GetString("prometheus.endpoint"),
		v.GetString("prometheus.addr"))

	go s.mon.RunPrometheusEndpoint()
	return nil
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
	err = s.runMonitoring(v)
	if err != nil {
		return err
	}
	s.initHandler()
	return http.ListenAndServe(s.config.Addr, s.handler)
}
