package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"git.maze.io/go/math32"
	"github.com/go-gl/gl/v4.1-core/gl" // OR: github.com/go-gl/gl/v2.1/gl
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width              = 720
	height             = 480
	vertexShaderSource = `
		#version 400
		uniform mat4 projection;
		uniform mat4 camera;
		uniform mat4 model;
		in vec3 vp;
		void main() {
			gl_Position = projection * camera * model * vec4(vp, 1.0);
		}
	` + "\x00"

	fragmentShaderSource = `
		#version 400
		out vec3 frag_colour;
		void main() {
			frag_colour = vec3(0.5, 0, 0.5);
		}
	` + "\x00"
)

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 0)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Test", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}

// https://github.com/go-gl/examples/blob/master/gl41core-cube/cube.go
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

// makeVao initializes and returns a vertex array from the points provided.
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

// generate a unit circle on XY-plane
func getCircleVertices(x float32, y float32, z float32, r float32, vertices int) []float32 {
	var sectorStep = 2 * math32.Pi / float32(vertices)
	var sectorAngle float32 // radian
	var unitCircleVertices = []float32{x, z, y}
	for i := 0; i <= vertices; i++ {
		sectorAngle = float32(i) * sectorStep
		unitCircleVertices = append(unitCircleVertices, x+r*math32.Cos(sectorAngle)) // x
		unitCircleVertices = append(unitCircleVertices, z)                           // z
		unitCircleVertices = append(unitCircleVertices, y+r*math32.Sin(sectorAngle)) // y
	}
	return unitCircleVertices
}

func drawCircle(vertex []float32) {
	makeVao(vertex)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(vertex)/3))
}

func drawPanel(vertex []float32) {
	makeVao(vertex)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(vertex)/3))
}

func drawCylinder(h float32, r float32, slices int, vertices int) {
	var top = getCircleVertices(0, 0, h/2, 1, vertices)
	var bottom = getCircleVertices(0, 0, -h/2, 1, vertices)
	for slice := 0; slice < slices; slice++ {
		var side []float32
		for i := 3; i < len(bottom); i += 3 {
			side = append(side, bottom[i])
			side = append(side, bottom[i+1]+float32(slice)*(h/float32(slices)))
			side = append(side, bottom[i+2])
			side = append(side, bottom[i])
			side = append(side, bottom[i+1]+float32(slice+1)*(h/float32(slices)))
			side = append(side, bottom[i+2])
		}
		drawPanel(side)
	}
	drawCircle(top)
	drawCircle(bottom)
}

func main() {
	runtime.LockOSThread()
	window := initGlfw()
	defer glfw.Terminate()
	program := initOpenGL()
	gl.UseProgram(program)

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(width)/height, 0.1, 10.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.LookAtV(mgl32.Vec3{4, 4, 4}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0, 0, 0, 1.0)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	angle := 0.0
	previousTime := glfw.GetTime()

	for !window.ShouldClose() {
		// clear bg
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// main drawing
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time
		angle += elapsed
		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		// Draw circle
		drawCylinder(4, 1, 6, 16)

		// update events
		glfw.PollEvents()
		window.SwapBuffers()

	}
}
