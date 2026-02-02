package services

import (
	"errors"

	"github.com/christ-saragih/project-management-be/models"
	"github.com/christ-saragih/project-management-be/repositories"
	"github.com/google/uuid"
)

type BoardService interface {
	Create(board *models.Board) error
	Update(board *models.Board) error
	GetByPublicID(publicID string) (*models.Board, error)
	AddMembers(boardPublicID string, userPublicIDs []string) error
	RemoveMembers(boardPublicID string, userPublicIDs []string) error
	GetAllByUserPaginated(userPublicID, filter, sort string, limit, offset int) ([]models.Board, int64, error)
}

type boardService struct {
	boardRepo repositories.BoardRepository
	userRepo repositories.UserRepository
	boardMemberRepo repositories.BoardMemberRepository
}

func NewBoardService(
	boardRepo repositories.BoardRepository, 
	userRepo repositories.UserRepository, 
	boardMemberRepo repositories.BoardMemberRepository) BoardService {
	return &boardService{ boardRepo, userRepo, boardMemberRepo }
}

func (s *boardService) Create(board *models.Board) error {
	user, err := s.userRepo.FindByPublicID(board.OwnerPublicID.String())
	if err != nil {
		return errors.New("Owner not found")
	}

	board.PublicID = uuid.New()
	board.OwnerID = user.InternalID
	return s.boardRepo.Create(board)
}

func (s *boardService) Update(board *models.Board) error {
	return s.boardRepo.Update(board)
}

func (s *boardService) GetByPublicID(publicID string) (*models.Board, error) {
	return s.boardRepo.FindByPublicID(publicID)
}

func (s *boardService) AddMembers(boardPublicID string, userPublicIDs []string) error {
	board, err := s.boardRepo.FindByPublicID(boardPublicID)
	if err != nil {
		return errors.New("Board not found")
	}

	var userInternalIDs []uint
	for _, userPublicID := range userPublicIDs {
		user, err := s.userRepo.FindByPublicID(userPublicID)
		if err != nil {
			return errors.New("User not found: " + userPublicID)
		}
		userInternalIDs = append(userInternalIDs, uint(user.InternalID))
	}

	existingMembers, err := s.boardMemberRepo.GetMembers(board.PublicID.String())
	if err != nil {
		return err
	}

	memberMap := make(map[uint]bool)
	for _, member := range existingMembers {
		memberMap[uint(member.InternalID)] = true
	}

	var newMemberIDs []uint
	for _, userID := range userInternalIDs {
		if !memberMap[userID] {
			newMemberIDs = append(newMemberIDs, userID)
		}
	}
	if len(newMemberIDs) == 0 {
		return nil
	}

	return s.boardRepo.AddMember(uint(board.InternalID), newMemberIDs)
}

func (s *boardService) RemoveMembers(boardPublicID string, userPublicIDs []string) error {
	board, err := s.boardRepo.FindByPublicID(boardPublicID)
	if err != nil {
		return errors.New("Board not found")
	}

	var userInternalIDs []uint
	for _, userPublicID := range userPublicIDs {
		user, err := s.userRepo.FindByPublicID(userPublicID)
		if err != nil {
			return errors.New("User not found: " + userPublicID)
		}
		userInternalIDs = append(userInternalIDs, uint(user.InternalID))
	}

	existingMembers, err := s.boardMemberRepo.GetMembers(board.PublicID.String())
	if err != nil {
		return err
	}

	memberMap := make(map[uint]bool)
	for _, member := range existingMembers {
		memberMap[uint(member.InternalID)] = true
	}

	var membersToRemove []uint
	for _, userID := range userInternalIDs {
		if memberMap[userID] {
			membersToRemove = append(membersToRemove, userID)
		}
	}

	return s.boardRepo.RemoveMember(uint(board.InternalID), membersToRemove)

}

func (s *boardService) GetAllByUserPaginated(userPublicID, filter, sort string, limit, offset int) ([]models.Board, int64, error) {
	return s.boardRepo.FindAllByUserPaginated(userPublicID, filter, sort, limit, offset)
}