package main

import (
	"fmt"
	"github.com/explodes/practice/games"
	"github.com/explodes/practice/games/appleseed/objects"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"math/rand"
	"os"
	"time"
)

const (
	title                     = "Appleseed"
	width, height             = 1024, 768
	canvasWidth, canvasHeight = width / 2, height / 2
	maxFps                    = 60
)

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
	win.SetSmooth(true)

	canvas := pixelgl.NewCanvas(pixel.R(0, 0, canvasWidth, canvasHeight))

	imd := imdraw.New(nil)
	imd.Precision = 32

	fpsLimit := games.NewFpsLimiter(maxFps)

	toon := &objects.Toon{
		Size:    3,
		Physics: games.NewPhysicsWithPosition(10, 10),
	}

	var apples []*objects.Apple

	for i := 0; i < 100; i++ {
		apples = append(apples, &objects.Apple{
			Physics: games.NewPhysicsWithPosition(rand.Float64()*canvasWidth, rand.Float64()*canvasHeight),
			Grower:  i%4 != 0,
		})
	}

	last := time.Now()
	score := 0

	for !win.Closed() {
		fpsLimit.StartFrame()
		dt := time.Since(last).Seconds()
		last = time.Now()

		const movespeed = 7500

		var dx, dy float64

		if win.Pressed(pixelgl.KeyUp) {
			if toon.Velocity.Y < 0 {
				dy += 200 * dt * movespeed
			}
			dy += movespeed
		}
		if win.Pressed(pixelgl.KeyDown) {
			if toon.Velocity.Y > 0 {
				dy -= 200 * dt * movespeed
			}
			dy -= movespeed
		}
		if win.Pressed(pixelgl.KeyLeft) {
			if toon.Velocity.X > 0 {
				dx -= 200 * dt * movespeed
			}
			dx -= movespeed
		}
		if win.Pressed(pixelgl.KeyRight) {
			if toon.Velocity.X < 0 {
				dx += 200 * dt * movespeed
			}
			dx += movespeed
		}
		toon.Move(dt*dx, dt*dy)

		toon.Update(dt)

		cb := canvas.Bounds()
		if toon.Position.X < cb.Min.X {
			toon.Position.X = cb.Min.X
			toon.Velocity.X = 0
			toon.Acceleration.X = 0
		} else if toon.Position.X > cb.Max.X {
			toon.Position.X = cb.Max.X
			toon.Velocity.X = 0
			toon.Acceleration.X = 0
		}
		if toon.Position.Y < cb.Min.Y {
			toon.Position.Y = cb.Min.Y
			toon.Velocity.Y = 0
			toon.Acceleration.Y = 0
		} else if toon.Position.Y > cb.Max.Y {
			toon.Position.Y = cb.Max.Y
			toon.Velocity.Y = 0
			toon.Acceleration.Y = 0
		}

		for _, apple := range apples {
			distance := games.Distance(apple.Position, toon.Position)
			if distance <= toon.Size {
				score += int(3 * toon.Size)
				if apple.Grower {
					toon.Grow()
				} else {
					toon.Shrink()
				}
				for i := 0; i < 10; i++ {
					newPos := pixel.V(rand.Float64()*canvasWidth, rand.Float64()*canvasHeight)
					if games.Distance(toon.Position, newPos) > toon.Size {
						apple.Position = newPos
						break
					}
				}
				apple.Velocity = pixel.ZV
				apple.Acceleration = pixel.ZV
			} else if distance <= 4*toon.Size {
				gx := 5 * dt * (toon.Position.X - apple.Position.X)
				gy := 5 * dt * (toon.Position.Y - apple.Position.Y)
				apple.Force(games.SignedSqrt(gx), games.SignedSqrt(gy))
				apple.Update(dt)
				apple.Position = games.LimitWithinRect(apple.Position, canvas.Bounds())
			}
		}

		imd.Clear()
		for _, apple := range apples {
			apple.Draw(imd)
		}
		toon.Draw(imd)

		canvas.Clear(colornames.Black)
		imd.Draw(canvas)

		games.DrawCanvasInWindow(colornames.White, win, canvas)

		fpsLimit.WaitForNextFrame()
		win.SetTitle(fmt.Sprintf("%s | score: %d | fps %.0f", title, score, fpsLimit.CurrentFrameFps()))
	}
}

func main() {
	pixelgl.Run(run)
}

func exitWith(err error, msg string, args ...interface{}) {
	fmt.Printf("%s: %v\n", fmt.Sprintf(msg, args...), err)
	os.Exit(2)
}
