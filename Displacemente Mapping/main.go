package main

import (
	"image"
	"log"
	"os"
	"runtime"
	"unsafe"

	"github.com/StevenTarazona/glcore/ge"
	"github.com/StevenTarazona/glcore/gfx"
	"github.com/StevenTarazona/glcore/win"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width  = 1080
	height = 720
	title  = "Core"
)

func programLoop(window *win.Window) error {

	// Shaders and textures
	vertShader, err := gfx.NewShaderFromFile("shaders/basic.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := gfx.NewShaderFromFile("shaders/basic.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}
	defer program.Delete()

	// Ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	program.Use()

	// Base model
	model := mgl32.Ident4()

	// Uniform locations
	WorldUniformLocation := program.GetUniformLocation("world")
	colorUniformLocation := program.GetUniformLocation("objectColor")
	lightColorUniformLocation := program.GetUniformLocation("lightColor")
	cameraUniformLocation := program.GetUniformLocation("camera")
	projectUniformLocation := program.GetUniformLocation("project")
	textureUniformLocation := program.GetUniformLocation("texture")

	// creates camara
	//camera := mgl32.LookAtV(mgl32.Vec3{10, 5, -1}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	camera := mgl32.LookAtV(mgl32.Vec3{10, 2, 0}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// creates light
	gl.Uniform3f(lightColorUniformLocation, 1, 1, 1)

	// Uncomment to turn on polygon mode
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Scene and animation
	angle := 0.0
	previousTime := glfw.GetTime()
	totalElapsed := float64(0)
	movementControlCount := 0

	movementTimes := []float64{}
	movementFunctions := []func(t float32){}

	// Textures
	iceTexture, err := gfx.NewTextureFromFile("images/snow3.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// Get primitive vertices and create VAOs
	imgFile, err := os.Open("images/noise2.jpg")
	if err != nil {
		panic(err.Error())
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		panic(err.Error())
	}

	squareVertices, squareTCoords, squareIndices := ge.GetSquareStripDisplaced(200, 50, .1, img, 5)
	squareVAO := ge.CreateVAO(squareVertices, squareTCoords, squareIndices)
	count := int32(len(squareIndices))
	squareVertices, squareTCoords, squareIndices = nil, nil, nil

	for !window.ShouldClose() {
		window.StartFrame()

		// background color
		gl.ClearColor(0.078, 0.094, 0.321, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		time := glfw.GetTime()
		elapsed := time - previousTime
		totalElapsed += elapsed
		previousTime = time
		angle += elapsed

		// Scene update
		if movementControlCount < len(movementFunctions) {
			if animationTime := movementTimes[movementControlCount]; totalElapsed <= animationTime {
				t := float32(totalElapsed / animationTime)
				movementFunctions[movementControlCount](t)
			} else {
				movementFunctions[movementControlCount](1)
				totalElapsed = 0
				movementControlCount++
			}
		}

		// You shall draw here

		gl.Uniform3f(colorUniformLocation, 1, 1, 1)
		iceTexture.Bind(gl.TEXTURE0)
		iceTexture.SetUniform(textureUniformLocation)

		gl.BindVertexArray(squareVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &model[0])
		gl.DrawElements(gl.TRIANGLE_STRIP, count, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		//gl.DrawArrays(gl.TRIANGLE_STRIP, 0, count)

		iceTexture.UnBind()

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
