package utils

import (
	"github.com/christ-saragih/project-management-be/models"
	"github.com/google/uuid"
)

func SortListsByPosition(lists []models.List, order []uuid.UUID) []models.List {
	if len(order) == 0 {
		return lists
	}
	
	ordered := make([]models.List,0,len(order))

	listMap := make(map[uuid.UUID]models.List)

	for _, list := range lists {
		listMap[list.PublicID] = list
	}

	// urutan sesuai order
	for _, id := range order {
		if list, ok := listMap[id]; ok {
			ordered = append(ordered, list)
		}
	}
	return ordered
}