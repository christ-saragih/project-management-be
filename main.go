package main

import (
	"log"

	"github.com/christ-saragih/project-management-be/config"
	"github.com/christ-saragih/project-management-be/controllers"
	"github.com/christ-saragih/project-management-be/database/seed"
	"github.com/christ-saragih/project-management-be/repositories"
	"github.com/christ-saragih/project-management-be/routes"
	"github.com/christ-saragih/project-management-be/services"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config.LoadEnv()
	config.ConnectDB()

	seed.SeedAdmin()

	app := fiber.New()

	userRepo := repositories.NewUserRepository()
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	routes.Setup(app, userController)

	port := config.AppConfig.AppPort
	log.Println("Server is running on port:", port)
	log.Fatal(app.Listen(":" + port))
}