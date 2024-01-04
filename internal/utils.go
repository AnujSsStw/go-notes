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

func GenerateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	return randomString[:length], nil
}
