package main

import (
	"fmt"
	"github.com/explodes/practice/games"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"math"
	"math/rand"
	"os"
	"time"
)

const (
	title                     = "Lights Out"
	width, height             = 700, 700
	canvasWidth, canvasHeight = 700, 700
	maxFps                    = 24

	gridSideLength = 8
	gridSquares    = gridSideLength * gridSideLength

	starPoints                 = 5
	starRotateDegreesPerSecond = 96
	starInnerRadiusFactor      = 0.5
	starColorTransitionSpeed   = 0.5
)

var squareColors = []pixel.RGBA{
	pixel.RGB(1, 0.1, 0.1),
	pixel.RGB(0.3, 0.3, 1),
	pixel.RGB(0.4, 0.5, 0.2),
}

type Grid struct {
	squares [gridSquares]bool
	colors  [gridSquares]pixel.RGBA
}

type Star struct {
	rotationDeg     float64
	color           pixel.RGBA
	colorTransition float64
}

func (s *Star) Update(dt float64) {
	s.rotationDeg += starRotateDegreesPerSecond * dt
	s.colorTransition += starColorTransitionSpeed * 360 * dt
	s.color.G = 0.5 + 0.25*(1+math.Cos(degToRad(s.colorTransition)))
}

func degToRad(d float64) float64 {
	return d * math.Pi / 180
}

func (s *Star) Draw(imd *imdraw.IMDraw, canvas *pixelgl.Canvas) {
	const pointDegDelta = 360.0 / (starPoints)
	const pointInnerDegDelta = 360.0 / (2.0 * starPoints)
	//fmt.Println(pointDegDelta)

	maxMagnitude := math.Min(canvasWidth, canvasHeight) / 2
	innerMagnitude := maxMagnitude * starInnerRadiusFactor

	imd.Color = s.color

	for point := 0; point < starPoints; point++ {
		degrees := float64(point)*pointDegDelta + 90 - s.rotationDeg

		x0, y0 := math.Cos(degToRad(degrees+pointInnerDegDelta))*innerMagnitude+canvasWidth/2, math.Sin(degToRad(degrees+pointInnerDegDelta))*innerMagnitude+canvasHeight/2
		x1, y1 := math.Cos(degToRad(degrees))*maxMagnitude+canvasWidth/2, math.Sin(degToRad(degrees))*maxMagnitude+canvasHeight/2
		x2, y2 := math.Cos(degToRad(degrees-pointInnerDegDelta))*innerMagnitude+canvasWidth/2, math.Sin(degToRad(degrees-pointInnerDegDelta))*innerMagnitude+canvasHeight/2

		imd.Push(pixel.V(x0, y0))
		imd.Push(pixel.V(x1, y1))
		imd.Push(pixel.V(x2, y2))
		imd.Polygon(0)

		imd.Push(pixel.V(x0, y0))
		imd.Push(pixel.V(canvasWidth/2, canvasHeight/2))
		imd.Push(pixel.V(x2, y2))
		imd.Polygon(0)
	}
}

func newGrid() *Grid {
	g := &Grid{}

	for i := 0; i < gridSquares; i++ {
		g.squares[i] = true
		g.colors[i] = squareColor(i)
	}

	return g
}

func squareColor(index int) pixel.RGBA {

	var mod int
	if gridSideLength%2 == 0 {
		mod = 3
	} else {
		mod = 2
	}

	return squareColors[index%mod]
}

func run() {

	rand.Seed(time.Now().UnixNano())

	cfg := pixelgl.WindowConfig{
		Title:  title,
		Bounds: pixel.R(0, 0, width, height),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		exitWith(err, "unable to create window")
	}

	canvas := pixelgl.NewCanvas(pixel.R(0, 0, canvasWidth, canvasHeight))

	imd := imdraw.New(nil)
	imd.Precision = 32

	fpsLimit := games.NewFpsLimiter(maxFps)

	grid := newGrid()
	star := &Star{
		rotationDeg: 0,
		color:       pixel.RGB(1, 1, 0),
	}

	//last := time.Now()

	// width and height in CANVAS pixels of a given square
	const dx = float64(canvasWidth) / float64(gridSideLength)
	const dy = float64(canvasHeight) / float64(gridSideLength)

	// width and height in WINDOW pixels of a given square
	const ssx = float64(width) / float64(gridSideLength)
	const ssy = float64(height) / float64(gridSideLength)

	moves := 0
	winner := false

	last := time.Now()

	for !win.Closed() {
		fpsLimit.StartFrame()
		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.JustPressed(pixelgl.KeyR) {
			grid = newGrid()
			winner = false
			moves = 0
		}

		if !winner && win.JustPressed(pixelgl.MouseButton1) {
			pos := win.MousePosition()
			x := int(pos.X / ssx)
			y := int(pos.Y / ssy)

			i := x + y*gridSideLength
			if i >= gridSquares {
				goto update
			}

			moves++

			grid.squares[i] = !grid.squares[i]
			if x > 0 {
				index := (x - 1) + y*gridSideLength
				grid.squares[index] = !grid.squares[index]
			}
			if x < gridSideLength-1 {
				index := (x + 1) + y*gridSideLength
				grid.squares[index] = !grid.squares[index]
			}
			if y > 0 {
				index := x + (y-1)*gridSideLength
				grid.squares[index] = !grid.squares[index]
			}
			if y < gridSideLength-1 {
				index := x + (y+1)*gridSideLength
				grid.squares[index] = !grid.squares[index]
			}

		}

		for _, square := range grid.squares {
			if square {
				goto update
			}
		}
		winner = true

	update:
		star.Update(dt)

		//draw:
		imd.Clear()

		if winner {
			star.Draw(imd, canvas)
		} else {
			// draw game into image
			for i := 0; i < gridSquares; i++ {
				if !grid.squares[i] {
					continue
				}

				x := float64(i%gridSideLength) * dx
				y := float64(i/gridSideLength) * dy
				bottomleft := pixel.V(x, y)
				topright := pixel.V(x+dx, y+dy)

				imd.Color = grid.colors[i]
				imd.Push(bottomleft, topright)
				imd.Rectangle(0)
			}
		}

		// draw image into canvas
		canvas.Clear(colornames.Black)
		imd.Draw(canvas)

		// draw canvas into window
		games.DrawCanvasInWindow(colornames.White, win, canvas)

		fpsLimit.WaitForNextFrame()
		win.SetTitle(fmt.Sprintf("%s | moves: %d | fps %.0f", title, moves, fpsLimit.CurrentFrameFps()))
	}
}

func drawStar(imd *imdraw.IMDraw, bounds pixel.Rect) {
	imd.Color = randomNiceColor()
	imd.Push(bounds.Min, bounds.Max)
	imd.Rectangle(0)
}

func randomNiceColor() pixel.RGBA {
again:
	r := rand.Float64()
	g := rand.Float64()
	b := rand.Float64()
	magnitude := math.Sqrt(r*r + g*g + b*b)
	if magnitude == 0 {
		goto again
	}
	return pixel.RGB(r/magnitude, g/magnitude, b/magnitude)
}

func main() {
	pixelgl.Run(run)
}

func exitWith(err error, msg string, args ...interface{}) {
	fmt.Printf("%s: %v\n", fmt.Sprintf(msg, args...), err)
	os.Exit(2)
}
