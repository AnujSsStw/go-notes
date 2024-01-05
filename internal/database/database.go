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
	ValidateApiKey(string) (*in.User, error)
	CreateNote(*in.Note) error
	GetNotes(string) ([]*in.Note, error)
	GetNote(string, string) (*in.Note, error)
	SetPrivacy(string, string) error
	DeleteNote(string, string) error
	GetNoteById(string) (*in.Note, error)
	UpdateNote(string, string, string, *in.Note) error
	SearchNote(string, string) ([]*in.Note, error)
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
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	s := &service{db: db}
	s.CreateTable()
	return s
}

func (s *service) SearchNote(userId, q string) ([]*in.Note, error) {
	query := `
	SELECT notes.id AS note_id, notes.title, notes.text, notes.created_at, notes.updated_at
	FROM notes
	JOIN users ON notes.user_id = users.id
	WHERE notes.user_id = $1 AND ts @@ to_tsquery('simple', $2)
	ORDER BY ts_rank(ts, to_tsquery('simple', $2)) DESC;
	`

	rows, err := s.db.Query(query, userId, q)
	if err != nil {
		return nil, err
	}

	notes := []*in.Note{}
	for rows.Next() {
		note := new(in.Note)
		if err := rows.Scan(&note.Id, &note.Title, &note.Text, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (s *service) UpdateNote(userId, noteId, flag string, note *in.Note) error {
	var query string
	// does note with noteid exist for that user
	if _, err := s.GetNote(userId, noteId); err != nil {
		return err
	}

	switch flag {
	case "BOTH":
		query = "UPDATE notes SET text = $1, title = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND user_id = $4;"
		_, err := s.db.Query(query, note.Text, note.Title, noteId, userId)
		return err
	case "TEXT":
		query = "UPDATE notes SET text = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3;"
		_, err := s.db.Query(query, note.Text, noteId, userId)
		return err
	case "TITLE":
		query = "UPDATE notes SET title = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3;"
		_, err := s.db.Query(query, note.Title, noteId, userId)
		return err
	default:
		return errors.New("FLag not set properly means not updating anything")
	}
}

func (s *service) DeleteNote(noteId, userId string) error {
	query := `DELETE FROM notes WHERE user_id = $1 AND id = $2;`
	_, err := s.db.Query(query, userId, noteId)
	return err
}

func (s *service) SetPrivacy(noteId, userId string) error {
	query := `UPDATE notes
	SET is_private = NOT is_private
	WHERE user_id = $1 AND id = $2;`

	_, err := s.db.Query(query, userId, noteId)
	return err
}

func (s *service) GetNoteById(noteId string) (*in.Note, error) {
	query := `
		SELECT notes.id AS note_id, notes.title, notes.text, notes.created_at, notes.updated_at
		FROM notes
		JOIN users ON notes.user_id = users.id
		WHERE notes.is_private = FALSE AND notes.id = $1;`
	note := new(in.Note)
	err := s.db.QueryRow(query, noteId).Scan(&note.Id, &note.Title, &note.Text, &note.CreatedAt, &note.UpdatedAt)

	switch err {
	case sql.ErrNoRows:
		return nil, errors.New("no note found")
	case nil:
		return note, nil
	default:
		return nil, err
	}
}

func (s *service) GetNote(userId, noteId string) (*in.Note, error) {
	query := `
	SELECT notes.id AS note_id, notes.title, notes.text, notes.created_at
	FROM notes
	JOIN users ON notes.user_id = users.id
	WHERE notes.user_id = $1 AND notes.id = $2;`

	note := new(in.Note)
	err := s.db.QueryRow(query, userId, noteId).Scan(&note.Id, &note.Title, &note.Text, &note.CreatedAt)
	switch err {
	case sql.ErrNoRows:

		return nil, errors.New("no note found")
	case nil:
		return note, nil
	default:
		return nil, err
	}
}

func (s *service) GetNotes(userId string) ([]*in.Note, error) {
	query := `
	SELECT notes.id AS note_id, notes.title, notes.text, notes.created_at
	FROM notes
	JOIN users ON notes.user_id = users.id
	WHERE notes.user_id = $1;`

	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}

	notes := []*in.Note{}
	for rows.Next() {
		note := new(in.Note)
		if err := rows.Scan(&note.Id, &note.Title, &note.Text, &note.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (s *service) CreateNote(n *in.Note) error {
	query := "INSERT INTO notes (user_id, title, text) VALUES ($1, $2, $3)"
	_, err := s.db.Query(query, n.UserId, n.Title, n.Text)
	return err
}

func (s *service) ValidateApiKey(u string) (*in.User, error) {
	query := `SELECT *
	FROM users
	WHERE api_key = $1;`
	rows, err := s.db.Query(
		query,
		u)

	if err != nil {
		return nil, err
	}

	user := new(in.User)
	for rows.Next() {
		if err := rows.Scan(&user.Id, &user.Username, &user.Password, &user.CreatedAt, &user.ApiKey); err != nil {
			return nil, err
		}
	}

	return user, nil
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
	title TEXT,
    text TEXT,
	is_private BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	search := `DO $$ 
	BEGIN 
	  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'notes' AND column_name = 'ts') THEN 
	    ALTER TABLE notes ADD COLUMN ts tsvector 
		GENERATED ALWAYS AS 
     	(setweight(to_tsvector('english', coalesce(title, '')), 'A') || 
		setweight(to_tsvector('english', coalesce(text, '')), 'B')) STORED;
	  END IF; 
	END $$;
	`
	idx := `CREATE INDEX IF NOT EXISTS ts_idx ON notes USING GIN (ts);`

	if _, err := s.db.Exec(userTable); err != nil {
		log.Fatalln("in user table", err)
	}
	if _, err := s.db.Exec(noteTable); err != nil {
		log.Fatalln("in notes table", err)
	}
	if _, err := s.db.Exec(search); err != nil {
		log.Fatalln("in user table", err)
	}
	if _, err := s.db.Exec(idx); err != nil {
		log.Fatalln("in user table", err)
	}
}
