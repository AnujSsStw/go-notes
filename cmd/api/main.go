package main

import (
	"fmt"
	"go-notes/internal/server"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

func main() {

	server := server.New()

	api := server.Group("/api")
	server.RegisterFiberRoutes(api)
	// server.Use(limiter.New(limiter.Config{
	// 	Max:               20,
	// 	Expiration:        30 * time.Second,
	// 	LimiterMiddleware: limiter.SlidingWindow{},
	// }))
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	err := server.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
