package main

import (
	"log"
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

var (
	//r := []int{0, 9, 13, 14, 16, 17, 18}
	r   = []int{0, 9, 14, 17, 18}
	rr  = []int{3, 6, 10, 11, 15, 21, 24, 27, 30, 33, 36, 39, 42, 43, 44, 45, 48, 49, 50, 51, 54}
	rb  = []int{1, 2, 22, 23, 25, 26, 40}
	rt  = []int{4, 5, 7, 8, 19, 20, 31}
	rtb = []int{12, 28, 29, 32, 34, 35, 37, 38, 41, 46, 47, 52, 53, 55, 56}

	//b := []int{0, 1, 2, 13, 16, 22, 25}
	b   = []int{0, 1, 2, 22, 25}
	bb  = []int{3, 5, 6, 7, 8, 10, 19, 27, 28, 29, 30, 31, 32, 33, 34, 35, 43, 49, 50, 52, 54, 55, 56}
	bl  = []int{12, 15, 20, 21, 24}
	br  = []int{14, 17, 18, 23, 26, 40}
	blr = []int{4, 36, 47, 38, 39, 42, 44, 45, 46, 47, 48, 51, 53}

	adjacencyList = [][][]int{
		// 0-8
		{rb, br},
		{rb, bb},
		{r, bl},
		{rt, bl},
		{rt, b},
		{rr, br},
		{rt, bl},
		{rt, b},
		{rr, br},

		// 9-17
		{rr, br},
		{rr, bb},
		{r, bl},
		{r, bl},
		{r, b},
		{rtb, br},
		{r, bl},
		{r, b},
		{rr, br},

		// 18-26
		{rt, b},
		{rt, b},
		{r, b},
		{rb, bb},
		{rb, blr},
		{rr, bb},
		{rb, bb},
		{rb, bb},
		{rr, bb},

		//27-35
		{rtb, blr},
		{rtb, blr},
		{rr, blr},
		{rr, blr},
		{rtb, br},
		{rt, bl},
		{rtb, blr},
		{rtb, bb},
		{rr, blr},

		// 36--44
		{rtb, blr},
		{rtb, blr},
		{rr, blr},
		{rr, blr},
		{rtb, bb},
		{rb, bb},
		{rr, blr},
		{rr, bb},
		{rr, blr},

		// 45-53
		{rtb, bb},
		{rtb, bb},
		{rr, bb},
		{rr, bb},
		{rr, bb},
		{rr, bb},
		{rtb, bb},
		{rtb, bb},
		{rr, bb},

		// 54-56
		{rtb, bb},
		{rtb, bb},
		{rr, bb},
	}
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
	camera := mgl32.LookAtV(mgl32.Vec3{0, 7, 7}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// creates light
	gl.Uniform3f(lightColorUniformLocation, 1, 1, 1)

	// Uncomment to turn on polygon mode
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Scene and animation
	angle := 0.0
	previousTime := glfw.GetTime()
	totalElapsed := float64(0)
	movementControlCount := 0

	movementTimes := []float64{}
	movementFunctions := []func(t float32){}

	// Textures
	grassTexture, err := gfx.NewTextureFromFile("images/farm.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// Get primitive vertices and create VAOs

	tileCords := [][]mgl32.Vec2{}
	first := mgl32.Vec2{64, 101}
	firstX := first.X()
	firstY := first.Y()
	for i := 0; i < 6; i++ {
		for j := 0; j < 9; j++ {
			tileCord := []mgl32.Vec2{first, first.Add(mgl32.Vec2{0, 97}), first.Add(mgl32.Vec2{97, 0}), first.Add(mgl32.Vec2{97, 97})}
			tileCords = append(tileCords, ge.Transform2(tileCord, mgl32.Vec2{1 / float32(grassTexture.Width), 1 / float32(grassTexture.Height)}))
			first = first.Add(mgl32.Vec2{110, 0})
		}
		firstY += 110
		first = mgl32.Vec2{firstX, firstY}
	}

	first = mgl32.Vec2{394, 761}
	for i := 0; i < 3; i++ {
		tileCord := []mgl32.Vec2{first, first.Add(mgl32.Vec2{0, 97}), first.Add(mgl32.Vec2{97, 0}), first.Add(mgl32.Vec2{97, 97})}
		tileCords = append(tileCords, ge.Transform2(tileCord, mgl32.Vec2{1 / float32(grassTexture.Width), 1 / float32(grassTexture.Height)}))
		first = first.Add(mgl32.Vec2{110, 0})
	}

	squareVertices, squareTCoords, squareIndices := ge.GetSquareWangTiles(40, 40, 1, tileCords, adjacencyList)
	squareVAO := ge.CreateVAO(squareVertices, squareTCoords, squareIndices)

	for !window.ShouldClose() {
		window.StartFrame()

		// background color
		gl.ClearColor(0, 0.27, 0.7, 1.0)
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
		grassTexture.Bind(gl.TEXTURE0)
		grassTexture.SetUniform(textureUniformLocation)

		gl.BindVertexArray(squareVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &model[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(squareIndices)), gl.UNSIGNED_INT, unsafe.Pointer(nil))

		grassTexture.UnBind()

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
