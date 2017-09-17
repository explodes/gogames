package games

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel"
	"math"
	"image/color"
)

func DrawCanvasInWindow(clearColor color.RGBA, win *pixelgl.Window, canvas *pixelgl.Canvas) {
	// stretch the canvas to the window
	win.Clear(clearColor)
	scale := math.Min( win.Bounds().W()/canvas.Bounds().W(), win.Bounds().H()/canvas.Bounds().H())
	win.SetMatrix(pixel.IM.Scaled(pixel.ZV, scale))
	canvas.Draw(win, pixel.IM.Moved(canvas.Bounds().Center()))
	win.Update()
}
