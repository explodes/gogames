package main

import (
	"runtime"

	"fmt"
	"github.com/explodes/gogames"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"log"
	"math"
	"math/rand"
	"strings"
)

const (
	title         = "Conway's Game of Life"
	width, height = 1000, 1000
	fps           = 60

	rows, columns = 500, 500

	threshold = 0.15

	vertexShaderSource = `
#version 410

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform float timing;

out vec4 vertexColor;
in vec3 vert;

void main() {
    gl_Position = projection * camera * model * vec4(vert, 1);
	vertexColor = gl_Position;
}
` + "\x00"

	fragmentShaderSource = `
#version 410

out vec4 frag_colour;
in vec4 vertexColor;

uniform vec3 colorshift;

const float offset = 1f;

void main() {
	frag_colour = vec4((offset+vertexColor.x)*(0.5f+colorshift.x*0.5f), (offset+vertexColor.y)*(0.5f+colorshift.y*0.5f), (offset+vertexColor.y+vertexColor.x)*(0.5f+colorshift.z*0.5f), 1);
}
` + "\x00"

	low  = 0.0
	mid  = 0.5
	high = 1.0
)

var (
	triangle = []float32{
		mid, high, mid, // top
		low, low, mid, // left
		high, low, mid, // right
	}
	square = []float32{
		low, high, mid,
		low, low, mid,
		high, low, mid,

		low, high, mid,
		high, high, mid,
		high, low, mid,
	}
)

type cell struct {
	drawable uint32

	alive     bool
	aliveNext bool

	x, y int
}

func (c *cell) draw(modelUniform int32) {
	if !c.alive {
		return
	}

	trans := mgl32.Translate3D(float32(c.x), float32(c.y), 0.5)
	scale := mgl32.Scale3D(float32(width)/float32(columns), float32(height)/float32(rows), 1)

	model := scale.Mul4(trans)

	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	gl.BindVertexArray(c.drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))

}

// checkState determines the state of the cell for the next tick of the game.
func (c *cell) checkState(cells []*cell) {
	c.alive = c.aliveNext
	c.aliveNext = c.alive

	liveCount := c.liveNeighbors(cells)
	if c.alive {
		// 1. Any live cell with fewer than two live neighbours dies, as if caused by underpopulation.
		if liveCount < 2 {
			c.aliveNext = false
		}

		// 2. Any live cell with two or three live neighbours lives on to the next generation.
		if liveCount == 2 || liveCount == 3 {
			c.aliveNext = true
		}

		// 3. Any live cell with more than three live neighbours dies, as if by overpopulation.
		if liveCount > 3 {
			c.aliveNext = false
		}
	} else {
		// 4. Any dead cell with exactly three live neighbours becomes a live cell, as if by reproduction.
		if liveCount == 3 {
			c.aliveNext = true
		}
	}
}

// liveNeighbors returns the number of live neighbors for a cell.
func (c *cell) liveNeighbors(cells []*cell) int {
	var liveCount int
	add := func(x, y int) {
		// If we're at an edge, check the other side of the board.
		if x == columns {
			x = 0
		} else if x == -1 {
			x = columns - 1
		}
		if y == rows {
			y = 0
		} else if y == -1 {
			y = rows - 1
		}

		if cells[x+y*columns].alive {
			liveCount++
		}
	}

	add(c.x-1, c.y)   // To the left
	add(c.x+1, c.y)   // To the right
	add(c.x, c.y+1)   // up
	add(c.x, c.y-1)   // down
	add(c.x-1, c.y+1) // top-left
	add(c.x+1, c.y+1) // top-right
	add(c.x-1, c.y-1) // bottom-left
	add(c.x+1, c.y-1) // bottom-right

	return liveCount
}

func main() {
	runtime.LockOSThread()

	window, err := initGlfw()
	if err != nil {
		exitWith(err, "cannot init window")
	}
	defer glfw.Terminate()

	program, err := initGl()
	if err != nil {
		exitWith(err, "unable to create OpenGL program")
	}

	cells := makeCells()

	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	colorshiftUniform := gl.GetUniformLocation(program, gl.Str("colorshift\x00"))
	timingUniform := gl.GetUniformLocation(program, gl.Str("timing\x00"))

	projection := mgl32.Ortho(0, width, 0, height, 0.1, 500)
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.LookAtV(mgl32.Vec3{0, 0, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	model := mgl32.Ident4()
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	gl.Uniform3f(colorshiftUniform, 1, 1, 1)

	fpsLimiter := games.NewFpsLimiter(fps)

	for !window.ShouldClose() {
		fpsLimiter.StartFrame()

		for _, cell := range cells {
			cell.checkState(cells)
		}

		gl.Uniform1f(timingUniform, float32(glfw.GetTime()))
		gl.Uniform3f(colorshiftUniform, float32(math.Sin(glfw.GetTime())), float32(math.Cos(glfw.GetTime())), float32(math.Sin(glfw.GetTime()))*float32(math.Cos(glfw.GetTime())))

		if err := draw(cells, window, program, projectionUniform, cameraUniform, modelUniform); err != nil {
			exitWith(err, "window draw failure")
		}

		fpsLimiter.WaitForNextFrame()
		fmt.Println("fps", fpsLimiter.CurrentFrameFps())
	}
}

func makeCells() []*cell {
	rand.Seed(100)

	drawable := makeVao(square)

	cells := make([]*cell, rows*columns, rows*columns)
	for x := 0; x < columns; x++ {
		for y := 0; y < rows; y++ {
			c := newCell(x, y, drawable)

			c.alive = rand.Float64() < threshold
			c.aliveNext = c.alive

			cells[x+y*columns] = c
		}
	}
	return cells
}

func newCell(x, y int, drawable uint32) *cell {
	return &cell{
		drawable: drawable,

		x: x,
		y: y,
	}
}

func initGlfw() (*glfw.Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, err
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return nil, err
	}
	window.MakeContextCurrent()

	return window, nil
}

func initGl() (uint32, error) {
	if err := gl.Init(); err != nil {
		return 0, err
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.UseProgram(program)
	gl.ClearColor(0, 0, 0, 1)

	return program, nil
}

func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao

}

func draw(cells []*cell, window *glfw.Window, program uint32, projectionUniform int32, cameraUniform int32, modelUniform int32) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	for _, cell := range cells {
		cell.draw(modelUniform)
	}

	glfw.PollEvents()
	window.SwapBuffers()

	return nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func exitWith(err error, msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	msg = fmt.Sprintf("%s: %v", msg, err)
	panic(msg)
}
