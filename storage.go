package main

import (
	"database/sql"
	"fmt"
	"log"

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
	CreateAccount(*Account) (*Account, error)
}

type PostgreSQLStore struct {
	db *sql.DB
}

func NewPostgreSQLStore() (*PostgreSQLStore, error) {

	config := LoadConfig()

	connStr := config.PostgreSQLConnStr()
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

	if err := s.CreateAccountsTable(); err != nil {
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

func (s *PostgreSQLStore) CreateAccountsTable() error {
	query := `
    CREATE table if not exists accounts (
      id SERIAL PRIMARY KEY,
      username VARCHAR(100) UNIQUE NOT NULL,
      password VARCHAR(255) NOT NULL, 
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
	query := `
    INSERT INTO gyms (name, description, created_at, updated_at)
    values ($1, $2, $3, $4)
    RETURNING id, name, description, created_at, updated_at`

	rows, err := s.db.Query(query, gym.Name, gym.Description, gym.CreatedAt, gym.UpdatedAt)

	if err != nil {
		log.Println("Error creating Gym: ", err.Error())
		return nil, err
	}

	for rows.Next() {
		return scanIntoGym(rows)
	}

	return nil, fmt.Errorf("Error creating Gym")
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

	query := `
    SELECT * from gyms
    WHERE  id=$1
  `

	rows, err := s.db.Query(query, id)

	if err != nil {
		log.Printf("Error getting gym with ID: %d - %s\n", id, err.Error())
		return nil, err
	}

	for rows.Next() {
		return scanIntoGym(rows)
	}

	return nil, fmt.Errorf("Gym with ID %d not found", id)
}

func (s *PostgreSQLStore) GetGyms() ([]*Gym, error) {

	gyms := []*Gym{}

	query := `SELECT * FROM gyms`

	rows, err := s.db.Query(query)

	if err != nil {
		log.Printf("Error fetching gyms: %s\n", err.Error())
		return nil, err
	}

	// For each row, save gym to memory and check for errors
	for rows.Next() {
		gym, err := scanIntoGym(rows)

		if err != nil {
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

func (s *PostgreSQLStore) CreateAccount(a *Account) (*Account, error) {

	query := `
    INSERT INTO accounts (username, password, created_at, updated_at)
    VALUES ($1, $2, $3, $4)
    RETURNING id, username, password, created_at, updated_at
  `

	rows, err := s.db.Query(query, a.UserName, a.Password, a.CreatedAt, a.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("DB error when creating account: `%s`", err.Error())
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("Error creating account")

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

func scanIntoGym(row *sql.Rows) (*Gym, error) {
	gym := new(Gym)

	err := row.Scan(
		&gym.ID,
		&gym.Name,
		&gym.Description,
		&gym.CreatedAt,
		&gym.UpdatedAt,
	)

	if err != nil {
		log.Printf("SQL Error when scanning Gym: %s", err.Error())
		return nil, err
	}

	return gym, nil
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
		log.Printf("SQL Error when scanning Rating: %s", err.Error())
		return nil, err
	}

	return createdRating, nil

}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	createdAccount := new(Account)

	err := rows.Scan(
		&createdAccount.ID,
		&createdAccount.UserName,
		&createdAccount.Password,
		&createdAccount.CreatedAt,
		&createdAccount.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return createdAccount, nil

}
