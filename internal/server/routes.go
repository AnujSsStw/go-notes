package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

var protectedURLs = []*regexp.Regexp{
	regexp.MustCompile("^/api/notes(/.*)?"),
	regexp.MustCompile("^/api/search(/.*)?"),
}

func authFilter(c *fiber.Ctx) bool {
	originalURL := strings.ToLower(c.OriginalURL())

	for _, pattern := range protectedURLs {
		if pattern.MatchString(originalURL) {
			return false
		}
	}
	return true
}

func (s *FiberServer) RegisterFiberRoutes() {
	s.App.Get("/", s.HelloWorldHandler)
	s.App.Get("/health", s.healthHandler)
	authMiddleware := keyauth.New(keyauth.Config{
		Next: authFilter,
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			user, err := s.db.ValidateApiKey(key)
			c.Locals("user", user.Id)
			if err != nil {
				log.Println(err)
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}
			hashedAPIKey := sha256.Sum256([]byte(user.ApiKey))
			hashedKey := sha256.Sum256([]byte(key))

			if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1 {
				return true, nil
			}
			return false, keyauth.ErrMissingOrMalformedAPIKey
		},
	})

	api := s.App.Group("/api", authMiddleware)

	// POST /api/auth/CreateUser: create a new user account.
	/*POST /api/auth/login: log in to an existing user account and receive an access token.
	produces token can be used as	--header "Authorization: Bearer my-super-secret-key" */
	api.Post("/auth/signup", s.CreateUser)
	api.Post("/auth/login", s.login)

	// GET /api/notes: get a list of all notes for the authenticated user.
	// GET /api/notes/:id: get a note by ID for the authenticated user.
	// POST /api/notes: create a new note for the authenticated user.
	// PUT /api/notes/:id: update an existing note by ID for the authenticated user.
	// DELETE /api/notes/:id: delete a note by ID for the authenticated user.
	// POST /api/notes/:id/share: share a note with another user for the authenticated user.
	// GET /api/search?q=:query: search for notes based on keywords for the authenticated user.
	// curl --header "Authorization: Bearer my-super-secret-key"  http://localhost:3000/api/notes...

	api.Get("/notes", s.getNotes)
	api.Get("/notes/:id", s.getNote)
	api.Post("/notes", s.createNote)
	api.Put("/notes/:id", s.updateNote)
	api.Delete("/notes/:id", s.deleteNote)
	api.Post("/notes/:id/share", s.shareNote)
	api.Get("/search", s.searchNote)

	// to set the note public or private
	api.Put("/notes/setPrivacy/:id", s.setPrivacy)

}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := map[string]string{
		"message": "Hello World",
	}
	c.Status(fiber.StatusOK)
	return c.JSON(resp)
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}
