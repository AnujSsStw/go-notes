package server

import (
	in "go-notes/internal"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func (s *FiberServer) RegisterFiberRoutes(api fiber.Router) {
	s.App.Get("/", s.HelloWorldHandler)
	s.App.Get("/health", s.healthHandler)

	// authMiddleware := keyauth.New(keyauth.Config{
	// 	Validator: func(c *fiber.Ctx, key string) (bool, error) {
	// 		hashedAPIKey := sha256.Sum256([]byte(apiKey))
	// 		hashedKey := sha256.Sum256([]byte(key))

	// 		if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1 {
	// 			return true, nil
	// 		}
	// 		return false, keyauth.ErrMissingOrMalformedAPIKey
	// 	},
	// })

	// POST /api/auth/CreateUser: create a new user account.
	// POST /api/auth/login: log in to an existing user account and receive an access token.
	api.Post("/auth/signup", s.CreateUser)
	api.Post("/auth/login", s.login)

	// GET /api/notes: get a list of all notes for the authenticated user.
	// GET /api/notes/:id: get a note by ID for the authenticated user.
	// POST /api/notes: create a new note for the authenticated user.
	// PUT /api/notes/:id: update an existing note by ID for the authenticated user.
	// DELETE /api/notes/:id: delete a note by ID for the authenticated user.
	// POST /api/notes/:id/share: share a note with another user for the authenticated user.
	// GET /api/search?q=:query: search for notes based on keywords for the authenticated user.
	/*
		api.Get("/notes")
		api.Get("/notes/:id")
		api.Post("/notes")
		api.Put("/notes/:id")
		api.Delete("/notes/:id")
		api.Post("/notes/:id/share")
		api.Get("/search")
	*/

}

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

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := map[string]string{
		"message": "Hello World",
	}
	return c.JSON(resp)
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}
