package wordclouds

import "math"

type circle struct {
	cx       float64
	cy       float64
	radius   float64
	step     int
	maxSteps int
	points   []point
}
type point struct {
	x float64
	y float64
}

func newCircle(cx float64, cy float64, radius float64, maxSteps int) *circle {
	pts := make([]point, maxSteps, maxSteps)
	for i := 0; i < maxSteps; i++ {
		pts[i].x = cx + radius*math.Cos(float64(i)*(2*math.Pi/float64(maxSteps)))
		pts[i].y = cy + radius*math.Sin(float64(i)*(2*math.Pi/float64(maxSteps)))
	}

	return &circle{
		cx:       cx,
		cy:       cy,
		radius:   radius,
		step:     0,
		maxSteps: maxSteps,
		points:   pts,
	}
}

func (c *circle) positions() []point {
	return c.points
}
