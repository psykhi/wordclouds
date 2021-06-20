package wordclouds

import "math"

const (
	SizeFunctionLinear      = "linear"
	SizeFunctionSqrt        = "sqrt"
	SizeFunctionSqrtInverse = "sqrtinverse"
)

// size function to work on a normalized interval [0:1]
type sizeFunction func(x float64) float64

// sizeLinear scales font 1:1
func sizeLinear(x float64) float64 {
	return x
}

// sizeSqrt scales based on sqrt function, meaning larger fonts earlier
func sizeSqrt(x float64) float64 {
	return math.Sqrt(x)
}

// sizeSqrt scales based on 1-sqrt(1-x) function, meaning larger fonts later
func sizeSqrtInverse(x float64) float64 {
	return 1 - math.Sqrt(1-x)
}
