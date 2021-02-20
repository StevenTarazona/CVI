package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"git.maze.io/go/math32"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width              = 1080
	height             = 720
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

type (
	Action       int
	InputManager struct {
		actionToKeyMap map[Action]glfw.Key
		keysPressed    [glfw.KeyLast]bool
	}
)

const (
	PLAYER_LEFT  Action = iota
	PLAYER_RIGHT Action = iota
	PROGRAM_QUIT Action = iota
	PLAYER_COLOR Action = iota
)

func NewInputManager() *InputManager {
	actionToKeyMap := map[Action]glfw.Key{
		PLAYER_LEFT:  glfw.KeyA,
		PLAYER_RIGHT: glfw.KeyD,
		PROGRAM_QUIT: glfw.KeyEscape,
		PLAYER_COLOR: glfw.KeySpace,
	}

	return &InputManager{
		actionToKeyMap: actionToKeyMap,
	}
}

// IsActive returns whether the given Action is currently active
func (im *InputManager) IsActive(a Action) bool {
	return im.keysPressed[im.actionToKeyMap[a]]
}

func (im *InputManager) keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {

	// timing for key events occurs differently from what the program loop requires
	// so just track what key actions occur and then access them in the program loop
	switch action {
	case glfw.Press:
		im.keysPressed[key] = true
	case glfw.Release:
		im.keysPressed[key] = false
	}
}

func getAngleIn(im *InputManager) float32 {
	angle := float32(0.0)
	if im.IsActive(PLAYER_LEFT) {
		angle = angle + 0.01
	}
	if im.IsActive(PLAYER_RIGHT) {
		angle = angle + -0.01
	}
	return angle
}

func getColorIn(im *InputManager, colorNum int, changeColor bool) ([]mgl32.Vec3, int, bool) {
	colors := [][]mgl32.Vec3{
		{{0, 0.9, 0}, {0, 0.75, 0}, {0, 0.5, 0}},
		{{1, 0.9, 0}, {1, 0.75, 0}, {1, 0.5, 0}},
		{{0.9, 0, 0}, {0.75, 0, 0}, {0.5, 0, 0}},
	}
	if im.IsActive(PLAYER_COLOR) {
		if changeColor {
			colorNum = int(math32.Mod(float32(colorNum+1), 3))
			return colors[colorNum], colorNum, false
		}
		return colors[colorNum], colorNum, changeColor
	}
	return colors[colorNum], colorNum, true

}

func getCircleVertices(x float32, y float32, z float32, r float32, vertices int) []mgl32.Vec3 {
	var sectorStep = 2 * math32.Pi / float32(vertices)
	var sectorAngle float32 // radian
	var unitCircleVertices = []mgl32.Vec3{{x, y, z}}
	for i := 0; i <= vertices; i++ {
		sectorAngle = float32(i) * sectorStep
		unitCircleVertices = append(unitCircleVertices, mgl32.Vec3{x + r*math32.Cos(sectorAngle), y, z + r*math32.Sin(sectorAngle)})
	}
	return unitCircleVertices
}

func getConeVertices(h float32, r float32, vertices int) (base []mgl32.Vec3, cone []mgl32.Vec3) {
	base = getCircleVertices(0, 0, 0, r, vertices)
	cone = base
	cone[0] = mgl32.Vec3{0, h, 0}
	return base, cone
}

func getRingVerticies(x float32, y float32, z float32, rIn float32, rOut float32, vertices int) []mgl32.Vec3 {

	in := getCircleVertices(x, y, z, rIn, vertices)
	out := getCircleVertices(x, y, z, rOut, vertices)

	var ring []mgl32.Vec3

	for i := 1; i < len(out); i++ {
		ring = append(ring, out[i])
		ring = append(ring, in[i])
	}
	return ring
}

func getCylinderVertices(h float32, r float32, vertices int) (side []mgl32.Vec3, top []mgl32.Vec3, bottom []mgl32.Vec3) {
	var slices int
	if slices = int(h / (math32.Sin(2*math32.Pi/float32(vertices)) * r)); slices == 0 {
		slices = 1
	}
	top = getCircleVertices(0, h, 0, r, vertices)
	bottom = getCircleVertices(0, 0, 0, r, vertices)
	for slice := 0; slice < slices; slice++ {
		for i := 1; i < len(bottom); i++ {
			side = append(side, mgl32.Vec3{bottom[i].X(), bottom[i].Y() + float32(slice)*(h/float32(slices)), bottom[i].Z()})
			side = append(side, mgl32.Vec3{bottom[i].X(), bottom[i].Y() + float32(slice+1)*(h/float32(slices)), bottom[i].Z()})
		}
	}
	return side, top, bottom
}

func getPipeVertices(h float32, rIn float32, rOut float32, vertices int) (sideIn []mgl32.Vec3, sideOut []mgl32.Vec3, top []mgl32.Vec3, bottom []mgl32.Vec3) {
	var slices int
	if slices = int(h / (math32.Sin(2*math32.Pi/float32(vertices)) * rOut)); slices == 0 {
		slices = 1
	}

	bottomIn := getCircleVertices(0, 0, 0, rIn, vertices)
	bottomOut := getCircleVertices(0, 0, 0, rOut, vertices)
	for slice := 0; slice < slices; slice++ {
		for i := 1; i < len(bottomOut); i++ {
			sideIn = append(sideIn, mgl32.Vec3{bottomIn[i].X(), bottomIn[i].Y() + float32(slice)*(h/float32(slices)), bottomIn[i].Z()})
			sideIn = append(sideIn, mgl32.Vec3{bottomIn[i].X(), bottomIn[i].Y() + float32(slice+1)*(h/float32(slices)), bottomIn[i].Z()})
			sideOut = append(sideOut, mgl32.Vec3{bottomOut[i].X(), bottomOut[i].Y() + float32(slice)*(h/float32(slices)), bottomOut[i].Z()})
			sideOut = append(sideOut, mgl32.Vec3{bottomOut[i].X(), bottomOut[i].Y() + float32(slice+1)*(h/float32(slices)), bottomOut[i].Z()})
		}
	}

	top = getRingVerticies(0, h, 0, rIn, rOut, vertices)
	bottom = getRingVerticies(0, 0, 0, rIn, rOut, vertices)

	return sideIn, sideOut, top, bottom
}

func getPlaneVertices(h int, w int, l int) []mgl32.Vec3 {
	var vertices []mgl32.Vec3
	var i = -1
	for row := -h / 2; row < h/2; row += l {
		i *= -1
		for col := -w / 2; col <= w/2; col += l {
			vertices = append(vertices, mgl32.Vec3{float32(col * i), 0, float32(row)})
			vertices = append(vertices, mgl32.Vec3{float32(col * i), 0, float32(row + l)})
		}
	}
	return vertices
}

func getSphereVertices(r float32, numVertex int, slices int) (sphere []mgl32.Vec3, top []mgl32.Vec3, bottom []mgl32.Vec3) {
	sliceHeight := r / float32(slices)
	auxHeight := r - sliceHeight
	auxR := math32.Sqrt(math32.Pow(r, 2) - math32.Pow(auxHeight, 2))
	auxNextRHeight := auxHeight - sliceHeight
	auxNextR := math32.Sqrt(math32.Pow(r, 2) - math32.Pow(auxNextRHeight, 2))
	bottom = getCircleVertices(0, r-auxHeight, 0, auxR, numVertex)
	bottom[0] = mgl32.Vec3{0, 0, 0}
	top = getCircleVertices(0, 2*r-sliceHeight, 0, auxR, numVertex)
	top[0] = mgl32.Vec3{0, 2 * r, 0}
	for i := 1; i < slices; i++ {
		currentCicle := getCircleVertices(0, r-auxHeight, 0, auxR, numVertex)
		nextCicle := getCircleVertices(0, r-auxNextRHeight, 0, auxNextR, numVertex)
		for i := 1; i < len(currentCicle); i++ {
			sphere = append(sphere, currentCicle[i])
			sphere = append(sphere, nextCicle[i])

		}
		auxHeight = auxNextRHeight
		auxR = auxNextR
		auxNextRHeight = auxHeight - sliceHeight
		auxNextR = math32.Sqrt(math32.Pow(r, 2) - math32.Pow(auxNextRHeight, 2))
	}

	semiSphereLen := len(sphere)
	for i := semiSphereLen - 2; i > 0; i-- {
		sphere = append(sphere, mgl32.Vec3{sphere[i].X(), -sphere[i].Y() + 2*r, sphere[i].Z()})
	}

	return
}

func getCubeVertices(x float32) []mgl32.Vec3 {
	var vertices = []mgl32.Vec3{
		{-x / 2, x, x / 2},   // Front-top-left
		{x / 2, x, x / 2},    // Front-top-right
		{-x / 2, 0, x / 2},   // Front-bottom-left
		{x / 2, 0, x / 2},    // Front-bottom-right
		{x / 2, 0, -x / 2},   // Back-bottom-right
		{x / 2, x, x / 2},    // Front-top-right
		{x / 2, x, -x / 2},   // Back-top-right
		{-x / 2, x, x / 2},   // Front-top-left
		{-x / 2, x, -x / 2},  // Back-top-left
		{-x / 2, -0, x / 2},  // Front-bottom-left
		{-x / 2, -0, -x / 2}, // Back-bottom-left
		{x / 2, -0, -x / 2},  // Back-bottom-right
		{-x / 2, x, -x / 2},  // Back-top-left
		{x / 2, x, -x / 2},   // Back-top-right
	}
	return vertices
}
func compileShader(src string, sType uint32) (uint32, error) {
	shader := gl.CreateShader(sType)
	glSrc, freeFn := gl.Strs(src + "\x00")
	defer freeFn()
	gl.ShaderSource(shader, 1, glSrc, nil)
	gl.CompileShader(shader)
	err := getGlError(shader, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog,
		"SHADER::COMPILE_FAILURE::")
	if err != nil {
		return 0, err
	}
	return shader, nil
}

type getObjIv func(uint32, uint32, *int32)
type getObjInfoLog func(uint32, int32, *int32, *uint8)

func getGlError(glHandle uint32, checkTrueParam uint32, getObjIvFn getObjIv,
	getObjInfoLogFn getObjInfoLog, failMsg string) error {

	var success int32
	getObjIvFn(glHandle, checkTrueParam, &success)

	if success == gl.FALSE {
		var logLength int32
		getObjIvFn(glHandle, gl.INFO_LOG_LENGTH, &logLength)

		log := gl.Str(strings.Repeat("\x00", int(logLength)))
		getObjInfoLogFn(glHandle, logLength, nil, log)

		return fmt.Errorf("%s: %s", failMsg, gl.GoStr(log))
	}

	return nil
}

/*
 * Creates the Vertex Array Object for a triangle.
 * indices is leftover from earlier samples and not used here.
 */
func createVAO(vertices []mgl32.Vec3) uint32 {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	var EBO uint32
	gl.GenBuffers(1, &EBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4*3, gl.Ptr(vertices), gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	var stride int32 = 3 * 4
	var offset int = 0

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(0)
	offset += 3 * 4

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO
}

func programLoop(window *glfw.Window) error {

	// the linked shader program determines how the data will be rendered
	vertShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertShader)
	gl.AttachShader(program, fragShader)
	gl.LinkProgram(program)
	err = getGlError(program, gl.LINK_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog,
		"PROGRAM::LINKING_FAILURE")
	if err != nil {
		return err
	}
	defer gl.DeleteShader(program)

	// ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.UseProgram(program)

	// creates camara
	eye, center, up := mgl32.Vec3{0, 0.5, 5}, mgl32.Vec3{0, 2, 0}, mgl32.Vec3{0, 1, 0}
	camera := mgl32.LookAtV(eye, center, up)
	cameraModel := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])

	// creates perspective
	//projectTransform := mgl32.Ortho(-5, 5, -5, 5, 0.1, 100)

	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width/height), 0.1, 100.0)
	gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("project\x00")), 1, false, &projectTransform[0])

	// light
	gl.Uniform3f(gl.GetUniformLocation(program, gl.Str("lightColor\x00")), 1, 1, 1)

	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	im := NewInputManager()
	window.SetKeyCallback(im.keyCallback)

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("world\x00"))
	colorModel := gl.GetUniformLocation(program, gl.Str("objectColor\x00"))

	cubeVertices := getCubeVertices(1)
	cubeVAO := createVAO(cubeVertices)

	sphereVertices, sphereTopVertices, sphereBottomVertices := getSphereVertices(1, 16, 4)
	sphereVAO, sphereTopVAO, sphereBottomVAO := createVAO(sphereVertices), createVAO(sphereTopVertices), createVAO(sphereBottomVertices)

	sideVertices, topVertices, bottomVertices := getCylinderVertices(1, 0.1, 5)
	sideVAO, topVAO, bottomVAO := createVAO(sideVertices), createVAO(topVertices), createVAO(bottomVertices)

	sideInVerticesPipe, sideOutVerticesPipe, topVerticesPipe, bottomVerticesPipe := getPipeVertices(0.75, 0.4, 0.5, 16)
	sideInVAOPipe, sideOutVAOPipe, topVAOPipe, bottomVAOPipe := createVAO(sideInVerticesPipe), createVAO(sideOutVerticesPipe), createVAO(topVerticesPipe), createVAO(bottomVerticesPipe)

	planeVertices := getPlaneVertices(10, 10, 1)
	planeVAO := createVAO(planeVertices)

	var treePositions = []mgl32.Vec3{
		{-2.5, 0, -0.5},
		{-2, 0, 1.5},
		{-2, 0, -3},
		{-1, 0, -1},
		{-1, 0, 1},
		{1, 0, -1},
		{2, 0, -1.5},
	}

	var r = float32(5)
	var angle = float32(0)
	var colorNum = 0
	var changeColor = true
	color, _, changeColor := getColorIn(im, colorNum, changeColor)

	for !window.ShouldClose() {

		time := float32(glfw.GetTime())

		//Camera
		angle = angle + getAngleIn(im)
		eye = mgl32.Vec3{r * math32.Cos(angle), 0.5, r * math32.Sin(angle)}
		camera = mgl32.LookAtV(eye, center, up)
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])

		// update events
		window.SwapBuffers()
		glfw.PollEvents()

		// background color
		gl.ClearColor(0, 0.27, 0.7, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT) // depth buffer needed for DEPTH_TEST

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
		wellTranslate := mgl32.Translate3D(1, 0, 1)

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

		//tree
		scale1 := 1 - math32.Abs(math32.Sin(time/1.5))*0.02
		scale2 := 1 - math32.Abs(math32.Cos(time/1.5))*0.02
		for _, pos := range treePositions {
			treeTranslate := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())

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
			color, colorNum, changeColor = getColorIn(im, colorNum, changeColor)
			gl.Uniform3f(colorModel, color[0].X(), color[0].Y(), color[0].Z())
			gl.BindVertexArray(cubeVAO)
			worldTranslate := treeTranslate.Mul4(mgl32.Scale3D(1, 1*scale1, 1)).Mul4(mgl32.Translate3D(0, 1.25, -0.5))
			gl.UniformMatrix4fv(modelUniform, 1, false, &worldTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

			gl.Uniform3f(colorModel, color[1].X(), color[1].Y(), color[1].Z())
			worldTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.75, 0.75*scale1, 0.75)).Mul4(mgl32.Translate3D(0, 1.2, 0.4))
			gl.UniformMatrix4fv(modelUniform, 1, false, &worldTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

			gl.Uniform3f(colorModel, color[2].X(), color[2].Y(), color[2].Z())
			worldTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.8, 0.8*scale2, 0.8)).Mul4(mgl32.Translate3D(-0.5, 1.5, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &worldTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
		}

		gl.Uniform3f(colorModel, 0.4, 0.6, 0)
		gl.BindVertexArray(planeVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(planeVertices)))

		gl.BindVertexArray(0)
		// end of draw loop
	}

	return nil
}

func main() {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to inifitialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 0)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "OpenGL", nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow (go function bindings)
	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialize gl bindings:", err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	err = programLoop(window)
	if err != nil {
		log.Fatalln(err)
	}
}
