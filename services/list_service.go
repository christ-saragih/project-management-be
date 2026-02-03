package services

import (
	"errors"
	"fmt"

	"github.com/christ-saragih/project-management-be/config"
	"github.com/christ-saragih/project-management-be/models"
	"github.com/christ-saragih/project-management-be/models/types"
	"github.com/christ-saragih/project-management-be/repositories"
	"github.com/christ-saragih/project-management-be/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListService interface {
	GetByBoardID(boardPublicID string) (*ListWithOrder, error)
	GetByID(id uint) (*models.List, error)
	GetByPublicID(publicID string) (*models.List, error)
	Create(list *models.List) error
	Update(list *models.List) error
	Delete(id uint) error
	UpdatePositions(boardPublicID string, positions []uuid.UUID) error
}

type ListWithOrder struct {
	Positions	[]uuid.UUID
	Lists		[]models.List
}

type listService struct {
	listRepo repositories.ListRepository
	listPositionRepo repositories.ListPositionRepository
	boardRepo repositories.BoardRepository
}

func NewListService(
	listRepo repositories.ListRepository,
	listPositionRepo repositories.ListPositionRepository,
	boardRepo repositories.BoardRepository) ListService {
	return &listService{listRepo,listPositionRepo,boardRepo}
}

func (s *listService) GetByBoardID(boardPublicID string) (*ListWithOrder, error) {

	_, err := s.boardRepo.FindByPublicID(boardPublicID)
	if err != nil {
		return nil, errors.New("Board not found")
	}

	position, err := s.listPositionRepo.GetListOrder(boardPublicID)
	if err != nil {
		return nil, errors.New("Failed to get list order: " + err.Error())
	}
	if len(position) == 0 {
		return nil, errors.New("List position is empty")
	}

	
	list, err := s.listRepo.FindByBoardID(boardPublicID)
	if err != nil {
		return nil, errors.New("Failed to get lists: " + err.Error())
	}

	// sorting by position
	orderedList := utils.SortListsByPosition(list, position)

	return &ListWithOrder{
		Positions: position,
		Lists: orderedList,
	}, nil

}

func (s *listService) GetByID(id uint) (*models.List, error) {
	return s.listRepo.FindByID(id)
}

func (s *listService) GetByPublicID(publicID string) (*models.List, error) {
	return s.listRepo.FindByPublicID(publicID)
}

func (s *listService) Create(list *models.List) error {
	// validasi board
	board, err := s.boardRepo.FindByPublicID(list.BoardPublicID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Board not found")
		}

		return fmt.Errorf("Failed to get board: %w", err)
	}

	list.BoardInternalID = board.InternalID

	if list.PublicID == uuid.Nil {
		list.PublicID = uuid.New()
	}

	// transaction
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} 
	}()

	// save list
	if err := tx.Create(list).Error;  err != nil {
		tx.Rollback()
		return fmt.Errorf("Failed to create list: %w", err)
	}

	// update list position
	var position models.ListPosition
	res := tx.Where("board_internal_id = ?", board.InternalID).First(&position)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		position = models.ListPosition{
			PublicID: uuid.New(),
			BoardID: board.InternalID,
			ListOrder: types.UUIDArray{list.PublicID},
		}
		if err := tx.Create(&position).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("Failed to create list position: %w", err)
		}
	} else if res.Error != nil {
		tx.Rollback()
		return fmt.Errorf("Failed to get list position: %w", res.Error)
	} else {
		position.ListOrder = append(position.ListOrder, list.PublicID)

		if err := tx.Model(&position).Update("list_order", position.ListOrder).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("Failed to update list position: %w", err)
		}
	}

	// commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}

	return nil
}

func (s *listService) Update(list *models.List) error {
	return s.listRepo.Update(list)
}

func (s *listService) Delete(id uint) error {
	return s.listRepo.Delete(id)
}

func (s *listService) UpdatePositions(boardPublicID string, positions []uuid.UUID) error {
	// validasi board
	board, err := s.boardRepo.FindByPublicID(boardPublicID)
	if err != nil {
		return errors.New("Board not found")
	}

	position, err := s.listPositionRepo.GetByBoard(board.PublicID.String())
	if err != nil {
		return errors.New("List position not found")
	}

	position.ListOrder = positions

	return s.listPositionRepo.UpdateListOrder(position)
}