package objects

import (
	"github.com/explodes/gogames"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
	"math"
)

var _ games.Updater = &Toon{}
var _ games.Drawer = &Toon{}

type Toon struct {
	*games.Physics
	Size float64
}

func (t *Toon) Update(dt float64) {
	t.Physics.Update(dt)
	//t.Velocity = t.Velocity.Scaled(100 * dt)
}

func (t *Toon) Draw(imd *imdraw.IMDraw) {
	imd.Color = colornames.Yellow
	imd.Push(t.Position)
	imd.Circle(t.Size, 0)
}

func (t *Toon) Move(x, y float64) {
	t.Force(x/t.Size*2, y/t.Size*2)
}

func (t *Toon) Grow() {
	t.Size = math.Min(100, t.Size+0.5)
}

func (t *Toon) Shrink() {
	t.Size = math.Max(3, t.Size-0.5)
}
