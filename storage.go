package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// This module is responsible for DB connections, and being DB agnostic!

type Storage interface {
	CreateGym(*Gym) error
	DeleteGym(int) error
	UpdateGym(*Gym) error
	GetGymByID(int) (*Gym, error)
	GetGyms() ([]*Gym, error)
}

type PostgreSQLStore struct {
	db *sql.DB
}

func NewPostgreSQLStore() (*PostgreSQLStore, error) {
	connStr := "user=postgres dbname=postgres password=mysecretpassword sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	// Ping the DB to healthcheck it
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgreSQLStore{
		db: db,
	}, nil
}

func (s *PostgreSQLStore) Init() error {

	err := s.CreateGymsTable()

	if err != nil {
		return err
	}

	if err := s.CreateRatingsTable(); err != nil {
		return err
	}

	return nil
}

func (s *PostgreSQLStore) CreateGymsTable() error {
	query := `create table if not exists gyms (
      id SERIAL PRIMARY KEY,
      name VARCHAR(100) NOT NULL,
      description TEXT,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  )`

	_, err := s.db.Query(query)

	if err != nil {
		return err
	}

	return nil

}
func (s *PostgreSQLStore) CreateRatingsTable() error {
	query := `create table if not exists ratings  (
      id SERIAL PRIMARY KEY,
      gym_id INT REFERENCES gyms(id) ON DELETE CASCADE,
      rating INT CHECK (rating >= 1 AND rating <= 5) NOT NULL,
      user_name VARCHAR(100) NOT NULL,
      review TEXT,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  )`

	_, err := s.db.Query(query)

	if err != nil {
		return err
	}

	return nil

}

func (s *PostgreSQLStore) CreateGym(gym *Gym) error {
	// To avoid SQL injection, avoid using your custom Sprintf format!
	// Instead use something like this
	query := `
    INSERT INTO gyms (name, description, created_at, updated_at)
    values ($1, $2, $3, $4)
    RETURNING id`

	id := 0
	err := s.db.QueryRow(query, gym.Name, gym.Description, gym.CreatedAt, gym.UpdatedAt).Scan(&id)

	if err != nil {
		return err
	}

	log.Printf("Created new record with ID: %v", id)

	return nil
}

func (s *PostgreSQLStore) DeleteGym(id int) error {
	return nil
}

func (s *PostgreSQLStore) UpdateGym(*Gym) error {
	return nil
}

func (s *PostgreSQLStore) GetGymByID(id int) (*Gym, error) {
	gym := NewGym("Example Gym", "This is an example gym")
	return gym, nil
}

func (s *PostgreSQLStore) GetGyms() ([]*Gym, error) {

	gyms := []*Gym{}

	query := `SELECT * FROM gyms`

	rows, err := s.db.Query(query)

	if err != nil {
		log.Printf("Error fetching gyms: %s\n", err.Error())
		return gyms, err
	}

	// For each row, save gym to memory and check for errors
	for rows.Next() {
		gym := new(Gym)
		err := rows.Scan( // Copy the values of row into our destination Gym

			&gym.ID,
			&gym.Name,
			&gym.Description,
			&gym.CreatedAt,
			&gym.UpdatedAt,
		)

		if err != nil {
			log.Printf("Error fetching gyms: %s\n", err.Error())
			return nil, err
		}

		gyms = append(gyms, gym)
	}

	return gyms, nil
}
