package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *server) secretSaveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Save handler!")

	/*
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	*/
	return
}

func (s *server) secretGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	fmt.Println("I'm here", hash)

	/*
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	*/
	return
}
