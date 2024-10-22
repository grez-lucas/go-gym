package main

import (
	"encoding/json"
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
	// This way we can abstract the DB to anything that implements the Storage interface
	store Storage
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

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

// Function to start our server up
func (s *APIServer) Run() {

	router := http.NewServeMux()

	router.HandleFunc("GET /healthcheck", makeHTTPHandleFunc(s.handleGetHealthcheck))
	router.HandleFunc("GET /gyms", makeHTTPHandleFunc(s.handleGetGyms))
	router.HandleFunc("GET /gyms/{id}", makeHTTPHandleFunc(s.handleGetGym))
	router.HandleFunc("POST /gyms", makeHTTPHandleFunc(s.handleCreateGym))
	router.HandleFunc("DELETE /gyms/{id}", makeHTTPHandleFunc(s.handleDeleteGym))
	router.HandleFunc("POST /gyms/{id}/ratings", makeHTTPHandleFunc(s.handleRateGym))

	server := http.Server{
		Addr:    s.listenAddr,
		Handler: router,
	}
	log.Println("Starting JSON API on port: ", s.listenAddr)
	server.ListenAndServe()

}

// Handlers: a handler handles a specific route
// name convention is handleFooBar
func (s *APIServer) handleGetHealthcheck(w http.ResponseWriter, req *http.Request) error {
	log.Println("Received Healthcheck request")

	return WriteJSON(w, http.StatusOK, "Healthcheck - OK")
}

func (s *APIServer) handleGetGyms(w http.ResponseWriter, req *http.Request) error {
	log.Println("Received method to GET all gyms")

	gyms, err := s.store.GetGyms()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, gyms)
}

func (s *APIServer) handleGetGym(w http.ResponseWriter, req *http.Request) error {
	id := req.PathValue("id")
	log.Println("Received method to GET a gym with id:", id)
	gym := NewGym("Sportlife", "A gym for adding sport to your life")
	return WriteJSON(w, http.StatusOK, gym)
}

func (s *APIServer) handleCreateGym(w http.ResponseWriter, req *http.Request) error {
	createGymRequest := new(CreateGymRequest)

	// Decode the json using our request struct
	if err := json.NewDecoder(req.Body).Decode(createGymRequest); err != nil {
		return err
	}

	gym := NewGym(createGymRequest.Name, createGymRequest.Description)

	// Store the gym instance into our Storage
	if err := s.store.CreateGym(gym); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, gym)
}

func (s *APIServer) handleRateGym(w http.ResponseWriter, req *http.Request) error {
	id := req.PathValue("id")
	log.Println("Received method to RATE gym with id:", id)
	return nil
}

func (s *APIServer) handleDeleteGym(w http.ResponseWriter, req *http.Request) error {
	id := req.PathValue("id")
	log.Println("Received method to DELETE gym with id:", id)
	return nil
}
