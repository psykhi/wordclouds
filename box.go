package wordclouds

import "fmt"

type Box struct {
	Top    float64
	Left   float64
	Right  float64
	Bottom float64
}

func (a *Box) x() float64 {
	return a.Left
}

func (a *Box) y() float64 {
	return a.Bottom
}

func (a *Box) w() float64 {
	return a.Right - a.Left
}

func (a *Box) h() float64 {
	return a.Top - a.Bottom
}

func (a *Box) fits(width float64, height float64) bool {
	return a.Bottom > 0 && a.Top < height && a.Left > 0 && a.Right < width
}
func (a *Box) overlaps(b *Box) bool {
	return a.Left <= b.Right && a.Right >= b.Left && a.Top >= b.Bottom && a.Bottom <= b.Top
}
func (a *Box) overlapsRaw(top float64, left float64, right float64, bottom float64) bool {
	return a.Left <= right && a.Right >= left && a.Top >= bottom && a.Bottom <= top
}

func (a *Box) String() string {
	return fmt.Sprintf("[x %f y %f w %f h %f]", a.x(), a.y(), a.w(), a.h())
}
