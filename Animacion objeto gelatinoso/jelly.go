package main

import (
	"log"
	"runtime"

	"git.maze.io/go/math32"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	ge "github.com/kaitsubaka/glae" // libreria propia implementada desde 0, https://github.com/kaitsubaka/glae
)

const (
	width              = 1080
	height             = 720
	windowName         = "A Wish"
	vertexShaderSource = `
		#version 410 core
		layout (location = 0) in vec3 position;
		out vec2 TexCoord;
		uniform mat4 world;
		uniform mat4 camera;
		uniform mat4 project;
		void main()
		{
			gl_Position = project * camera * world * vec4(position, 1.0);
		}`

	fragmentShaderSource = `
		#version 410 core
		out vec4 color;
		uniform vec3 objectColor;
		uniform vec3 lightColor;
		void main()
		{
			color = vec4(objectColor * lightColor, 1.0f);
		}`
)

var (
	pathPoints = []mgl32.Vec3{
		{5, 5, -5},
		{0, 5, 0},
		{0, 0, 0},
	}
	pathPoints2 = []mgl32.Vec3{
		{0, 0, 0},
		{-4, 0, 4},
	}
)

func programLoop(window *glfw.Window) error {

	// Compile shaders and link to program
	vertShader, err := ge.CompileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fragShader, err := ge.CompileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}
	var shaders = []uint32{vertShader, fragShader}
	program, err := ge.CreateProgram(shaders)
	if err != nil {
		return nil
	}

	// Ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.UseProgram(program)

	// Base model
	model := mgl32.Ident4()

	// Uniform locations
	WorldUniformLocation := gl.GetUniformLocation(program, gl.Str("world\x00"))
	colorUniformLocation := gl.GetUniformLocation(program, gl.Str("objectColor\x00"))
	lightColorUniformLocation := gl.GetUniformLocation(program, gl.Str("lightColor\x00"))
	cameraUniformLocation := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	projectUniformLocation := gl.GetUniformLocation(program, gl.Str("project\x00"))

	// creates camara
	camera := mgl32.LookAtV(mgl32.Vec3{0, 4, 10}, mgl32.Vec3{0, 1, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// creates light
	gl.Uniform3f(lightColorUniformLocation, 1, 1, 1)

	// Uncomment to turn on polygon mode
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Get primitive vertices and create VAOs
	capsuleVertices, capsuleVerticesT, capsuleVerticesB := ge.GetCapsuleVertices3(2, .6, .3, 18)
	capsuleVAO, capsuleVAOT, capsuleVAOB := ge.CreateVAO(capsuleVertices), ge.CreateVAO(capsuleVerticesT), ge.CreateVAO(capsuleVerticesB)

	planeVertices := ge.GetPlaneVertices3(12, 12, 1)
	planeVAO := ge.CreateVAO(planeVertices)

	// Animation models that will get updated
	pathModel := model

	// Scene and animation
	angle := 0.0
	previousTime := glfw.GetTime()
	totalElapsed := float64(0)
	movementControlCount := 0

	movementTimes := []float64{}
	movementFunctions := []func(t float32){}

	// Capsule fall
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		pathModel = mgl32.Translate3D(mgl32.BezierCurve3D(math32.Pow(t, 3), pathPoints).Elem())
	}), append(movementTimes, 1)
	// Capsule deform and jump
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		waves := float32(10)
		amplitude := float32(.8)
		scale := math32.Sin(2*math32.Pi*waves*math32.Pow(t, 1.3)) * (1 - t) * amplitude
		jump := -scale * 2
		if jump < 0 {
			jump = 0
		}
		pos := math32.Pow(t*1.6, 1.0/2)
		if pos > 1 {
			pos = 1
		}
		pathModel = mgl32.Translate3D(pathPoints[len(pathPoints)-1].Elem()).
			Mul4(mgl32.Translate3D(0, jump, 0)).
			Mul4(mgl32.Translate3D(mgl32.BezierCurve3D(pos, pathPoints2).Elem())).
			Mul4(mgl32.Scale3D(1+scale, 1-scale, 1+scale))
	}), append(movementTimes, 3.5)

	// Main loop
	for !window.ShouldClose() {

		time := glfw.GetTime()
		elapsed := time - previousTime
		totalElapsed += elapsed
		previousTime = time
		angle += elapsed

		ge.NewFrame(window, mgl32.Vec4{0, 0, 0, 0})

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

		gl.Uniform3f(colorUniformLocation, 1, 0, 1)
		//model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0}).Mul4(pathModel)
		capsuleModel := pathModel

		gl.BindVertexArray(capsuleVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &capsuleModel[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(capsuleVertices)))

		gl.BindVertexArray(capsuleVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &capsuleModel[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(capsuleVerticesT)))

		gl.BindVertexArray(capsuleVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &capsuleModel[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(capsuleVerticesB)))

		gl.Uniform3f(colorUniformLocation, 1, 1, 1)
		gl.BindVertexArray(planeVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &model[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(planeVertices)))
		gl.BindVertexArray(0)
	}

	return nil
}

func main() {
	runtime.LockOSThread()
	window := ge.InitGlfw(width, height, 4, 0, windowName)
	defer glfw.Terminate()

	ge.InitGl()

	err := programLoop(window)
	if err != nil {
		log.Fatalln(err)
	}
}
