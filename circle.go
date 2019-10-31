package wordclouds

import "math"

type circle struct {
	cx       float64
	cy       float64
	radius   float64
	step     int
	maxSteps int
}
type point struct {
	x float64
	y float64
}

func newCircle(cx float64, cy float64, radius float64, maxSteps int) *circle {
	return &circle{
		cx:       cx,
		cy:       cy,
		radius:   radius,
		step:     0,
		maxSteps: maxSteps,
	}
}

func (c *circle) positions() []point {
	pts := make([]point, c.maxSteps, c.maxSteps)
	for i := 0; i < c.maxSteps; i++ {
		pts[i].x = c.cx + c.radius*math.Cos(float64(i)*(2*math.Pi/float64(c.maxSteps)))
		pts[i].y = c.cy + c.radius*math.Sin(float64(i)*(2*math.Pi/float64(c.maxSteps)))
	}
	return pts
}
