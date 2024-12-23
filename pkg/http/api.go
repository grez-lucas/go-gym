package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/grez-lucas/go-gym/pkg/domain"
	"github.com/grez-lucas/go-gym/pkg/storage"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)               // Write the status
	return json.NewEncoder(w).Encode(v) // To encode anything
}

type APIServer struct {
	listenAddr string
	// This way we can abstract the DB to anything that implements the Storage interface
	store storage.Storage
}

type APIFunc func(http.ResponseWriter, *http.Request) error

type APIError struct {
	Error string `json:"error"`
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

func NewAPIServer(listenAddr string, store storage.Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

// Function to start our server up
func (s *APIServer) Run() {

	router := http.NewServeMux()

	router.HandleFunc("GET /healthcheck", makeHTTPHandleFunc(s.handleGetHealthcheck))
	router.HandleFunc("GET /login", makeHTTPHandleFunc(s.handleGetLogin))
	router.HandleFunc("GET /gyms", makeHTTPHandleFunc(s.handleGetGyms))
	router.HandleFunc("GET /gyms/{id}", makeHTTPHandleFunc(s.handleGetGym))
	router.HandleFunc("POST /gyms", makeHTTPHandleFunc(s.handleCreateGym))
	router.HandleFunc("DELETE /gyms/{id}", makeHTTPHandleFunc(s.handleDeleteGym))
	router.HandleFunc("POST /gyms/{id}/ratings", WithJWTAuth(makeHTTPHandleFunc(s.handleRateGym)))
	router.HandleFunc("GET /accounts", WithJWTAuth(makeHTTPHandleFunc(s.handleGetAccounts)))
	router.HandleFunc("POST /accounts", makeHTTPHandleFunc(s.handleCreateAccount))

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

func (s *APIServer) handleGetLogin(w http.ResponseWriter, req *http.Request) error {
	var loginRequest LoginRequest

	if err := json.NewDecoder(req.Body).Decode(&loginRequest); err != nil {
		return err
	}

	acc, err := s.store.GetAccountByUsername(loginRequest.Username)

	if err != nil {
		return err
	}

	// Validate password

	if !storage.VerifyHashedPassword(loginRequest.Password, acc.Password) {
		return fmt.Errorf("Invalid Password")
	}

	token, err := CreateJWT(acc)

	if err != nil {
		return err
	}

	resp := LoginResponse{Token: token, AccountID: acc.ID}

	return WriteJSON(w, http.StatusOK, resp)
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
	id, err := GetID(req)
	if err != nil {
		return err
	}
	log.Println("Received method to GET a gym with id:", id)

	gym, err := s.store.GetGymByID(id)

	if err != nil {
		return err
	}

	// Get the average rating for said gym
	avgRating, err := s.store.GetAverageRating(id)

	if err != nil {
		return err
	}

	gym.Rating = avgRating

	return WriteJSON(w, http.StatusOK, gym)
}

func (s *APIServer) handleCreateGym(w http.ResponseWriter, req *http.Request) error {
	createGymRequest := new(domain.CreateGymRequest)

	// Decode the json using our request struct
	if err := json.NewDecoder(req.Body).Decode(createGymRequest); err != nil {
		return err
	}

	gym := domain.NewGym(createGymRequest.Name, createGymRequest.Description) // Interface for passed gym parameters

	createdGym, err := s.store.CreateGym(gym)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, createdGym)
}

func (s *APIServer) handleRateGym(w http.ResponseWriter, req *http.Request) error {
	accountID, ok := AccountIDFromContext(req.Context())

	if !ok {
		return WriteJSON(w, http.StatusUnauthorized, APIError{Error: "Unable to retrieve ID from context"})
	}

	acc, err := s.store.GetAccountByID(int(accountID))

	if err != nil {
		return err
	}

	gymId, err := GetID(req)
	if err != nil {
		return err
	}
	log.Println("Received method to RATE gym with id:", gymId)

	createRatingRequest := new(domain.CreateRatingRequest)
	if err := json.NewDecoder(req.Body).Decode(createRatingRequest); err != nil {
		return err
	}

	gym, err := s.store.GetGymByID(gymId)

	if err != nil {
		return err
	}

	if gym == nil {
		return WriteJSON(w, http.StatusNotFound, fmt.Sprintf("Gym with ID: %d not found.", gymId))
	}

	rating := domain.NewRating(
		gymId,
		createRatingRequest.Rating,
		acc.UserName,
		createRatingRequest.Review,
	)

	createdRating, err := s.store.CreateRating(rating)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, createdRating)
}

func (s *APIServer) handleDeleteGym(w http.ResponseWriter, req *http.Request) error {
	id, err := GetID(req)
	if err != nil {
		return err
	}
	log.Println("Received method to DELETE gym with id:", id)

	if err = s.store.DeleteGym(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int{"Gym successfully deleted": id})
}

func GetID(req *http.Request) (int, error) {
	reqId := req.PathValue("id")

	id, err := strconv.Atoi(reqId)
	if err != nil {
		return id, fmt.Errorf("Invalid id given %s", reqId)
	}

	return id, nil
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, req *http.Request) error {
	createAccountRequest := new(domain.CreateAccountRequest)
	if err := json.NewDecoder(req.Body).Decode(createAccountRequest); err != nil {
		return err
	}

	account := domain.NewAccount(createAccountRequest.UserName, createAccountRequest.Password)

	createdAccount, err := s.store.CreateAccount(account)

	if err != nil {
		return err
	}

	// Create a JWT for said account

	tokenStr, err := CreateJWT(createdAccount)

	if err != nil {
		return err
	}

	log.Printf("Created JWT Token `%s` for account `%d`", tokenStr, createdAccount.ID)

	return WriteJSON(w, http.StatusCreated, createdAccount)
}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, req *http.Request) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}
