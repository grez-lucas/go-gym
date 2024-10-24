package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// This module is responsible for DB connections, and being DB agnostic!

type Storage interface {
	CreateGym(*Gym) (*Gym, error)
	DeleteGym(int) error
	UpdateGym(*Gym) error
	GetGymByID(int) (*Gym, error)
	GetGyms() ([]*Gym, error)
	CreateRating(*Rating) (*Rating, error)
	GetAverageRating(int) (float32, error)
}

type PostgreSQLStore struct {
	db *sql.DB
}

func NewPostgreSQLStore() (*PostgreSQLStore, error) {
	connStr, found := os.LookupEnv("DATABASE_URL")

	// For running locally
	if !found {
		connStr = "host=localhost user=postgres dbname=postgres password=gogym sslmode=disable"
	}

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

func (s *PostgreSQLStore) CreateGym(gym *Gym) (*Gym, error) {
	// To avoid SQL injection, avoid using your custom Sprintf format!
	// Instead use something like this
	createdGym := new(Gym)

	query := `
    INSERT INTO gyms (name, description, created_at, updated_at)
    values ($1, $2, $3, $4)
    RETURNING id, name, description, created_at, updated_at`

	err := s.db.QueryRow(query, gym.Name, gym.Description, gym.CreatedAt, gym.UpdatedAt).Scan(
		&createdGym.ID,
		&createdGym.Name,
		&createdGym.Description,
		&createdGym.CreatedAt,
		&createdGym.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	log.Printf("Created new record with ID: %v", &createdGym.ID)

	return createdGym, nil
}

func (s *PostgreSQLStore) DeleteGym(id int) error {

	query := `
    DELETE FROM gyms
    WHERE id=$1
  `

	_, err := s.db.Exec(query, id)

	if err != nil {
		return err
	}

	log.Printf("Gym with id %d successfully deleted\n", id)

	return nil
}

func (s *PostgreSQLStore) UpdateGym(*Gym) error {
	return nil
}

func (s *PostgreSQLStore) GetGymByID(id int) (*Gym, error) {

	gym := new(Gym)

	query := `
    SELECT * from gyms
    WHERE  id=$1
  `

	row := s.db.QueryRow(query, id)

	err := row.Scan(
		&gym.ID,
		&gym.Name,
		&gym.Description,
		&gym.CreatedAt,
		&gym.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error fetching gym with ID %d: %s\n", id, err.Error())
		return nil, err
	}

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

		log.Println("Getting average rating for gym with ID:", gym.ID)
		avgRating, err := s.GetAverageRating(gym.ID)

		if err != nil {
			return nil, err
		}

		gym.Rating = avgRating

		gyms = append(gyms, gym)
	}

	return gyms, nil
}

func (s *PostgreSQLStore) CreateRating(r *Rating) (*Rating, error) {
	query := `
    INSERT INTO ratings (gym_id, rating, user_name, review, created_at, updated_at)
    values ($1, $2, $3, $4, $5, $6)
    RETURNING id, gym_id, rating, user_name, review, created_at, updated_at
  `

	row := s.db.QueryRow(query, r.GymID, r.Rating, r.UserName, r.Review, r.CreatedAt, r.UpdatedAt)

	return scanIntoRating(row)
}

func (s *PostgreSQLStore) GetAverageRating(id int) (float32, error) {

	query := `
    SELECT COALESCE( AVG(rating), 0 ) AS average_rating
    FROM ratings
    WHERE gym_id=$1
  `

	var avgRating float32
	if err := s.db.QueryRow(query, id).Scan(&avgRating); err != nil {
		log.Printf("Error calculating average rating: %s", err.Error())
		return avgRating, err
	}

	return avgRating, nil
}

func scanIntoRating(row *sql.Row) (*Rating, error) {
	createdRating := new(Rating)

	err := row.Scan(
		&createdRating.ID,
		&createdRating.GymID,
		&createdRating.Rating,
		&createdRating.UserName,
		&createdRating.Review,
		&createdRating.CreatedAt,
		&createdRating.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return createdRating, nil

}
