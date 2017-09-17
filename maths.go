package games

import (
	"github.com/faiface/pixel"
	"math"
)

func Distance(v1, v2 pixel.Vec) float64 {
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func Near(v1, v2 pixel.Vec, distance float64) bool {
	return Distance(v1, v2) <= distance
}
func LimitWithinBounds(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func LimitWithinVec(value float64, vec pixel.Vec) float64 {
	return LimitWithinBounds(value, vec.X, vec.Y)
}

func LimitWithinRect(v pixel.Vec, r pixel.Rect) pixel.Vec {
	return pixel.Vec{
		X: LimitWithinBounds(v.X, r.Min.X, r.Max.X),
		Y: LimitWithinBounds(v.Y, r.Min.Y, r.Max.Y),
	}
}

func SignedSqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	if x < 0 {
		return -math.Sqrt(-x)
	}
	return math.Sqrt(x)
}
