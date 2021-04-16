package main

import (
	"log"
	"runtime"

	"github.com/kaitsubaka/glutils/gfx"
	"github.com/kaitsubaka/glutils/win"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width  = 1080
	height = 720
	title  = "Simple Light"
)

var (
	points = []float32{
		-0.5, 0.5, 1.0, 0.0, 0.0, // top-left
		0.5, 0.5, 0.0, 1.0, 0.0, // top-right
		0.5, -0.5, 0.0, 0.0, 1.0, // bottom-right
		-0.5, -0.5, 1.0, 1.0, 0., // bottom-left
	}
)

func createVAO(points []float32) uint32 {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.BindVertexArray(VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(points)*4, gl.Ptr(points), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.BindVertexArray(0)

	return VAO
}

func programLoop(window *win.Window) error {

	// Shaders and textures
	vertShader, err := gfx.NewShaderFromFile("shaders/geometry.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fragShader, err := gfx.NewShaderFromFile("shaders/geometry.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}
	geomShader, err := gfx.NewShaderFromFile("shaders/geometry.geom", gl.GEOMETRY_SHADER)
	if err != nil {
		return err
	}

	program, err := gfx.NewProgram(vertShader, fragShader, geomShader)
	if err != nil {
		return err
	}
	defer program.Delete()

	// Ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Base model
	model := mgl32.Ident4()

	// Uniform
	modelUniformLocation := program.GetUniformLocation("model")
	viewUniformLocation := program.GetUniformLocation("view")
	projectUniformLocation := program.GetUniformLocation("projection")

	// creates camara
	eye := mgl32.Vec3{0, 0, 1}
	center := mgl32.Vec3{0, 0, 0}
	camera := mgl32.LookAtV(eye, center, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(viewUniformLocation, 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// Colors
	backgroundColor := mgl32.Vec3{0, 0, 0}

	// Uncomment to turn on polygon mode
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Geometry
	VAO := createVAO(points)

	// Scene and animation always needs to be after the model and buffers initialization
	animationCtl := gfx.NewAnimationManager()

	animationCtl.Init() // always needs to be before the main loop in order to get correct times
	// main loop
	for !window.ShouldClose() {
		window.StartFrame()

		// background color
		gl.ClearColor(backgroundColor.X(), backgroundColor.Y(), backgroundColor.Z(), 1.)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Scene update
		animationCtl.Update()

		// You shall draw here
		program.Use()
		gl.UniformMatrix4fv(modelUniformLocation, 1, false, &model[0])
		gl.BindVertexArray(VAO)
		gl.DrawArrays(gl.POINTS, 0, 4)

		gl.BindVertexArray(0)
	}

	return nil
}

func main() {

	runtime.LockOSThread()
	win.InitGlfw(4, 0)
	defer glfw.Terminate()
	window := win.NewWindow(width, height, title)
	gfx.InitGl()

	err := programLoop(window)
	if err != nil {
		log.Fatal(err)
	}
}
