package main

import (
	"log"
	"runtime"

	"git.maze.io/go/math32"
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

	// set texture0 to uniform0 in the fragment shader

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
	camera := mgl32.LookAtV(mgl32.Vec3{3.5, 0.5, 3.5}, mgl32.Vec3{0, 2, 0}, mgl32.Vec3{0, 1, 0})
	//camera := mgl32.LookAtV(mgl32.Vec3{5, 5, 5}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
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
	leavesTexture, err := gfx.NewTextureFromFile("images/leaves.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	grassTexture, err := gfx.NewTextureFromFile("images/grass1.png",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	roofTexture, err := gfx.NewTextureFromFile("images/roof1.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// Get primitive vertices and create VAOs
	cubeVertices := ge.GetCubicHexahedronVertices3(1, 1, 1)
	cubeTextureCoords := ge.GetCubicHexahedronTextureCoords(1, 1, 1)
	cubeVAO := ge.CreateVAO(cubeVertices, cubeTextureCoords)

	sphereVertices, sphereVerticesT, sphereVerticesB := ge.GetSphereVertices3(1, 32)
	sphereVAO, sphereVAOT, sphereVAOB := ge.CreateVAO(sphereVertices, []mgl32.Vec2{}), ge.CreateVAO(sphereVerticesT, []mgl32.Vec2{}), ge.CreateVAO(sphereVerticesB, []mgl32.Vec2{})

	cylinderVertices, cylinderVerticesT, cylinderVerticesB := ge.GetCylinderVertices3(1, 0.1, 0.1, 5)
	cylinderVAO, cylinderVAOT, cylinderVAOB := ge.CreateVAO(cylinderVertices, []mgl32.Vec2{}), ge.CreateVAO(cylinderVerticesT, []mgl32.Vec2{}), ge.CreateVAO(cylinderVerticesB, []mgl32.Vec2{})

	PipeVerticesSI, PipeVerticesSO, PipeVerticesT, PipeVerticesSB := ge.GetPipeVertices3(0.75, 0.4, 0.5, 16)
	PipeVAOSI, PipeVAOSO, PipeVAOT, PipeVAOB := ge.CreateVAO(PipeVerticesSI, []mgl32.Vec2{}), ge.CreateVAO(PipeVerticesSO, []mgl32.Vec2{}), ge.CreateVAO(PipeVerticesT, []mgl32.Vec2{}), ge.CreateVAO(PipeVerticesSB, []mgl32.Vec2{})

	planeVertices := ge.GetPlaneVertices3(12, 12, 1)
	planeTextureCoords := ge.GetPlaneTextureCoords(12, 12, 1)
	planeVAO := ge.CreateVAO(planeVertices, planeTextureCoords)

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
		wellTranslate := mgl32.Translate3D(1, 0, 1)

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
		roofTexture.Bind(gl.TEXTURE0)
		roofTexture.SetUniform(textureUniformLocation)

		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

		wellTransform = wellTranslate.Mul4(mgl32.Translate3D(0, 1.75, -0.2)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(-45))).Mul4(mgl32.Scale3D(1.25, 0.1, 0.75)).Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(180)))

		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &wellTransform[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

		roofTexture.UnBind()

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
			scale1 := 1 - math32.Pow(math32.Abs(math32.Sin(float32(time*.7))), (1.3))*0.03
			scale2 := 1 - math32.Pow(math32.Abs(math32.Cos(float32(time*.7))), (1.3))*0.03

			gl.BindVertexArray(cubeVAO)
			leavesTexture.Bind(gl.TEXTURE0)
			leavesTexture.SetUniform(textureUniformLocation)

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

			leavesTexture.UnBind()
		}

		gl.BindVertexArray(planeVAO)
		grassTexture.Bind(gl.TEXTURE0)
		grassTexture.SetUniform(textureUniformLocation)
		gl.Uniform3f(colorUniformLocation, 0.4, 0.6, 0)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &model[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(planeVertices)))
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
