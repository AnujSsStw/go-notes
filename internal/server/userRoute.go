package server

import (
	in "go-notes/internal"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func (s *FiberServer) CreateUser(c *fiber.Ctx) error {
	u := new(in.User)

	if err := c.BodyParser(u); err != nil {
		return err
	}
	encpw, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(encpw)
	apikey, err := in.GenerateRandomString(20)
	if err != nil {
		return err
	}
	u.ApiKey = apikey

	if err := s.db.CreateUser(u); err != nil {
		return c.JSON(map[string]string{
			"message": "username already exist",
		})
	}
	return c.JSON(u.Username)
}

func (s *FiberServer) login(c *fiber.Ctx) error {
	u := new(in.User)

	if err := c.BodyParser(u); err != nil {
		return err
	}
	if user, err := s.db.RetriveUser(u); err != nil {
		return c.JSON(map[string]string{
			"message": "wrong username and password combo",
			"error":   err.Error(),
		})
	} else {
		return c.JSON(user.ApiKey)
	}
}
