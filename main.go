package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Println("Hello there")

	app := fiber.New()

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"msg": "hello there"})
	})

	log.Fatal(app.Listen("0.0.0.0:" + port))
}
