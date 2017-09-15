package main

import (
	"fmt"
	"github.com/explodes/practice/games"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"math"
	"math/rand"
	"os"
	"time"
)

const (
	width        = 1024
	height       = 768
	canvasWidth  = width / 2
	canvasHeight = height / 2
	maxFps       = 60
	slowmoFactor = 10
)

type particle struct {
	*games.Physics

	color pixel.RGBA

	size      float64
	shrinkage float64

	shouldTwinkle bool
	twinkleLife   float64
	twinkled      bool
}

var _ games.Updater = &particle{}
var _ games.Drawer = &particle{}

func (d *particle) Update(dt float64) {
	d.Force(0, -250) // gravity
	d.Physics.Update(dt)

	d.color.A *= dt * 0.5
	d.size *= 1 - (dt * d.shrinkage)

	if d.shouldTwinkle {
		d.twinkleLife -= 12 * dt // 12 twinkles per second
		if !d.twinkled && d.twinkleLife <= -12 {
			d.twinkleLife = 20 + 15*rand.Float64()
			d.twinkled = true
		}
	}

}

func (d *particle) Draw(imd *imdraw.IMDraw) {

	if d.size < 0.1 {
		return
	}

	if d.shouldTwinkle {
		// if we're done twinkling, do not draw
		if d.twinkled && d.twinkleLife <= 0 {
			return
		}
		// if we're in the off-phase of twinkling, do not draw
		if d.twinkleLife > 0 && int(d.twinkleLife)%2 == 0 {
			return
		}
	}

	size := d.size * 0.5
	imd.Color = d.color
	imd.Push(d.Position)
	imd.Circle(size, 0)
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

func randomFireColor() pixel.RGBA {
again:
	r := 0.30 + 0.70*rand.Float64()
	g := 0.10 + 0.35*rand.Float64()
	b := 0.00 + 0.10*rand.Float64()
	magnitude := math.Sqrt(r*r + g*g + b*b)
	if magnitude == 0 {
		goto again
	}
	return pixel.RGB(r/magnitude, g/magnitude, b/magnitude)
}

func makeParticles() []*particle {
	shouldTwinkle := rand.Intn(100) > 25 // ~25% chance of NOT twinkling
	numParticles := 100 + rand.Intn(2000)
	//numParticles := 1
	particles := make([]*particle, 0, numParticles)
	for i := 0; i < numParticles; i++ {
		explosive := 600 + 200*rand.Float64()
		force := -(explosive / 2) + explosive*rand.Float64()
		angle := 2 * math.Pi * rand.Float64()
		dir := pixel.V(math.Cos(angle)*force, math.Sin(angle)*force)

		particle := &particle{
			Physics:       games.NewPhysicsWithVelocity(0, 0, dir.X, dir.Y),
			color:         randomFireColor(),
			size:          2 + 5*rand.Float64(),
			shrinkage:     0.7 + 0.1*rand.Float64(),
			shouldTwinkle: shouldTwinkle,
		}
		particles = append(particles, particle)
	}
	return particles
}

func run() {
	rand.Seed(time.Now().UnixNano())

	cfg := pixelgl.WindowConfig{
		Title:     "Explosion",
		Bounds:    pixel.R(0, 0, width, height),
		VSync:     true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		exitWith(err, "unable to create window")
	}

	canvas := pixelgl.NewCanvas(pixel.R(-canvasWidth/2, -canvasHeight/2, canvasWidth/2, canvasHeight/2))

	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	instructionsTxt := text.New(pixel.V(canvas.Bounds().Min.X+10, canvas.Bounds().Max.Y-basicAtlas.Ascent()-10), basicAtlas)
	instructionsTxt.WriteString("Press ENTER to explode, hold SPACE to slow down time")

	basicTxt := text.New(pixel.V(canvas.Bounds().Min.X+10, canvas.Bounds().Max.Y-2*basicAtlas.Ascent()-10), basicAtlas)

	particles := makeParticles()

	imd := imdraw.New(nil)
	imd.Precision = 32

	fpsLimit := games.NewFpsLimiter(maxFps)

	canvas.Clear(colornames.Black)

	last := time.Now()

	for !win.Closed() {
		fpsLimit.StartFrame()

		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.Pressed(pixelgl.KeySpace) {
			dt /= slowmoFactor
		}

		if win.JustPressed(pixelgl.KeyEnter) {
			canvas.Clear(colornames.Black)
			particles = makeParticles()
		}

		for _, d := range particles {
			d.Update(dt)
		}

		canvas.Clear(colornames.Black)
		imd.Clear()
		for _, d := range particles {
			d.Draw(imd)
		}
		imd.Draw(canvas)

		instructionsTxt.Draw(canvas, pixel.IM)
		basicTxt.Draw(canvas, pixel.IM)

		// stretch the canvas to the window
		win.Clear(colornames.White)
		win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
			math.Min(
				win.Bounds().W()/canvas.Bounds().W(),
				win.Bounds().H()/canvas.Bounds().H(),
			),
		).Moved(win.Bounds().Center()))
		canvas.Draw(win, pixel.IM.Moved(canvas.Bounds().Center()))
		win.Update()

		fpsLimit.WaitForNextFrame()

		basicTxt.Clear()
		basicTxt.Dot = basicTxt.Orig
		fmt.Fprintf(basicTxt, "fps: %.0f", fpsLimit.CurrentFrameFps())
	}
}

func main() {
	pixelgl.Run(run)
}

func exitWith(err error, msg string, args ...interface{}) {
	fmt.Printf("%s: %v\n", fmt.Sprintf(msg, args...), err)
	os.Exit(2)
}
