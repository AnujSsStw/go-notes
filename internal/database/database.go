package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	in "go-notes/internal"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Health() map[string]string
	RetriveUser(*in.User) (*in.User, error)
	CreateUser(*in.User) error
}

type service struct {
	db *sql.DB
}

var (
	database = os.Getenv("DB_DATABASE")
	password = os.Getenv("DB_PASSWORD")
	username = os.Getenv("DB_USERNAME")
	port     = os.Getenv("DB_PORT")
	host     = os.Getenv("DB_HOST")
)

func New() Service {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	fmt.Println(connStr)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	s := &service{db: db}
	s.CreateTable()
	return s
}

func (s *service) CreateUser(u *in.User) error {
	query := `insert into users 
	(username, password, api_key)
	values ($1, $2, $3)`

	_, err := s.db.Query(
		query,
		u.Username, u.Password, u.ApiKey)

	if err != nil {
		return err
	}

	return nil
}
func (s *service) RetriveUser(u *in.User) (*in.User, error) {
	query := `SELECT *
	FROM users
	WHERE username = $1;`
	rows, err := s.db.Query(
		query,
		u.Username)

	if err != nil {
		return nil, err
	}

	user := new(in.User)
	for rows.Next() {
		if err := rows.Scan(&user.Id, &user.Username, &user.Password, &user.CreatedAt, &user.ApiKey); err != nil {
			return nil, err
		}
	}

	if user.Username == "" {
		return nil, errors.New("usr not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)) == nil {
		return user, nil
	} else {
		return nil, errors.New("password not matches")
	}

}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.PingContext(ctx)
	if err != nil {
		log.Fatalf(fmt.Sprintf("db down: %v", err))
	}

	return map[string]string{
		"message": "It's healthy",
	}
}

func (s *service) CreateTable() {
	userTable := `CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	api_key VARCHAR(255) UNIQUE NOT NULL
	);`

	noteTable := `CREATE TABLE IF NOT EXISTS notes (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
	note_title TEXT,
    note_text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := s.db.Exec(userTable); err != nil {
		log.Fatalln("in user table", err)
	}
	if _, err := s.db.Exec(noteTable); err != nil {
		log.Fatalln("in notes table", err)
	}
}
