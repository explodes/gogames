package graphical

import "time"

type FpsLimiter struct {
	wait      time.Duration
	startTime time.Time
}

func NewFpsLimiter(maxFps int) *FpsLimiter {
	fpsLimiter := &FpsLimiter{}
	fpsLimiter.SetLimit(maxFps)
	return fpsLimiter
}

func (f *FpsLimiter) StartFrame() {
	f.startTime = time.Now()
}

func (f *FpsLimiter) WaitForNextFrame() {
	time.Sleep(f.wait - time.Since(f.startTime))
}

func (f *FpsLimiter) SetLimit(maxFps int) {
	f.wait = time.Second / time.Duration(maxFps)
}

func (f *FpsLimiter) CurrentFrameFps() float32 {
	duration := time.Since(f.startTime)
	return float32(time.Second) / float32(duration)
}
