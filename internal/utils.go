package internal

import (
	"crypto/rand"
	"encoding/base64"
)

type User struct {
	Id        string `json:"id" form:"id"`
	Username  string `json:"username" form:"username"`
	Password  string `json:"password" form:"password"`
	CreatedAt string `json:"created_at" form:"created_at"`
	ApiKey    string `json:"api_key" form:"api_key"`
}

type Note struct {
	Id        string `json:"id" form:"id"`
	UserId    string `db:"user_id"`
	Title     string `json:"title" form:"title"`
	Text      string `json:"text" form:"text"`
	CreatedAt string `json:"created_at" form:"created_at"`
	UpdatedAt string `json:"updated_at" form:"updated_at"`
}

func GenerateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	return randomString[:length], nil
}
