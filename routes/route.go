package routes

import (
	"log"

	"github.com/christ-saragih/project-management-be/config"
	"github.com/christ-saragih/project-management-be/controllers"
	"github.com/christ-saragih/project-management-be/utils"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/joho/godotenv"
)

func Setup(app *fiber.App, 
	uc *controllers.UserController, 
	bc *controllers.BoardController,
	lc *controllers.ListController,) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error load .env file")
	}

	app.Post("/v1/auth/register", uc.Register)
	app.Post("/v1/auth/login", uc.Login)
	
	api := app.Group("/api/v1", jwtware.New(jwtware.Config{
		SigningKey: []byte(config.AppConfig.JWTSecret),
		ContextKey: "user",
		ErrorHandler: func (c *fiber.Ctx, err error) error {
			return utils.Anauthorized(c, "Error unauthorized", err.Error())
		},
	}))

	userGroup := api.Group("/users")
	userGroup.Get("/page", uc.GetUserPagination)
	userGroup.Get("/:id", uc.GetUser)
	userGroup.Put("/:id", uc.UpdateUser)
	userGroup.Delete("/:id", uc.DeleteUser)

	boardGroup := api.Group("/boards")
	boardGroup.Post("/", bc.CreateBoard)
	boardGroup.Put("/:id", bc.UpdateBoard)
	boardGroup.Post("/:id/members", bc.AddBoardMembers)
	boardGroup.Delete("/:id/members", bc.RemoveBoardMembers)
	boardGroup.Get("/my", bc.GetMyBoardPaginated)
	boardGroup.Get("/:board_id/lists", lc.GetListOnBoard)

	listGroup := api.Group("/lists")
	listGroup.Post("/", lc.CreateList)
	listGroup.Put("/:id", lc.UpdateList)
	listGroup.Delete("/:id", lc.DeleteList)
	// listGroup.Put("/positions/:boardPublicID", lc.UpdateListPositions)
}