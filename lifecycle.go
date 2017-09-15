package games

import (
	"github.com/faiface/pixel/imdraw"
)

type Updater interface {
	Update(dt float64)
}

type Drawer interface {
	Draw(imd *imdraw.IMDraw)
}
