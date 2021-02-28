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
	windowName         = "Core"
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

func programLoop(window *glfw.Window) error {

	// the linked shader program determines how the data will be rendered
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

	// ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.UseProgram(program)

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("world\x00"))
	colorModel := gl.GetUniformLocation(program, gl.Str("objectColor\x00"))
	cameraModel := gl.GetUniformLocation(program, gl.Str("camera\x00"))

	// creates camara
	camera := mgl32.LookAtV(mgl32.Vec3{5, 1, 4}, mgl32.Vec3{1, 1, 1.7}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("project\x00")), 1, false, &projectTransform[0])

	// light
	gl.Uniform3f(gl.GetUniformLocation(program, gl.Str("lightColor\x00")), 1, 1, 1)

	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	cubeVertices := ge.GetCubicHexahedronVertices3(1, 1, 1)
	cubeVAO := ge.CreateVAO(cubeVertices)

	sphereVertices, sphereTopVertices, sphereBottomVertices := ge.GetSphereVertices3(1, 16)
	sphereVAO, sphereTopVAO, sphereBottomVAO := ge.CreateVAO(sphereVertices), ge.CreateVAO(sphereTopVertices), ge.CreateVAO(sphereBottomVertices)

	sideVertices, topVertices, bottomVertices := ge.GetCylinderVertices3(1, 0.1, 0.1, 5)
	sideVAO, topVAO, bottomVAO := ge.CreateVAO(sideVertices), ge.CreateVAO(topVertices), ge.CreateVAO(bottomVertices)

	sideInVerticesPipe, sideOutVerticesPipe, topVerticesPipe, bottomVerticesPipe := ge.GetPipeVertices3(0.75, 0.4, 0.5, 16)
	sideInVAOPipe, sideOutVAOPipe, topVAOPipe, bottomVAOPipe := ge.CreateVAO(sideInVerticesPipe), ge.CreateVAO(sideOutVerticesPipe), ge.CreateVAO(topVerticesPipe), ge.CreateVAO(bottomVerticesPipe)

	personSideVertices, personTopVertices, personBottomVertices := ge.GetCapsuleVertices3(0.7, 0.2, 0.1, 8)
	personSideVAO, personTopVAO, personBottomVAO := ge.CreateVAO(personSideVertices), ge.CreateVAO(personTopVertices), ge.CreateVAO(personBottomVertices)

	extremitySideVertices, extremityTopVertices, extremityBottomVertices := ge.GetCapsuleVertices3(0.4, 0.07, 0.05, 8)
	extremitySideVAO, extremityTopVAO, extremityBottomVAO := ge.CreateVAO(extremitySideVertices), ge.CreateVAO(extremityTopVertices), ge.CreateVAO(extremityBottomVertices)

	headVertices, headVerticesT, headVerticesB := ge.GetSphereVertices3(0.2, 18)
	headVAO, headVAOT, headVAOB := ge.CreateVAO(headVertices), ge.CreateVAO(headVerticesT), ge.CreateVAO(headVerticesB)

	coinVertices, coinVerticesT, coinVerticesB := ge.GetCylinderVertices3(0.02, 0.05, 0.05, 8)
	coinVertices, coinVerticesT, coinVerticesB = ge.Translate(coinVertices, mgl32.Vec3{0, -0.01, 0}), ge.Translate(coinVerticesT, mgl32.Vec3{0, -0.01, 0}), ge.Translate(coinVerticesB, mgl32.Vec3{0, -0.01, 0})
	coinVAO, coinVAOT, coinVAOB := ge.CreateVAO(coinVertices), ge.CreateVAO(coinVerticesT), ge.CreateVAO(coinVerticesB)

	planeVertices := ge.GetPlaneVertices3(12, 12, 1)
	planeVAO := ge.CreateVAO(planeVertices)

	var (
		wellPosition     = mgl32.Vec3{1, 0, 1}
		wellViewPosition = mgl32.Vec3{1, 1, 1}

		camPositions = [][2]mgl32.Vec3{
			{{-4.5, 1, 4.5}, wellViewPosition},
			{{-1, 1, 3}, wellViewPosition},
			{{4.5, 2, 4}, {1, 0, 2}},
			{{3, 1, 3}, {1, 1, 1.5}},
		}

		camPathPoints = [][]mgl32.Vec3{
			{{-4, 0, -4}, {4, 0, -4}, {4, 0, 3}, {4, 0, 4}},
			{{4, 0, 4}, {4, 0, -4}, {-4, 0, -3}, {-4, 0, -4}},
			{{-4, 0, -4}, {4, 0, -4}, wellPosition, wellViewPosition},
		}

		treePositions = []mgl32.Vec3{
			{-2.5, 0, -0.5},
			{-2, 0, 1.5},
			{-2, 0, -3},
			{-1, 0, -1},
			{-1, 0, 1},
			{1, 0, -1},
			{2, 0, -1.5},
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

	personPathModel := mgl32.Translate3D(personPathPoints[0].Elem()).Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(-45)))
	armModel := mgl32.HomogRotate3DX(mgl32.DegToRad(45)).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(135)))
	coinPathModel := model

	// escenas y animacion
	angle := 0.0
	previousTime := glfw.GetTime()
	var totalElapsed float64 = 0
	movementControlCount := 0

	movementTimes := []float64{}
	var movementFunctions []func(t float32)

	// Camera movement
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], camPathPoints[0][0], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, .5)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], mgl32.BezierCurve3D(t, camPathPoints[0]), mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, 1.0)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], mgl32.BezierCurve3D(t, camPathPoints[1]), mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, 1.0)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], mgl32.BezierCurve3D(t, camPathPoints[2]), mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, .5)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[0][0], camPathPoints[2][len(camPathPoints[2])-1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, .5)
	// Camara well framing
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[1][0], camPositions[1][1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, 1)
	// Camara zoom-out framing
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[2][0], camPositions[2][1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, 1)

	// Person movement to the well
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		waves := float32(3)
		rot := math32.Sin(2*math32.Pi*waves*t) * mgl32.DegToRad(5)
		point := mgl32.BezierCurve3D(t, personPathPoints)
		roty := math32.Atan((point.X()+5)/(point.Y()+5)) - math32.Pi/4
		personPathModel = mgl32.Translate3D(point.Elem()).Mul4(mgl32.HomogRotate3DZ(rot)).Mul4(mgl32.HomogRotate3DY(roty))
	}), append(movementTimes, 2.0)

	// Camara well framing
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera := mgl32.LookAtV(camPositions[3][0], camPositions[3][1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, 1)

	// Movement of the person's arm
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		armModel = mgl32.HomogRotate3DX(float32(t*math32.Pi/2) + mgl32.DegToRad(-160))
	}), append(movementTimes, .5)
	// Movement of the coin
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		flips := float32(3)
		rot := 2 * math32.Pi * flips * t
		coinPathModel = mgl32.Translate3D(mgl32.BezierCurve3D(t, coinPathPoints).Elem()).Mul4(mgl32.HomogRotate3DX(float32(rot)))
	}), append(movementTimes, 1.0)

	for !window.ShouldClose() {

		time := glfw.GetTime()
		elapsed := time - previousTime
		totalElapsed += elapsed
		previousTime = time
		angle += elapsed

		ge.NewFrame(window, mgl32.Vec4{0, 0.27, 0.7, 1.0})
		// update
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

		//moon
		moonTranslate := mgl32.Translate3D(0, 6, -5)
		gl.Uniform3f(colorModel, 0.850, 0.850, 0.850)
		gl.BindVertexArray(sphereVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &moonTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereVertices)))

		gl.BindVertexArray(sphereTopVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &moonTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereTopVertices)))

		gl.BindVertexArray(sphereBottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &moonTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereBottomVertices)))

		//well
		wellTranslate := mgl32.Translate3D(wellPosition.Elem())

		gl.Uniform3f(colorModel, 0.301, 0.301, 0.301)
		gl.BindVertexArray(sideInVAOPipe)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideInVerticesPipe)))

		gl.BindVertexArray(sideOutVAOPipe)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideOutVerticesPipe)))

		gl.BindVertexArray(topVAOPipe)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(topVerticesPipe)))

		gl.BindVertexArray(bottomVAOPipe)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(bottomVerticesPipe)))

		//pillars
		gl.Uniform3f(colorModel, 0.301, 0.149, 0)
		wellTransform := wellTranslate.Mul4(mgl32.Translate3D(-0.55, 0, 0)).Mul4(mgl32.Scale3D(0.75, 2, 0.75))

		gl.BindVertexArray(sideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVertices)))

		gl.BindVertexArray(topVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(topVertices)))

		gl.BindVertexArray(bottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVertices)))

		wellTransform = wellTranslate.Mul4(mgl32.Translate3D(0.55, 0, 0)).Mul4(mgl32.Scale3D(0.75, 2, 0.75))

		gl.BindVertexArray(sideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVertices)))

		gl.BindVertexArray(topVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(topVertices)))

		gl.BindVertexArray(bottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVertices)))

		wellTransform = wellTranslate.Mul4(mgl32.Translate3D(0.55, 1.75, -0.15)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(90))).Mul4(mgl32.Scale3D(0.5, 0.3, 0.5))

		gl.BindVertexArray(sideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVertices)))

		gl.BindVertexArray(topVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(topVertices)))

		gl.BindVertexArray(bottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVertices)))

		wellTransform = wellTransform.Mul4(mgl32.Translate3D(-2.2, 0, 0))

		gl.BindVertexArray(sideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVertices)))

		gl.BindVertexArray(topVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(topVertices)))

		gl.BindVertexArray(bottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVertices)))

		//roof
		gl.Uniform3f(colorModel, 0.623, 0.141, 0.078)

		wellTransform = wellTranslate.Mul4(mgl32.Translate3D(0, 1.75, 0.2)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45))).Mul4(mgl32.Scale3D(1.25, 0.1, 0.75))
		gl.BindVertexArray(cubeVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

		wellTransform = wellTranslate.Mul4(mgl32.Translate3D(0, 1.75, -0.2)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(-45))).Mul4(mgl32.Scale3D(1.25, 0.1, 0.75))
		gl.BindVertexArray(cubeVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

		// Person
		personTranslate := personPathModel

		// Body
		gl.Uniform3f(colorModel, 0.24, 0.24, 0.24)
		bodyTranslate := personTranslate.Mul4(mgl32.Translate3D(0, 0.3, 0))

		gl.BindVertexArray(personSideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &bodyTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(personSideVertices)))

		gl.BindVertexArray(personTopVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &bodyTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(personTopVertices)))

		gl.BindVertexArray(personBottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &bodyTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(personBottomVertices)))

		// Legs

		gl.Uniform3f(colorModel, 0.13, 0.13, 0.13)
		legTranslate := bodyTranslate.Mul4(mgl32.Translate3D(0.1, 0.1, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(180)))

		gl.BindVertexArray(extremitySideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremitySideVertices)))

		gl.BindVertexArray(extremityTopVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityTopVertices)))

		gl.BindVertexArray(extremityBottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityBottomVertices)))

		legTranslate = bodyTranslate.Mul4(mgl32.Translate3D(-0.1, 0.1, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(180)))

		gl.BindVertexArray(extremitySideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremitySideVertices)))

		gl.BindVertexArray(extremityTopVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityTopVertices)))

		gl.BindVertexArray(extremityBottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &legTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityBottomVertices)))

		// Arms
		armsTranslate := bodyTranslate.Mul4(mgl32.Translate3D(0.1, 0.6, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45))).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(-135)))

		gl.Uniform3f(colorModel, 0.84, 0.8, 0.66)

		gl.BindVertexArray(extremitySideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremitySideVertices)))

		gl.BindVertexArray(extremityTopVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityTopVertices)))

		gl.BindVertexArray(extremityBottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityBottomVertices)))

		armsTranslate = bodyTranslate.Mul4(mgl32.Translate3D(-0.1, 0.6, 0)).Mul4(armModel)

		gl.BindVertexArray(extremitySideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremitySideVertices)))

		gl.BindVertexArray(extremityTopVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityTopVertices)))

		gl.BindVertexArray(extremityBottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityBottomVertices)))

		// Head
		headTranslate := bodyTranslate.Mul4(mgl32.Translate3D(0, 0.65, 0))
		gl.BindVertexArray(headVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &headTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(headVertices)))

		gl.BindVertexArray(headVAOT)
		gl.UniformMatrix4fv(modelUniform, 1, false, &headTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(headVerticesT)))

		gl.BindVertexArray(headVAOB)
		gl.UniformMatrix4fv(modelUniform, 1, false, &headTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(headVerticesB)))

		// Coin
		coinTranslate := armsTranslate.Mul4(mgl32.Translate3D(0, 0.4, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(90))).Mul4(coinPathModel)

		gl.Uniform3f(colorModel, 1, 1, 0)
		gl.BindVertexArray(coinVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &coinTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(coinVertices)))

		gl.BindVertexArray(coinVAOT)
		gl.UniformMatrix4fv(modelUniform, 1, false, &coinTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(coinVerticesT)))

		gl.BindVertexArray(coinVAOB)
		gl.UniformMatrix4fv(modelUniform, 1, false, &coinTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(coinVerticesB)))

		//tree
		for _, pos := range treePositions {
			treeTranslate := mgl32.Translate3D(pos.Elem())

			//trunk
			gl.Uniform3f(colorModel, 0.4, 0.2, 0)
			gl.BindVertexArray(sideVAO)
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVertices)))

			gl.BindVertexArray(topVAO)
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(topVertices)))

			gl.BindVertexArray(bottomVAO)
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVertices)))

			//leaves
			scale1 := 1 - math32.Abs(math32.Sin(float32(time)))*0.04
			scale2 := 1 - math32.Abs(math32.Cos(float32(time)))*0.04

			gl.Uniform3f(colorModel, 0, 0.9, 0)
			gl.BindVertexArray(cubeVAO)
			worldTranslate := treeTranslate.Mul4(mgl32.Scale3D(1, 1*scale1, 1)).Mul4(mgl32.Translate3D(0, 1.25, -0.5))
			gl.UniformMatrix4fv(modelUniform, 1, false, &worldTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

			gl.Uniform3f(colorModel, 0, 0.75, 0)
			worldTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.75, 0.75*scale1, 0.75)).Mul4(mgl32.Translate3D(0, 1.2, 0.4))
			gl.UniformMatrix4fv(modelUniform, 1, false, &worldTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

			gl.Uniform3f(colorModel, 0, 0.5, 0)
			worldTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.8, 0.8*scale2, 0.8)).Mul4(mgl32.Translate3D(-0.5, 1.5, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &worldTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
		}

		gl.Uniform3f(colorModel, 0.4, 0.6, 0)
		gl.BindVertexArray(planeVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
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
