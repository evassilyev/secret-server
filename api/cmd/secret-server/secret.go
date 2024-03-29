package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func (s *server) secretSaveHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	var parameters struct {
		secret           string
		expireAfterViews int
		expireAfter      int
	}

	beginTime := time.Now()
	defer func() {
		r.Body.Close()
		go func() { s.mon.PostRequestRT <- time.Since(beginTime) }()
	}()

	go func() {
		s.mon.PostRequest <- true
	}()

	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		http.Error(w, "Invalid input", http.StatusMethodNotAllowed)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid input", http.StatusMethodNotAllowed)
		return
	}

	parameters.secret = r.Form.Get("secret")

	parameters.expireAfterViews, err = strconv.Atoi(r.Form.Get("expireAfterViews"))
	if err != nil || parameters.expireAfterViews <= 0 {
		http.Error(w, "Invalid input", http.StatusMethodNotAllowed)
		return
	}

	parameters.expireAfter, err = strconv.Atoi(r.Form.Get("expireAfter"))
	if err != nil {
		http.Error(w, "Invalid input", http.StatusMethodNotAllowed)
		return
	}

	secret, err := s.services.Secret.Save(parameters.secret, parameters.expireAfterViews, parameters.expireAfter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataResponse(w, r, secret)
}

func (s *server) secretGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	beginTime := time.Now()
	defer func() {
		go func() { s.mon.GetRequestRT <- time.Since(beginTime) }()
	}()

	go func() {
		s.mon.GetRequest <- true
	}()

	secret, err := s.services.Secret.Get(hash)
	if err != nil {
		http.Error(w, "Secret not found", http.StatusNotFound)
		return
	}

	dataResponse(w, r, secret)
}

func dataResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	// Extendable part
	switch r.Header.Get("Accept") {
	case "application/json":
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/xml":
		err := xml.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Invalid input", http.StatusMethodNotAllowed)
		return
	}
}
