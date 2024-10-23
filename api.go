package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	router.HandleFunc("POST /gyms/{gymId}/ratings", makeHTTPHandleFunc(s.handleRateGym))

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
	reqId := req.PathValue("id")
	log.Println("Received method to GET a gym with id:", reqId)

	id, err := strconv.Atoi(reqId)
	if err != nil {
		return err
	}

	gym, err := s.store.GetGymByID(id)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, gym)
}

func (s *APIServer) handleCreateGym(w http.ResponseWriter, req *http.Request) error {
	createGymRequest := new(CreateGymRequest)

	// Decode the json using our request struct
	if err := json.NewDecoder(req.Body).Decode(createGymRequest); err != nil {
		return err
	}

	gym := NewGym(createGymRequest.Name, createGymRequest.Description) // Interface for passed gym parameters

	createdGym, err := s.store.CreateGym(gym)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, createdGym)
}

func (s *APIServer) handleRateGym(w http.ResponseWriter, req *http.Request) error {

	reqId := req.PathValue("gymId")
	createRatingRequest := new(CreateRatingRequest)

	gymId, err := strconv.Atoi(reqId)

	if err != nil {
		return err
	}

	if err := json.NewDecoder(req.Body).Decode(createRatingRequest); err != nil {
		return err
	}

	log.Println("Received method to RATE gym with id:", gymId)

	gym, err := s.store.GetGymByID(gymId)

	if err != nil {
		return err
	}

	if gym == nil {
		return WriteJSON(w, http.StatusNotFound, fmt.Sprintf("Gym with ID: %d not found.", gymId))
	}

	rating := NewRating(
		gymId,
		createRatingRequest.Rating,
		createRatingRequest.UserName,
		createRatingRequest.Review,
	)

	createdRating, err := s.store.CreateRating(rating)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, createdRating)
}

func (s *APIServer) handleDeleteGym(w http.ResponseWriter, req *http.Request) error {
	reqId := req.PathValue("id")
	log.Println("Received method to DELETE gym with id:", reqId)
	id, err := strconv.Atoi(reqId)

	if err != nil {
		return err
	}

	if err = s.store.DeleteGym(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, "Gym successfully deleted")
}
