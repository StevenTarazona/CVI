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
	wellPosition     = mgl32.Vec3{1, 0, 1}
	wellViewPosition = mgl32.Vec3{1, 1, 1}

	camPositions = [][2]mgl32.Vec3{
		{{-4.5, 1, 4.5}, {-4, 0, -4}},
		{{-1, 1, 3}, wellViewPosition},
		{{4.5, 2, 4}, {1, 0, 2}},
		{{3, 1, 3}, {1, 1, 1.5}},
	}

	treePositions = []mgl32.Vec3{
		{-2.5, 0, -0.5},
		{-2, 0, 1.5},
		{-2, 0, -3},
		{-1, 0, -1},
		{-1, 0, 1},

		{-4.5, 0, -0.5},
		{-4, 0, 1.5},
		{-4, 0, -3},
		{-3, 0, -1},
		{-3, 0, 1},

		{2.5, 0, -0.5},
		{2, 0, -1.5},
		{2, 0, -3},
		{1, 0, -2},
		{-1, 0, -2},
		{0, 0, -2.5},

		{4.5, 0, -0.5},
		{4, 0, 1.5},
		{4, 0, -3},
		{3, 0, -1},
		{3, 0, 1},
	}

	camPathPoints = [][]mgl32.Vec3{
		{{-4, 0, -4}, {4, 0, -4}, {4, 0, 3}, {4, 0, 4}},
		{{4, 0, 4}, {4, 0, -4}, {-4, 0, -3}, {-4, 0, -4}},
		{{-4, 0, -4}, {4, 0, -4}, wellPosition, wellViewPosition},
	}
	personPathPoints = []mgl32.Vec3{
		{-5, 0, 5},
		{1, 0, 5},
		{1, 0, 2.5},
	}
	coinPathPoints = []mgl32.Vec3{
		{0, 0, 0},
		{0, 1.5, -0.7},
		{0, -1.6, -1},
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
	camera := mgl32.LookAtV(mgl32.Vec3{5, 1, 4}, mgl32.Vec3{1, 1, 1.7}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// creates light
	gl.Uniform3f(lightColorUniformLocation, 1, 1, 1)

	// Uncomment to turn on polygon mode
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Get primitive vertices and create VAOs

	cubeVertices := ge.GetCubicHexahedronVertices3(1, 1, 1)
	cubeVAO := ge.CreateVAO(cubeVertices)

	sphereVertices, sphereVerticesT, sphereVerticesB := ge.GetSphereVertices3(1, 32)
	sphereVAO, sphereVAOT, sphereVAOB := ge.CreateVAO(sphereVertices), ge.CreateVAO(sphereVerticesT), ge.CreateVAO(sphereVerticesB)

	cylinderVertices, cylinderVerticesT, cylinderVerticesB := ge.GetCylinderVertices3(1, 0.1, 0.1, 5)
	cylinderVAO, cylinderVAOT, cylinderVAOB := ge.CreateVAO(cylinderVertices), ge.CreateVAO(cylinderVerticesT), ge.CreateVAO(cylinderVerticesB)

	PipeVerticesSI, PipeVerticesSO, PipeVerticesT, PipeVerticesSB := ge.GetPipeVertices3(0.75, 0.4, 0.5, 16)
	PipeVAOSI, PipeVAOSO, PipeVAOT, PipeVAOB := ge.CreateVAO(PipeVerticesSI), ge.CreateVAO(PipeVerticesSO), ge.CreateVAO(PipeVerticesT), ge.CreateVAO(PipeVerticesSB)

	personVertices, personVerticesT, personVerticesB := ge.GetCapsuleVertices3(0.7, 0.2, 0.1, 8)
	personVAO, personVAOT, personVAOB := ge.CreateVAO(personVertices), ge.CreateVAO(personVerticesT), ge.CreateVAO(personVerticesB)

	extremityVertices, extremityVerticesT, extremityVerticesB := ge.GetCapsuleVertices3(0.4, 0.07, 0.05, 8)
	extremityVAO, extremityVAOT, extremityVAOB := ge.CreateVAO(extremityVertices), ge.CreateVAO(extremityVerticesT), ge.CreateVAO(extremityVerticesB)

	headVertices, headVerticesT, headVerticesB := ge.GetSphereVertices3(0.2, 18)
	headVAO, headVAOT, headVAOB := ge.CreateVAO(headVertices), ge.CreateVAO(headVerticesT), ge.CreateVAO(headVerticesB)

	coinVertices, coinVerticesT, coinVerticesB := ge.GetCylinderVertices3(0.02, 0.05, 0.05, 8)
	coinVertices, coinVerticesT, coinVerticesB = ge.Translate(coinVertices, mgl32.Vec3{0, -0.01, 0}), ge.Translate(coinVerticesT, mgl32.Vec3{0, -0.01, 0}), ge.Translate(coinVerticesB, mgl32.Vec3{0, -0.01, 0})
	coinVAO, coinVAOT, coinVAOB := ge.CreateVAO(coinVertices), ge.CreateVAO(coinVerticesT), ge.CreateVAO(coinVerticesB)

	planeVertices := ge.GetPlaneVertices3(12, 12, 1)
	planeVAO := ge.CreateVAO(planeVertices)

	// Animation models that will get updated
	personPathModel := mgl32.Translate3D(personPathPoints[0].Elem()).Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(-45)))
	armModelL := mgl32.HomogRotate3DX(mgl32.DegToRad(45)).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(135)))
	armModelR := mgl32.HomogRotate3DX(mgl32.DegToRad(45)).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(-135)))
	coinPathModel := model

	// Scene and animation
	angle := 0.0
	previousTime := glfw.GetTime()
	totalElapsed := float64(0)
	movementControlCount := 0

	movementTimes := []float64{}
	movementFunctions := []func(t float32){}

	// Camera movement
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], camPositions[0][1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])
	}), append(movementTimes, .5)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], mgl32.BezierCurve3D(t, camPathPoints[0]), mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])
	}), append(movementTimes, 1.0)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], mgl32.BezierCurve3D(t, camPathPoints[1]), mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])
	}), append(movementTimes, 1.0)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], mgl32.BezierCurve3D(t, camPathPoints[2]), mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])
	}), append(movementTimes, .5)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], camPathPoints[2][len(camPathPoints[2])-1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])
	}), append(movementTimes, .5)
	// Camara well framing
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[1][0], camPositions[1][1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])
	}), append(movementTimes, 1)
	// Camara zoom-out framing
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[2][0], camPositions[2][1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])
	}), append(movementTimes, 0.1)
	// Person movement to the well
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		// Number of waves
		waves := float32(3)
		rot := math32.Sin(2*math32.Pi*waves*t) * mgl32.DegToRad(10)
		point := mgl32.BezierCurve3D(t, personPathPoints)
		roty := math32.Atan((point.X()+5)/(point.Y()+5)) - math32.Pi/4
		personPathModel = mgl32.Translate3D(point.Elem()).Mul4(mgl32.HomogRotate3DZ(rot)).Mul4(mgl32.HomogRotate3DY(roty))
	}), append(movementTimes, 2.0)
	// Camara well framing
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[3][0], camPositions[3][1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])
	}), append(movementTimes, .5)
	// Movement of the person's arm
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		armModelL = mgl32.HomogRotate3DX(mgl32.DegToRad(45 - 90*t)).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(135 + 45*t)))
	}), append(movementTimes, .5)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		armModelL = mgl32.HomogRotate3DX(mgl32.DegToRad(155*t - 225))
	}), append(movementTimes, .5)
	// Movement of the coin
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		// Number of flips
		flips := float32(3)
		rot := 2 * math32.Pi * flips * t
		coinPathModel = mgl32.Translate3D(mgl32.BezierCurve3D(t, coinPathPoints).Elem()).Mul4(mgl32.HomogRotate3DX(float32(rot)))
	}), append(movementTimes, 1.0)
	// Delete coin
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		coinPathModel = mgl32.Mat4{}
	}), append(movementTimes, .5)
	// Movement of the person's arm
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		armModelL = mgl32.HomogRotate3DX(mgl32.DegToRad(25*t - 70)).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(45 * t)))
		armModelR = mgl32.HomogRotate3DX(mgl32.DegToRad(45 + 90*t)).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(-135)))
	}), append(movementTimes, .3)

	// Main loop
	for !window.ShouldClose() {

		time := glfw.GetTime()
		elapsed := time - previousTime
		totalElapsed += elapsed
		previousTime = time
		angle += elapsed

		ge.NewFrame(window, mgl32.Vec4{0, 0.27, 0.7, 1.0})

		// Scene update
		if movementControlCount < len(movementFunctions) {
			if animationTime := movementTimes[movementControlCount]; totalElapsed <= animationTime {
				t := float32(totalElapsed) / float32(animationTime)
				movementFunctions[movementControlCount](t)
			} else {
				movementFunctions[movementControlCount](1)
				totalElapsed = 0
				movementControlCount++
			}
		}

		// You shall draw here

		// Moon
		moonTranslate := mgl32.Translate3D(-5, 3, 2)
		gl.Uniform3f(colorUniformLocation, 0.850, 0.850, 0.850)
		gl.BindVertexArray(sphereVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &moonTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereVertices)))

		gl.BindVertexArray(sphereVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &moonTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereVerticesT)))

		gl.BindVertexArray(sphereVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &moonTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereVerticesB)))

		// Well
		wellTranslate := mgl32.Translate3D(wellPosition.Elem())

		gl.Uniform3f(colorUniformLocation, 0.301, 0.301, 0.301)
		gl.BindVertexArray(PipeVAOSI)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(PipeVerticesSI)))

		gl.BindVertexArray(PipeVAOSO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(PipeVerticesSO)))

		gl.BindVertexArray(PipeVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(PipeVerticesT)))

		gl.BindVertexArray(PipeVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(PipeVerticesSB)))

		// Pillars
		gl.Uniform3f(colorUniformLocation, 0.301, 0.149, 0)
		wellTransform := wellTranslate.Mul4(mgl32.Translate3D(-0.55, 0, 0)).Mul4(mgl32.Scale3D(0.75, 2, 0.75))

		gl.BindVertexArray(cylinderVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cylinderVertices)))

		gl.BindVertexArray(cylinderVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesT)))

		gl.BindVertexArray(cylinderVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesB)))

		wellTransform = wellTranslate.Mul4(mgl32.Translate3D(0.55, 0, 0)).Mul4(mgl32.Scale3D(0.75, 2, 0.75))

		gl.BindVertexArray(cylinderVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cylinderVertices)))

		gl.BindVertexArray(cylinderVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesT)))

		gl.BindVertexArray(cylinderVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesB)))

		wellTransform = wellTranslate.Mul4(mgl32.Translate3D(0.55, 1.75, -0.15)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(90))).Mul4(mgl32.Scale3D(0.5, 0.3, 0.5))

		gl.BindVertexArray(cylinderVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cylinderVertices)))

		gl.BindVertexArray(cylinderVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesT)))

		gl.BindVertexArray(cylinderVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesB)))

		wellTransform = wellTransform.Mul4(mgl32.Translate3D(-2.2, 0, 0))

		gl.BindVertexArray(cylinderVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cylinderVertices)))

		gl.BindVertexArray(cylinderVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesT)))

		gl.BindVertexArray(cylinderVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesB)))

		// Roof
		gl.Uniform3f(colorUniformLocation, 0.623, 0.141, 0.078)
		wellTransform = wellTranslate.Mul4(mgl32.Translate3D(0, 1.75, 0.2)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45))).Mul4(mgl32.Scale3D(1.25, 0.1, 0.75))

		gl.BindVertexArray(cubeVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

		wellTransform = wellTranslate.Mul4(mgl32.Translate3D(0, 1.75, -0.2)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(-45))).Mul4(mgl32.Scale3D(1.25, 0.1, 0.75))

		gl.BindVertexArray(cubeVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

		// Person
		personTranslate := personPathModel

		// Body
		gl.Uniform3f(colorUniformLocation, 0.24, 0.24, 0.24)
		bodyTranslate := personTranslate.Mul4(mgl32.Translate3D(0, 0.3, 0))

		gl.BindVertexArray(personVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &bodyTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(personVertices)))

		gl.BindVertexArray(personVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &bodyTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(personVerticesT)))

		gl.BindVertexArray(personVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &bodyTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(personVerticesB)))

		// Legs
		gl.Uniform3f(colorUniformLocation, 0.13, 0.13, 0.13)
		legTranslate := bodyTranslate.Mul4(mgl32.Translate3D(0.1, 0.1, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(180)))

		gl.BindVertexArray(extremityVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremityVertices)))

		gl.BindVertexArray(extremityVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityVerticesT)))

		gl.BindVertexArray(extremityVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityVerticesB)))

		legTranslate = bodyTranslate.Mul4(mgl32.Translate3D(-0.1, 0.1, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(180)))

		gl.BindVertexArray(extremityVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremityVertices)))

		gl.BindVertexArray(extremityVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityVerticesT)))

		gl.BindVertexArray(extremityVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityVerticesB)))

		// Arms
		gl.Uniform3f(colorUniformLocation, 0.84, 0.8, 0.66)
		armsTranslate := bodyTranslate.Mul4(mgl32.Translate3D(0.1, 0.6, 0)).Mul4(armModelR)

		gl.BindVertexArray(extremityVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremityVertices)))

		gl.BindVertexArray(extremityVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityVerticesT)))

		gl.BindVertexArray(extremityVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityVerticesB)))

		armsTranslate = bodyTranslate.Mul4(mgl32.Translate3D(-0.1, 0.6, 0)).Mul4(armModelL)

		gl.BindVertexArray(extremityVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremityVertices)))

		gl.BindVertexArray(extremityVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityVerticesT)))

		gl.BindVertexArray(extremityVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityVerticesB)))

		// Head
		headTranslate := bodyTranslate.Mul4(mgl32.Translate3D(0, 0.65, 0))
		gl.BindVertexArray(headVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &headTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(headVertices)))

		gl.BindVertexArray(headVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &headTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(headVerticesT)))

		gl.BindVertexArray(headVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &headTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(headVerticesB)))

		// Coin
		gl.Uniform3f(colorUniformLocation, 1, 1, 0)
		coinTranslate := armsTranslate.Mul4(mgl32.Translate3D(0, 0.4, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(90))).Mul4(coinPathModel)

		gl.BindVertexArray(coinVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &coinTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(coinVertices)))

		gl.BindVertexArray(coinVAOT)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &coinTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(coinVerticesT)))

		gl.BindVertexArray(coinVAOB)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &coinTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(coinVerticesB)))

		// Trees
		for _, pos := range treePositions {
			// Tree
			treeTranslate := mgl32.Translate3D(pos.Elem())

			// Trunk
			gl.Uniform3f(colorUniformLocation, 0.4, 0.2, 0)

			gl.BindVertexArray(cylinderVAO)
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cylinderVertices)))

			gl.BindVertexArray(cylinderVAOT)
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesT)))

			gl.BindVertexArray(cylinderVAOB)
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(cylinderVerticesB)))

			// Leaves
			scale1 := 1 - math32.Abs(math32.Sin(float32(time)))*0.04
			scale2 := 1 - math32.Abs(math32.Cos(float32(time)))*0.04

			gl.BindVertexArray(cubeVAO)

			gl.Uniform3f(colorUniformLocation, 0, 0.9, 0)
			worldTranslate := treeTranslate.Mul4(mgl32.Scale3D(1, 1*scale1, 1)).Mul4(mgl32.Translate3D(0, 1.25, -0.5))
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &worldTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

			gl.Uniform3f(colorUniformLocation, 0, 0.75, 0)
			worldTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.75, 0.75*scale1, 0.75)).Mul4(mgl32.Translate3D(0, 1.2, 0.4))
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &worldTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

			gl.Uniform3f(colorUniformLocation, 0, 0.5, 0)
			worldTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.8, 0.8*scale2, 0.8)).Mul4(mgl32.Translate3D(-0.5, 1.5, 0))
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &worldTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
		}

		gl.Uniform3f(colorUniformLocation, 0.4, 0.6, 0)
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
