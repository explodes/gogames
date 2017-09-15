package games

import (
	"time"
)

type FpsLimiter struct {
	wait       time.Duration
	frameStart time.Time
}

func NewFpsLimiter(maxFps int) *FpsLimiter {
	fpsLimiter := &FpsLimiter{}
	fpsLimiter.SetLimit(maxFps)
	return fpsLimiter
}

func (f *FpsLimiter) StartFrame() {
	f.frameStart = time.Now()
}

func (f *FpsLimiter) WaitForNextFrame() {
	time.Sleep(f.wait - time.Since(f.frameStart))
}

func (f *FpsLimiter) SetLimit(maxFps int) {
	f.wait = time.Second / time.Duration(maxFps)
}

func (f *FpsLimiter) CurrentFrameFps() float64 {
	return 1 / time.Since(f.frameStart).Seconds()
}
