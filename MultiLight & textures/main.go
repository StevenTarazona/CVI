package main

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"git.maze.io/go/math32"
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
	pointLightPositions = []mgl32.Vec3{
		{0, 0, 0},
		{0, 0, 0},
	}
)

func createVAO(vertices, normals, tCoords []float32, indices []uint32) uint32 {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.BindVertexArray(VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	var NBO uint32
	gl.GenBuffers(1, &NBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, NBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(normals)*4, gl.Ptr(normals), gl.STATIC_DRAW)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	if len(tCoords) > 0 {
		var TBO uint32
		gl.GenBuffers(1, &TBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, TBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(tCoords)*4, gl.Ptr(tCoords), gl.STATIC_DRAW)
		gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(2)
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	}

	var EBO uint32
	gl.GenBuffers(1, &EBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindVertexArray(0)

	return VAO
}

func pointLightsUniformLocations(program *gfx.Program) [][]int32 {
	uniformLocations := [][]int32{}
	for i := 0; i < len(pointLightPositions); i++ {
		uniformLocations = append(uniformLocations,
			[]int32{program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].position")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].ambient")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].diffuse")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].specular")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].constant")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].linear")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].quadratic"))})
	}
	return uniformLocations
}

func programLoop(window *win.Window) error {

	// Shaders and textures
	vertShader, err := gfx.NewShaderFromFile("shaders/phong_ml.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fragShader, err := gfx.NewShaderFromFile("shaders/phong_ml.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}
	defer program.Delete()

	sourceVertShader, err := gfx.NewShaderFromFile("shaders/source.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	sourceFragShader, err := gfx.NewShaderFromFile("shaders/source.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	// special shader program so that lights themselves are not affected by lighting
	sourceProgram, err := gfx.NewProgram(sourceVertShader, sourceFragShader)
	if err != nil {
		return err
	}

	// Ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Base model
	model := mgl32.Ident4()

	// Uniform
	modelUniformLocation := program.GetUniformLocation("model")
	viewUniformLocation := program.GetUniformLocation("view")
	projectUniformLocation := program.GetUniformLocation("projection")
	lightColorUniformLocation := program.GetUniformLocation("lightColor")
	objectColorUniformLocation := program.GetUniformLocation("objectColor")
	viewPosUniformLocation := program.GetUniformLocation("viewPos")
	numLightsUniformLocation := program.GetUniformLocation("numLights")
	textureUniformLocation := program.GetUniformLocation("texSampler")

	modelSourceUniformLocation := sourceProgram.GetUniformLocation("model")
	viewSourceUniformLocation := sourceProgram.GetUniformLocation("view")
	projectSourceUniformLocation := sourceProgram.GetUniformLocation("projection")
	objectColorSourceUniformLocation := sourceProgram.GetUniformLocation("objectColor")
	textureSourceUniformLocation := sourceProgram.GetUniformLocation("texSampler")

	pointLightsUniformLocations := pointLightsUniformLocations(program)

	// creates camara
	eye := mgl32.Vec3{0, 0, 5}
	center := mgl32.Vec3{1, 1, 0}
	camera := mgl32.LookAtV(eye, center, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(viewUniformLocation, 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// Textures
	earthTexture, err := gfx.NewTextureFromFile("textures/earth3.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	starsTexture, err := gfx.NewTextureFromFile("textures/stars.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// Colors
	objectColor := mgl32.Vec3{1., 0., 1.}
	backgroundColor := mgl32.Vec3{0, 0, 0}
	lightColor := mgl32.Vec3{1, 1, 0.7}

	// Uncomment to turn on polygon mode
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Geometry
	xSegments := 30
	ySegments := 30
	VAO := createVAO(Sphere(xSegments, ySegments))
	lightVAO := VAO
	skyVAO := createVAO(Cube(10, 10, 10))

	// Scene and animation always needs to be after the model and buffers initialization
	animationCtl := gfx.NewAnimationManager()

	r := float32(3)
	animationCtl.AddContunuousAnimation(func() {
		pointLightPositions[0] = mgl32.Vec3{r * math32.Cos(float32(animationCtl.GetAngle()/2)), 0, r * math32.Sin(float32(animationCtl.GetAngle()/2))}
		pointLightPositions[1] = mgl32.Vec3{0, r * math32.Sin(float32(animationCtl.GetAngle()/2)), r * math32.Cos(float32(animationCtl.GetAngle()/2))}
	})

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
		gl.UniformMatrix4fv(viewUniformLocation, 1, false, &camera[0])
		gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

		gl.Uniform3fv(viewPosUniformLocation, 1, &eye[0])
		gl.Uniform3f(objectColorUniformLocation, objectColor.X(), objectColor.Y(), objectColor.Z())
		gl.Uniform3f(lightColorUniformLocation, lightColor.X(), lightColor.Y(), lightColor.Z())
		gl.Uniform1i(numLightsUniformLocation, int32(len(pointLightPositions)))

		//luces
		for index, pointLightPosition := range pointLightPositions {
			gl.Uniform3fv(pointLightsUniformLocations[index][0], 1, &pointLightPosition[0])
			gl.Uniform3f(pointLightsUniformLocations[index][1], (backgroundColor.X()+0.2)*0.3, (backgroundColor.Y()+0.2)*0.3, (backgroundColor.Z()+0.2)*0.3)
			gl.Uniform3f(pointLightsUniformLocations[index][2], 0.8, 0.8, 0.8)
			gl.Uniform3f(pointLightsUniformLocations[index][3], 1., 1., 1.)
			gl.Uniform1f(pointLightsUniformLocations[index][4], 1.)
			gl.Uniform1f(pointLightsUniformLocations[index][5], 0.09)
			gl.Uniform1f(pointLightsUniformLocations[index][6], 0.032)
		}

		// render models
		gl.BindVertexArray(VAO)
		earthTexture.Bind(gl.TEXTURE0)
		earthTexture.SetUniform(textureUniformLocation)

		boxModel := model.Mul4(mgl32.HomogRotate3DY(float32(animationCtl.GetAngle())))

		gl.UniformMatrix4fv(modelUniformLocation, 1, false, &boxModel[0])
		gl.DrawElements(gl.TRIANGLES, int32(xSegments*ySegments)*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		earthTexture.UnBind()
		gl.BindVertexArray(0)

		// obj is colored, light is white
		sourceProgram.Use()
		gl.UniformMatrix4fv(projectSourceUniformLocation, 1, false, &projectTransform[0])
		gl.UniformMatrix4fv(viewSourceUniformLocation, 1, false, &camera[0])
		gl.Uniform3f(objectColorSourceUniformLocation, lightColor.X(), lightColor.Y(), lightColor.Z())
		gl.BindVertexArray(lightVAO)
		for _, lp := range pointLightPositions {
			cubeM := mgl32.Ident4()
			cubeM = cubeM.Mul4(mgl32.Translate3D(lp.Elem())).Mul4(mgl32.Scale3D(0.2, 0.2, 0.2))
			gl.UniformMatrix4fv(modelSourceUniformLocation, 1, false, &cubeM[0])
			gl.DrawElements(gl.TRIANGLES, int32(xSegments*ySegments)*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		}
		//gl.BindVertexArray(0)

		gl.BindVertexArray(skyVAO)
		starsTexture.Bind(gl.TEXTURE0)
		starsTexture.SetUniform(textureSourceUniformLocation)

		gl.UniformMatrix4fv(modelSourceUniformLocation, 1, false, &model[0])
		gl.DrawElements(gl.TRIANGLES, 6*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		starsTexture.UnBind()
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
