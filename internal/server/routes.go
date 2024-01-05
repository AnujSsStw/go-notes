package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	in "go-notes/internal"
	"log"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"golang.org/x/crypto/bcrypt"
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

func (s *FiberServer) setPrivacy(c *fiber.Ctx) error {
	if err := s.db.SetPrivacy(c.Locals("user").(string), c.Params("id")); err != nil {
		return err
	}
	return c.JSON(c.BaseURL() + "/api/notes/" + c.Params("id") + "/share")
}

func (s *FiberServer) getNotes(c *fiber.Ctx) error {
	if notes, err := s.db.GetNotes(c.Locals("user").(string)); err != nil {
		return err
	} else {
		return c.JSON(notes)
	}
}
func (s *FiberServer) getNote(c *fiber.Ctx) error {
	if note, err := s.db.GetNote(c.Locals("user").(string), c.Params("id")); err != nil {
		return err
	} else {
		return c.JSON(note)
	}
}
func (s *FiberServer) createNote(c *fiber.Ctx) error {
	n := new(in.Note)
	if err := c.BodyParser(n); err != nil {
		return err
	}
	n.UserId = c.Locals("user").(string)
	if err := s.db.CreateNote(n); err != nil {
		return err
	}

	return c.JSON("created")
}
func (s *FiberServer) updateNote(c *fiber.Ctx) error {
	n := new(in.Note)
	if err := c.BodyParser(n); err != nil {
		return err
	}

	var flag string

	if len(n.Text) > 0 && len(n.Title) > 0 {
		fmt.Println("updating both text and title")
		flag = "BOTH"
	} else if len(n.Text) > 0 {
		fmt.Println("updating text")
		flag = "TEXT"
	} else if len(n.Title) > 0 {
		fmt.Println("updating title")
		flag = "TITLE"
	}

	if err := s.db.UpdateNote(c.Locals("user").(string), c.Params("id"), flag, n); err != nil {
		return err
	}
	return c.JSON("Updated")
}
func (s *FiberServer) deleteNote(c *fiber.Ctx) error {
	if err := s.db.DeleteNote(c.Locals("user").(string), c.Params("id")); err != nil {
		return err
	} else {
		return c.JSON("Deleted")
	}
}
func (s *FiberServer) shareNote(c *fiber.Ctx) error {
	if note, err := s.db.GetNoteById(c.Params("id")); err != nil {
		return err
	} else {
		return c.JSON(note)
	}
}
func (s *FiberServer) searchNote(c *fiber.Ctx) error {
	m := c.Queries()
	if notes, err := s.db.SearchNote(c.Locals("user").(string), m["q"]+":*"); err != nil {
		return err
	} else {
		return c.JSON(notes)
	}
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
