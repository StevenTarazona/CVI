package main

import (
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	ge "github.com/kaitsubaka/glae" // libreria propia implementada desde 0, https://github.com/kaitsubaka/glae
)

const (
	width              = 1080
	height             = 720
	windowName         = "Movement test"
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
	points    = getPointsFromFile("WBDSascii/WBDS30walkO06Cmkr.txt")
	positions = []mgl32.Vec3{}
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getPointsFromFile(file string) (points [][]mgl32.Vec3) {
	content, err := ioutil.ReadFile(file)
	check(err)

	r := csv.NewReader(strings.NewReader(string(content)))
	r.Comma = '\t'

	index := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)
		if index > 0 {
			for i := 1; i < len(record)-2; i += 3 {
				xValue, err := strconv.ParseFloat(record[i], 32)
				check(err)
				yValue, err := strconv.ParseFloat(record[i+1], 32)
				check(err)
				zValue, err := strconv.ParseFloat(record[i+2], 32)
				check(err)
				if index == 1 {
					points = append(points, []mgl32.Vec3{{float32(xValue) / 1000, float32(yValue) / 1000, float32(zValue) / 1000}})
				} else {
					points[i/3] = append(points[i/3], mgl32.Vec3{float32(xValue) / 1000, float32(yValue) / 1000, float32(zValue) / 1000})
				}
			}
		}
		index++
	}
	return
}

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
	//model := mgl32.Ident4()

	// Uniform locations
	WorldUniformLocation := gl.GetUniformLocation(program, gl.Str("world\x00"))
	colorUniformLocation := gl.GetUniformLocation(program, gl.Str("objectColor\x00"))
	lightColorUniformLocation := gl.GetUniformLocation(program, gl.Str("lightColor\x00"))
	cameraUniformLocation := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	projectUniformLocation := gl.GetUniformLocation(program, gl.Str("project\x00"))

	// creates camara
	camera := mgl32.LookAtV(mgl32.Vec3{1, 1, 3}, mgl32.Vec3{1, 1, 0}, mgl32.Vec3{0, 1, 0})
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
	sphereVertices, sphereVerticesT, sphereVerticesB := ge.GetSphereVertices3(.025, 8)
	sphereVAO, sphereVAOT, sphereVAOB := ge.CreateVAO(sphereVertices), ge.CreateVAO(sphereVerticesT), ge.CreateVAO(sphereVerticesB)

	// Scene and animation
	angle := 0.0
	previousTime := glfw.GetTime()
	totalElapsed := float64(0)
	movementControlCount := 0

	movementTimes := []float64{}
	movementFunctions := []func(t float32){}

	// Animations
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		positions = []mgl32.Vec3{}
		for _, v := range points {
			positions = append(positions, v[int(float32(len(v)-1)*t)])
		}
	}), append(movementTimes, 5)

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

		for _, pos := range positions {
			moonTranslate := mgl32.Translate3D(pos.Elem())
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
		}
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
