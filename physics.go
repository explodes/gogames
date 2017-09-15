package games

import "github.com/faiface/pixel"

type Physics struct {
	Position     pixel.Vec
	Velocity     pixel.Vec
	Acceleration pixel.Vec
}

var _ Updater = &Physics{}

func NewPhysics() *Physics {
	return NewPhysicsWithAcceleration(0, 0, 0, 0, 0, 0)
}

func NewPhysicsWithPosition(x, y float64) *Physics {
	return NewPhysicsWithAcceleration(x, y, 0, 0, 0, 0)
}

func NewPhysicsWithVelocity(x, y, dx, dy float64) *Physics {
	return NewPhysicsWithAcceleration(x, y, dx, dy, 0, 0)
}

func NewPhysicsWithAcceleration(x, y, dx, dy, ddx, ddy float64) *Physics {
	return &Physics{
		Position:     pixel.V(x, y),
		Velocity:     pixel.V(dx, dy),
		Acceleration: pixel.V(ddx, ddy),
	}
}

func (p *Physics) Update(dt float64) {
	p.Velocity = p.Velocity.Add(p.Acceleration.Scaled(dt))
	p.Position = p.Position.Add(p.Velocity.Scaled(dt))
	p.Acceleration = pixel.ZV
}

func (p *Physics) Force(x, y float64) {
	p.Acceleration = p.Acceleration.Add(pixel.V(x, y))
}
