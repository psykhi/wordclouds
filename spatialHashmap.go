package wordclouds

import "github.com/satori/go.uuid"

type uniqueBox struct {
	uuid.UUID
	b *Box
}

type spatialHashMap struct {
	mat [][][]*uniqueBox
	rw  float64
	rh  float64
}

func (s *spatialHashMap) Candidates(b *Box) []*Box {
	res := make([]*Box, 0, 16)
	top, left, right, bottom := s.toGridCoords(b)
	for i := left; i <= right; i++ {
		for j := bottom; j <= top; j++ {
			for _, ub := range s.mat[i][j] {
				res = append(res, ub.b)
			}

		}
	}
	return res
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
	id := uuid.NewV4()
	top, left, right, bottom := s.toGridCoords(b)
	for i := left; i <= right; i++ {
		for j := bottom; j <= top; j++ {
			s.mat[i][j] = append(s.mat[i][j], &uniqueBox{id, b})
		}
	}
}

func NewSpatialHashMap(windowWidth float64, windowHeight float64, gridSize int) *spatialHashMap {
	rw := windowWidth / float64(gridSize)
	rh := windowHeight / float64(gridSize)

	mat := make([][][]*uniqueBox, int(windowWidth))
	for i := 0; i < int(windowWidth); i++ {
		mat[i] = make([][]*uniqueBox, int(windowHeight))
		for j := 0; j < int(windowHeight); j++ {
			mat[i][j] = make([]*uniqueBox, 0)
		}
	}

	return &spatialHashMap{
		mat: mat,
		rw:  rw,
		rh:  rh,
	}
}

func (s *spatialHashMap) toGridCoords(b *Box) (int, int, int, int) {
	return int(b.top / s.rh), int(b.left / s.rw), int(b.right / s.rw), int(b.bottom / s.rh)
}
