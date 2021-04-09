package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"unsafe"

	"git.maze.io/go/math32"
	"github.com/kaitsubaka/glutils/gfx"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width  = 1280
	height = 720
	title  = "Lights & Textures"
)

var (
	pointLightPositions = []mgl32.Vec3{
		{-2, .5, 0},

		{-1, -3, 10},
		{0, -3, 10},
		{1, -3, 10},
	}
	pointLightColors = []mgl32.Vec3{
		{1, 1, 0.7},

		{1, 0, 1},
		{0, 1, 1},
		{1, 1, 0},
	}
	lightPathPoints = [][]mgl32.Vec3{
		{pointLightPositions[1], {-1, 4, 10}, {-2, 4, 5}, {-2, 4, 0}, {0, 4, 0}},
		{pointLightPositions[2], {0, 4, 10}, {0, 4, 2}, {0, 4, 0}, {-2, 4, 0}},
		{pointLightPositions[3], {1, 4, 10}, {2, 4, 5}, {2, 4, 0}, {0, 4, 0}},
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
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].quadratic")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].lightColor"))})
	}
	return uniformLocations
}

func treePos(z0, zf, x0, xf, sparse float32) (positions []mgl32.Vec3, angles []float32) {
	z := z0 + sparse/2
	for z <= zf-sparse/2 {
		x := x0 + sparse/2
		for x <= xf-sparse/2 {
			amplitude := (sparse / 2) * .75
			min := x - amplitude
			max := x + amplitude
			xPos := min + rand.Float32()*(max-min)
			yPos := float32(0)
			min = z - amplitude
			max = z + amplitude
			zPos := min + rand.Float32()*(max-min)
			positions = append(positions, mgl32.Vec3{xPos, yPos, zPos})
			angles = append(angles, rand.Float32()*math32.Pi*2)
			x += sparse
		}
		z += sparse
	}
	return
}

func turnLight(im *InputManager, colorNum int, changeColor bool) (mgl32.Vec3, int, bool) {
	colors := []mgl32.Vec3{
		{1, 1, 0.7},
		{0.1, 0.1, 0.1},
	}
	if im.IsActive(SWITCH) {
		if changeColor {
			colorNum = (colorNum + 1) % 2
			return colors[colorNum], colorNum, false
		}
		return colors[colorNum], colorNum, changeColor
	} else {
		return colors[colorNum], colorNum, true
	}
}

func programLoop(window *Window) error {

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

	// Uniform
	modelUniformLocation := program.GetUniformLocation("model")
	viewUniformLocation := program.GetUniformLocation("view")
	projectUniformLocation := program.GetUniformLocation("projection")
	objectColorUniformLocation := program.GetUniformLocation("objectColor")
	viewPosUniformLocation := program.GetUniformLocation("viewPos")
	numLightsUniformLocation := program.GetUniformLocation("numLights")
	texture0UniformLocation := program.GetUniformLocation("texSampler0")
	texture1UniformLocation := program.GetUniformLocation("texSampler1")
	pointLightsUniformLocations := pointLightsUniformLocations(program)

	modelSourceUniformLocation := sourceProgram.GetUniformLocation("model")
	viewSourceUniformLocation := sourceProgram.GetUniformLocation("view")
	projectSourceUniformLocation := sourceProgram.GetUniformLocation("projection")
	objectColorSourceUniformLocation := sourceProgram.GetUniformLocation("objectColor")
	textureSourceUniformLocation := sourceProgram.GetUniformLocation("texSampler")

	// creates camara
	eye := mgl32.Vec3{0, 1, -5}
	camera := NewFpsCamera(eye, mgl32.Vec3{0, -1, 0}, 90, 0, window.InputManager())

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 150.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// Textures
	earthTexture, err := gfx.NewTextureFromFile("textures/snow.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	pathTexture, err := gfx.NewTextureFromFile("textures/path.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	starsTexture, err := gfx.NewTextureFromFile("textures/stars.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	woodTexture, err := gfx.NewTextureFromFile("textures/wood.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	energyTexture, err := gfx.NewTextureFromFile("textures/energy.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// Colors
	objectColor := mgl32.Vec3{1.0, 1.0, 1.0}
	backgroundColor := mgl32.Vec3{1, 1, 1}

	// Uncomment to turn on polygon mode
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Models
	model := mgl32.Ident4()
	skyRotate := model
	orbRotate := model

	// Geometry
	xLightSegments, yLighteSegments := 30, 30
	lightVAO := createVAO(Sphere(xLightSegments, yLighteSegments))
	xPlaneSegments, yPlaneSegments := 15, 15
	planeVAO := createVAO(Square(xPlaneSegments, yPlaneSegments, 1))
	skyVAO := createVAO(Cube(50, 50, 50))
	xTrunkSegments, yTrunkSegments, zTrunkSegments := 15, 15, 2
	trunkVAO := createVAO(Cylinder(xTrunkSegments, yTrunkSegments, zTrunkSegments))
	treePos, treeAngles := treePos(-float32(yPlaneSegments)/2, float32(yPlaneSegments)/2, -float32(xPlaneSegments)/2, float32(xPlaneSegments)/2, 1.5)

	var numColor int
	var change bool
	pointLightColors[0], numColor, change = turnLight(window.InputManager(), 0, true)

	// Scene and animation always needs to be after the model and buffers initialization
	animationCtl := gfx.NewAnimationManager()
	startDancing := false
	Dance := func() {
		pointLightColors[0], numColor, change = turnLight(window.InputManager(), 1, true)
		animationCtl.AddAnimation(
			func(t float32) {
				for i, v := range lightPathPoints {
					pointLightPositions[i+1] = mgl32.BezierCurve3D(math32.Pow(t, 2), v)
				}
			}, 4,
		)
		animationCtl.AddAnimation(
			func(t float32) {
				y := float32(4)
				r := float32(2)
				tTemp := t * 2 * math32.Pi * 2
				pointLightPositions[2] = mgl32.Vec3{-r * math32.Cos(tTemp), y, -r * math32.Sin(tTemp)}
				r = float32(5)
				min := -math32.Pi / float32(2)
				max := (float32(3) / 2) * math32.Pi
				tTemp = min + t*(max-min)
				pointLightPositions[1] = mgl32.Vec3{r * math32.Cos(tTemp), y, r * math32.Cos(tTemp) * math32.Sin(tTemp)}
				pointLightPositions[3] = mgl32.Vec3{-r * math32.Cos(tTemp), y, r * math32.Cos(tTemp) * math32.Sin(tTemp)}
				if t == 1 {
					animationCtl.Repeat()
				}
			}, 8,
		)
		animationCtl.Reset()
	}
	animationCtl.AddContunuousAnimation(func() {
		skyRotate = mgl32.HomogRotate3D(float32(animationCtl.GetAngle()*.02), mgl32.Vec3{1, 0, 1})
		orbRotate = mgl32.HomogRotate3DY(float32(animationCtl.GetAngle() - float64(mgl32.DegToRad(float32(camera.getAngle())))))
	})
	animationCtl.Init() // always needs to be before the main loop in order to get correct times
	// main loop
	for !window.ShouldClose() {
		window.StartFrame()
		camera.Update(window.SinceLastFrame())
		eye = camera.getPos()
		pointLightPositions[0] = eye.Add(camera.getFront()).Add(mgl32.Vec3{0, -.25, 0})
		if eye.Z() > 0 && !startDancing {
			startDancing = true
			Dance()
		}

		pointLightColors[0], numColor, change = turnLight(window.InputManager(), numColor, change)

		// background color
		gl.ClearColor(backgroundColor.X(), backgroundColor.Y(), backgroundColor.Z(), 1.)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Scene update
		animationCtl.Update()

		// You shall draw here
		program.Use()

		camTransform := camera.GetTransform()
		gl.UniformMatrix4fv(viewUniformLocation, 1, false, &camTransform[0])
		gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

		gl.Uniform3fv(viewPosUniformLocation, 1, &eye[0])
		gl.Uniform3f(objectColorUniformLocation, objectColor.X(), objectColor.Y(), objectColor.Z())
		gl.Uniform1i(numLightsUniformLocation, int32(len(pointLightPositions)))

		//Lights
		for index, pointLightPosition := range pointLightPositions {
			if index == 0 {
				gl.Uniform3fv(pointLightsUniformLocations[index][0], 1, &pointLightPosition[0])
				gl.Uniform3f(pointLightsUniformLocations[index][1], .01, .01, .01)
				gl.Uniform3f(pointLightsUniformLocations[index][2], 1, 1, 1)
				gl.Uniform3f(pointLightsUniformLocations[index][3], 2, 2, 2)
				gl.Uniform1f(pointLightsUniformLocations[index][4], 1.)
				gl.Uniform1f(pointLightsUniformLocations[index][5], 0.09)
				gl.Uniform1f(pointLightsUniformLocations[index][6], 0.032)
				gl.Uniform3f(pointLightsUniformLocations[index][7], pointLightColors[index].X(), pointLightColors[index].Y(), pointLightColors[index].Z())

			} else {
				gl.Uniform3fv(pointLightsUniformLocations[index][0], 1, &pointLightPosition[0])
				gl.Uniform3f(pointLightsUniformLocations[index][1], (backgroundColor.X())*0.01, (backgroundColor.Y())*0.01, (backgroundColor.Z())*0.01)
				gl.Uniform3f(pointLightsUniformLocations[index][2], 0.8, 0.8, 0.8)
				gl.Uniform3f(pointLightsUniformLocations[index][3], 1., 1., 1.)
				gl.Uniform1f(pointLightsUniformLocations[index][4], 1.)
				gl.Uniform1f(pointLightsUniformLocations[index][5], 0.09)
				gl.Uniform1f(pointLightsUniformLocations[index][6], 0.032)
				gl.Uniform3f(pointLightsUniformLocations[index][7], pointLightColors[index].X(), pointLightColors[index].Y(), pointLightColors[index].Z())
			}
		}

		// render models
		//Trees

		gl.BindVertexArray(trunkVAO)
		woodTexture.Bind(gl.TEXTURE0)
		woodTexture.SetUniform(texture0UniformLocation)
		for i, pos := range treePos {
			if pos.X() > -1 && pos.X() < 1 {
				continue
			}
			//Trnk
			trunkTranslate := model.Mul4(mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())).Mul4(mgl32.HomogRotate3DY(treeAngles[i]))
			trunkTransform := trunkTranslate.Mul4(mgl32.Scale3D(.2, 2.0, .2))
			gl.UniformMatrix4fv(modelUniformLocation, 1, false, &trunkTransform[0])
			gl.DrawElements(gl.TRIANGLES, int32(6*xTrunkSegments*(yTrunkSegments+2*zTrunkSegments-1)), gl.UNSIGNED_INT, unsafe.Pointer(nil))

			//Branches
			branchScale := mgl32.Scale3D(.1, 1.0, .1)

			trunkTranslate = trunkTranslate.Mul4(mgl32.Translate3D(0, 1.3, 0))
			trunkTransform = trunkTranslate.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(45), mgl32.Vec3{1, 0, 1})).Mul4(branchScale)
			gl.UniformMatrix4fv(modelUniformLocation, 1, false, &trunkTransform[0])
			gl.DrawElements(gl.TRIANGLES, int32(6*xTrunkSegments*(yTrunkSegments+2*zTrunkSegments-1)), gl.UNSIGNED_INT, unsafe.Pointer(nil))

			trunkTranslate = trunkTranslate.Mul4(mgl32.Translate3D(0, .1, 0))
			trunkTransform = trunkTranslate.Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(-45))).Mul4(branchScale)
			gl.UniformMatrix4fv(modelUniformLocation, 1, false, &trunkTransform[0])
			gl.DrawElements(gl.TRIANGLES, int32(6*xTrunkSegments*(yTrunkSegments+2*zTrunkSegments-1)), gl.UNSIGNED_INT, unsafe.Pointer(nil))

			trunkTranslate = trunkTranslate.Mul4(mgl32.Translate3D(0, .2, 0))
			trunkTransform = trunkTranslate.Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(45))).Mul4(branchScale)
			gl.UniformMatrix4fv(modelUniformLocation, 1, false, &trunkTransform[0])
			gl.DrawElements(gl.TRIANGLES, int32(6*xTrunkSegments*(yTrunkSegments+2*zTrunkSegments-1)), gl.UNSIGNED_INT, unsafe.Pointer(nil))

			trunkTranslate = trunkTranslate.Mul4(mgl32.Translate3D(0, .1, 0))
			trunkTransform = trunkTranslate.Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45))).Mul4(branchScale)
			gl.UniformMatrix4fv(modelUniformLocation, 1, false, &trunkTransform[0])
			gl.DrawElements(gl.TRIANGLES, int32(6*xTrunkSegments*(yTrunkSegments+2*zTrunkSegments-1)), gl.UNSIGNED_INT, unsafe.Pointer(nil))

			trunkTranslate = trunkTranslate.Mul4(mgl32.Translate3D(0, 0, 0))
			trunkTransform = trunkTranslate.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(-45), mgl32.Vec3{1, 0, -.5})).Mul4(branchScale)
			gl.UniformMatrix4fv(modelUniformLocation, 1, false, &trunkTransform[0])
			gl.DrawElements(gl.TRIANGLES, int32(6*xTrunkSegments*(yTrunkSegments+2*zTrunkSegments-1)), gl.UNSIGNED_INT, unsafe.Pointer(nil))

			trunkTranslate = trunkTranslate.Mul4(mgl32.Translate3D(0, .1, 0))
			trunkTransform = trunkTranslate.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(-45), mgl32.Vec3{.5, 0, 1})).Mul4(branchScale)
			gl.UniformMatrix4fv(modelUniformLocation, 1, false, &trunkTransform[0])
			gl.DrawElements(gl.TRIANGLES, int32(6*xTrunkSegments*(yTrunkSegments+2*zTrunkSegments-1)), gl.UNSIGNED_INT, unsafe.Pointer(nil))
		}

		woodTexture.UnBind()
		gl.BindVertexArray(0)

		//Plane
		gl.BindVertexArray(planeVAO)
		earthTexture.Bind(gl.TEXTURE0)
		earthTexture.SetUniform(texture0UniformLocation)

		pathTexture.Bind(gl.TEXTURE1)
		pathTexture.SetUniform(texture1UniformLocation)

		gl.UniformMatrix4fv(modelUniformLocation, 1, false, &model[0])
		gl.DrawElements(gl.TRIANGLES, int32(xPlaneSegments*yPlaneSegments)*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		earthTexture.UnBind()
		pathTexture.UnBind()
		gl.BindVertexArray(0)

		//Source program
		sourceProgram.Use()
		gl.UniformMatrix4fv(projectSourceUniformLocation, 1, false, &projectTransform[0])
		gl.UniformMatrix4fv(viewSourceUniformLocation, 1, false, &camTransform[0])

		//Sky box
		gl.BindVertexArray(skyVAO)
		starsTexture.Bind(gl.TEXTURE0)
		starsTexture.SetUniform(textureSourceUniformLocation)
		gl.Uniform3f(objectColorSourceUniformLocation, backgroundColor.X(), backgroundColor.Y(), backgroundColor.Z())
		gl.UniformMatrix4fv(modelSourceUniformLocation, 1, false, &skyRotate[0])
		gl.DrawElements(gl.TRIANGLES, 6*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		starsTexture.UnBind()
		gl.BindVertexArray(0)

		//Light objects
		gl.BindVertexArray(lightVAO)
		for i, lp := range pointLightPositions {
			gl.Uniform3f(objectColorSourceUniformLocation, pointLightColors[i].X(), pointLightColors[i].Y(), pointLightColors[i].Z())
			cubeM := mgl32.Ident4()
			if i == 0 {
				energyTexture.Bind(gl.TEXTURE0)
				energyTexture.SetUniform(textureSourceUniformLocation)
				cubeM = cubeM.Mul4(mgl32.Translate3D(lp.Elem())).Mul4(orbRotate).Mul4(mgl32.Scale3D(0.1, 0.1, 0.1))
			} else {
				cubeM = cubeM.Mul4(mgl32.Translate3D(lp.Elem())).Mul4(mgl32.Scale3D(0.2, 0.2, 0.2))
			}
			gl.UniformMatrix4fv(modelSourceUniformLocation, 1, false, &cubeM[0])
			gl.DrawElements(gl.TRIANGLES, int32(xLightSegments*yLighteSegments)*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
			energyTexture.UnBind()
		}
		gl.BindVertexArray(0)
	}

	return nil
}

func main() {

	runtime.LockOSThread()
	InitGlfw(4, 0)
	defer glfw.Terminate()
	window := NewWindow(width, height, title)
	gfx.InitGl()

	err := programLoop(window)
	if err != nil {
		log.Fatal(err)
	}
}
