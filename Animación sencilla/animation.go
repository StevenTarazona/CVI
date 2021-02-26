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

func vect2Vertex3(points []mgl32.Vec3) (vertices []ge.Vertex3) {
	for _, element := range points {
		vertices = append(vertices, ge.Vertex3{element.X(), element.Y(), element.Z()})
	}
	return
}

func vertex32Vect(points []ge.Vertex3) (vertices []mgl32.Vec3) {
	for _, element := range points {
		vertices = append(vertices, mgl32.Vec3{element.X, element.Y, element.Z})
	}
	return
}

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

	// creates camara
	camera := mgl32.LookAtV(mgl32.Vec3{10, 10, 10}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("camera\x00")), 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("project\x00")), 1, false, &projectTransform[0])

	// light
	gl.Uniform3f(gl.GetUniformLocation(program, gl.Str("lightColor\x00")), 1, 1, 1)

	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("world\x00"))
	colorModel := gl.GetUniformLocation(program, gl.Str("objectColor\x00"))

	sphereVertices, sphereTopVertices, sphereBottomVertices := ge.GetSphereVertices3(1, 18)
	sphereVAO, sphereTopVAO, sphereBottomVAO := ge.CreateVAO(sphereVertices), ge.CreateVAO(sphereTopVertices), ge.CreateVAO(sphereBottomVertices)

	planeVertices := ge.GetPlaneVertices3(10, 10, 1)
	planeVAO := ge.CreateVAO(planeVertices)

	pathAux0 := []ge.Vertex3{}
	for i := 1; i < 5; i++ {
		circle := ge.GetCircleVertices3(math32.Pow(float32(i), 2), 10)
		l := len(circle) - 1
		pathAux0 = append(pathAux0, ge.Translate(circle[1:l], ge.Vertex3{X: float32(0), Y: float32(i), Z: float32(0)})...)
	}

	pathAux := vertex32Vect(pathAux0)

	path := mgl32.MakeBezierCurve3D(100, pathAux)
	vao := ge.CreateVAO(vect2Vertex3(path))
	sec := float32(0)
	pos := 0

	var t0Value float64 = glfw.GetTime()

	moonTranslate := mgl32.Translate3D(0, 0, 0)

	for !window.ShouldClose() {

		ge.NewFrame(window, mgl32.Vec4{0, 0, 0, 1})

		// You shall draw here

		if currentTime := glfw.GetTime(); (currentTime - t0Value) >= .05 {
			t0Value = glfw.GetTime()

			sec++
			moonTranslate = mgl32.Translate3D(path[pos].Elem())
			pos = int(math32.Mod(sec, float32(len(path))))

		}
		gl.Uniform3f(colorModel, 1, 0, 1)
		gl.BindVertexArray(sphereVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &moonTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereVertices)))

		gl.BindVertexArray(sphereTopVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &moonTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereTopVertices)))

		gl.BindVertexArray(sphereBottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &moonTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereBottomVertices)))

		gl.Uniform3f(colorModel, 1, 1, 1)
		gl.BindVertexArray(planeVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(planeVertices)))

		gl.Uniform3f(colorModel, 0, 0, 1)
		gl.BindVertexArray(vao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
		gl.DrawArrays(gl.LINE_STRIP, 0, int32(len(path)))
		gl.BindVertexArray(0)

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
