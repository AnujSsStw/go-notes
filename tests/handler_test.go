package tests

import (
	"go-notes/internal/server"
	"io"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/utils"
)

func TestRateLimiter(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(limiter.New(limiter.Config{
		Max:               50,
		Expiration:        2 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	// Inject the Fiber app into the server
	s := &server.FiberServer{App: app}
	// Define a route in the Fiber app
	app.Get("/", s.HelloWorldHandler)

	var wg sync.WaitGroup
	singleRequest := func(wg *sync.WaitGroup) {
		defer wg.Done()
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		utils.AssertEqual(t, nil, err)
		expected := "{\"message\":\"Hello World\"}"
		utils.AssertEqual(t, expected, string(body))
	}

	for i := 0; i <= 49; i++ {
		wg.Add(1)
		go singleRequest(&wg)
	}

	wg.Wait()

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 429, resp.StatusCode)

	time.Sleep(3 * time.Second)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 200, resp.StatusCode)
}
