package server

import (
	"fmt"
	in "go-notes/internal"

	"github.com/gofiber/fiber/v2"
)

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
		c.Status(fiber.StatusBadRequest)
		return err
	} else {
		return c.JSON(map[string]string{
			"message": "Deleted",
		})
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
