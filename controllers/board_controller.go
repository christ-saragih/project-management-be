package controllers

import (
	"math"
	"strconv"

	"github.com/christ-saragih/project-management-be/models"
	"github.com/christ-saragih/project-management-be/services"
	"github.com/christ-saragih/project-management-be/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type BoardController struct {
	service services.BoardService
}

func NewBoardController(s services.BoardService) *BoardController {
	return &BoardController{service: s}
}

func (c *BoardController) CreateBoard(ctx *fiber.Ctx) error {
	var userID uuid.UUID
	var err error 

	board := new(models.Board)
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	if err := ctx.BodyParser(board); err != nil {
		return utils.BadRequest(ctx, "Failed to parsing data", err.Error())
	}

	userID, err = uuid.Parse(claims["pub_id"].(string))
	if err != nil {
		return utils.BadRequest(ctx, "Failed to get user id", err.Error())
	}
	board.OwnerPublicID = userID

	if err := c.service.Create(board); err != nil {
		return utils.BadRequest(ctx, "Failed to create board", err.Error())
	}

	return utils.Success(ctx, "Board created successfully", board)
}

func (c *BoardController) UpdateBoard(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id")
	board := new(models.Board)

	if err := ctx.BodyParser(board); err != nil {
		return utils.BadRequest(ctx, "Failed to parsing data", err.Error())
	}

	if _, err := uuid.Parse(publicID); err != nil {
		return utils.BadRequest(ctx, "Invalid board id", err.Error())
	}

	existingBoard, err := c.service.GetByPublicID(publicID)
	if err != nil {
		return utils.NotFound(ctx, "Board not found", err.Error())
	}
	board.InternalID = existingBoard.InternalID
	board.PublicID = existingBoard.PublicID
	board.OwnerID = existingBoard.OwnerID
	board.OwnerPublicID = existingBoard.OwnerPublicID
	board.CreatedAt = existingBoard.CreatedAt

	if err := c.service.Update(board); err != nil {
		return utils.BadRequest(ctx, "Failed to update board", err.Error())
	}

	return utils.Success(ctx, "Board updated successfully", board)
}

func (c *BoardController) AddBoardMembers(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id")
	var userIDs []string

	if err := ctx.BodyParser(&userIDs); err != nil {
		return utils.BadRequest(ctx, "Failed to parsing data", err.Error())
	}

	if err := c.service.AddMembers(publicID, userIDs); err != nil {
		return utils.BadRequest(ctx, "Failed to add board members", err.Error())
	}

	return utils.Success(ctx, "Board members added successfully", nil)
}

func (c *BoardController) RemoveBoardMembers(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id")
	var userIDs []string

	if err := ctx.BodyParser(&userIDs); err != nil {
		return utils.BadRequest(ctx, "Failed to parsing data", err.Error())
	}
	
	if err := c.service.RemoveMembers(publicID, userIDs); err != nil {
		return utils.BadRequest(ctx, "Failed to remove board members", err.Error())
	}

	return utils.Success(ctx, "Board members removed successfully", nil)
}

func (c *BoardController) GetMyBoardPaginated(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["pub_id"].(string)

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset := (page - 1) * limit

	filter := ctx.Query("filter", "")
	sort := ctx.Query("sort", "")


	boards, total, err := c.service.GetAllByUserPaginated(userID, filter, sort, limit, offset)
	if err != nil {
		return utils.InternalServerError(ctx, "Failed to get boards", err.Error())
	}

	meta := utils.PaginationMeta{
		Page: page,
		Limit: limit,
		Total: int(total),
		TotalPage: int(math.Ceil(float64(total) / float64(limit))),
		Filter: filter,
		Sort: sort,
	}

	return utils.SuccessPagination(ctx, "Boards retrieved successfully", boards, meta)
}