package objects

import (
	"github.com/explodes/practice/games"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
)

var _ games.Updater = &Toon{}
var _ games.Drawer = &Toon{}

type Apple struct {
	*games.Physics
	Grower bool
}

func (a *Apple) Draw(imd *imdraw.IMDraw) {
	if a.Grower {
		imd.Color = colornames.Red
	} else {
		imd.Color = colornames.Blue
	}
	imd.Push(a.Position)
	imd.Circle(3, 0)
}
