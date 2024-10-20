package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)               // Write the status
	return json.NewEncoder(w).Encode(v) // To encode anything
}

type APIServer struct {
	listenAddr string
}

type APIFunc func(http.ResponseWriter, *http.Request) error

type APIError struct {
	Error string
}

// To decorate our APIFunc into an HTTP handler
// This way we can handle errors in our func logic, not the handler
func makeHTTPHandleFunc(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := f(w, req)
		if err != nil {
			// Handle the error
			WriteJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

// Function to start our server up
func (s *APIServer) Run() {

	// router := mux.NewRouter()
	//
	// router.handler

	http.HandleFunc("/gyms", makeHTTPHandleFunc(s.handleGyms))

	log.Println("JSON API Running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, nil)

}

// Handlers: a handler handles a specific route
// name convention is handleFooBar
func (s *APIServer) handleGyms(w http.ResponseWriter, req *http.Request) error {
	// This is the entry function which differentiates between POST, GET and DELETE requests
	if req.Method == "GET" {
		return s.handleGetGym(w, req)
	}

	if req.Method == "POST" {
		return s.handleCreateGym(w, req)
	}
	return fmt.Errorf("Method not allowed `%s` ", req.Method)
}

func (s *APIServer) handleGetGyms(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func (s *APIServer) handleGetGym(w http.ResponseWriter, req *http.Request) error {
	gym := NewGym("Sportlife", "A gym for adding sport to your life")
	return WriteJSON(w, http.StatusOK, gym)
}

func (s *APIServer) handleCreateGym(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func (s *APIServer) handleRateGym(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func (s *APIServer) handleDeleteGym(w http.ResponseWriter, req *http.Request) error {
	return nil
}
