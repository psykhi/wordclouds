package wordclouds

import (
	"github.com/satori/go.uuid"
	"fmt"
)

type uniqueBox struct {
	uuid.UUID
	b *Box
}

type spatialHashMap struct {
	mat      [][][]*uniqueBox
	rw       float64
	rh       float64
	gridSize int
}

func (s *spatialHashMap) TestCollision(b *Box, test func(a *Box, b *Box) bool) (bool, int) {
	overlaps := 0
	top, left, right, bottom := s.toGridCoords(b)
	for i := left; i <= right; i++ {
		for j := bottom; j <= top; j++ {
			for _, ub := range s.mat[i][j] {
				overlaps++
				if test(ub.b, b) {
					return true, overlaps
				}
			}
		}
	}
	return false, overlaps
}

func (s *spatialHashMap) Add(b *Box) {
	id, err := uuid.NewV4()
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return
	}
	top, left, right, bottom := s.toGridCoords(b)
	for i := left; i <= right; i++ {
		for j := bottom; j <= top; j++ {
			s.mat[i][j] = append(s.mat[i][j], &uniqueBox{id, b})
		}
	}
}

func newSpatialHashMap(windowWidth float64, windowHeight float64, gridSize int) *spatialHashMap {
	rw := windowWidth / float64(gridSize)
	rh := windowHeight / float64(gridSize)

	mat := make([][][]*uniqueBox, gridSize)
	for i := 0; i < gridSize; i++ {
		mat[i] = make([][]*uniqueBox, gridSize)
		for j := 0; j < gridSize; j++ {
			mat[i][j] = make([]*uniqueBox, 0)
		}
	}

	return &spatialHashMap{
		mat:      mat,
		rw:       rw,
		rh:       rh,
		gridSize: gridSize,
	}
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *spatialHashMap) toGridCoords(b *Box) (int, int, int, int) {
	return min(int(b.Top/s.rh), s.gridSize-1), int(b.Left / s.rw), min(int(b.Right/s.rw), s.gridSize-1), int(b.Bottom / s.rh)
}
