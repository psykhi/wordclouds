package wordclouds

import "fmt"

type Box struct {
	top    float64
	left   float64
	right  float64
	bottom float64
}

func (a *Box) x() float64 {
	return a.left
}

func (a *Box) y() float64 {
	return a.bottom
}

func (a *Box) w() float64 {
	return a.right - a.left
}

func (a *Box) h() float64 {
	return a.top - a.bottom
}

func (a *Box) fits(width float64, height float64) bool {
	return a.bottom > 0 && a.top < height && a.left > 0 && a.right < width
}
func (a *Box) overlaps(b *Box) bool {
	return a.left <= b.right && a.right >= b.left && a.top >= b.bottom && a.bottom <= b.top
}
func (a *Box) overlapsRaw(top float64, left float64, right float64, bottom float64) bool {
	return a.left <= right && a.right >= left && a.top >= bottom && a.bottom <= top
}

func (a *Box) String() string {
	return fmt.Sprintf("[x %f y %f w %f h %f]", a.x(), a.y(), a.w(), a.h())
}
