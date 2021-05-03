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
	title  = "Particles"
)

var (
	lightPositions = []mgl32.Vec3{
		{0, .5, 3},
		{0, .5, 0},
	}
	lightColors = []mgl32.Vec3{
		{1, 0, 1},
		{1, 1, 0.7},
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

func createParticleVAO(points []float32) (uint32, uint32) {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.BindVertexArray(VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(points)*4, gl.Ptr(points), gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	var stride int32 = 3*4 + 4*4
	var offset int = 0

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(1)
	offset += 3 * 4
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.BindVertexArray(0)

	return VAO, VBO
}

func pointLightsUL(program *gfx.Program) [][]int32 {
	uniformLocations := [][]int32{}
	for i := 0; i < len(lightPositions); i++ {
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

func programLoop(window *win.Window) error {

	// Shaders and textures
	vS, err := gfx.NewShaderFromFile("shaders/phong_ml.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fS, err := gfx.NewShaderFromFile("shaders/phong_ml.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program, err := gfx.NewProgram(vS, fS)
	if err != nil {
		return err
	}
	defer program.Delete()

	sourceVS, err := gfx.NewShaderFromFile("shaders/source.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	sourceFS, err := gfx.NewShaderFromFile("shaders/source.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	// special shader program so that lights themselves are not affected by lighting
	sourceProgram, err := gfx.NewProgram(sourceVS, sourceFS)
	if err != nil {
		return err
	}
	defer sourceProgram.Delete()

	particlesVS, err := gfx.NewShaderFromFile("shaders/particles.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	particlesFS, err := gfx.NewShaderFromFile("shaders/particles.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}
	particlesGS, err := gfx.NewShaderFromFile("shaders/particles.geom", gl.GEOMETRY_SHADER)
	if err != nil {
		return err
	}

	particlesProgram, err := gfx.NewProgram(particlesVS, particlesFS, particlesGS)
	if err != nil {
		return err
	}
	defer particlesProgram.Delete()

	// Ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)

	// Base model
	model := mgl32.Ident4()

	// Uniform
	modelUL := program.GetUniformLocation("model")
	viewUL := program.GetUniformLocation("view")
	projectUL := program.GetUniformLocation("projection")
	objectColorUL := program.GetUniformLocation("objectColor")
	viewPosUL := program.GetUniformLocation("viewPos")
	numLightsUL := program.GetUniformLocation("numLights")
	//texture0UL := program.GetUniformLocation("texSampler0")
	//texture1UL := program.GetUniformLocation("texSampler1")
	pointLightsUL := pointLightsUL(program)

	sourceModelUL := sourceProgram.GetUniformLocation("model")
	sourceViewUL := sourceProgram.GetUniformLocation("view")
	sourceProjectUL := sourceProgram.GetUniformLocation("projection")
	sourceObjectColorUL := sourceProgram.GetUniformLocation("objectColor")
	//sourceTextureUL := sourceProgram.GetUniformLocation("texSampler")

	particlesModelUL := particlesProgram.GetUniformLocation("model")
	particlesViewUL := particlesProgram.GetUniformLocation("view")
	particlesProjectUL := particlesProgram.GetUniformLocation("projection")
	particlesSizeUL := particlesProgram.GetUniformLocation("particle_size")
	particlesTextureUL := particlesProgram.GetUniformLocation("tex0")

	// creates camara
	eye := mgl32.Vec3{0, 1.5, 5}
	center := mgl32.Vec3{0, 0, 0}
	camera := mgl32.LookAtV(eye, center, mgl32.Vec3{0, 1, 0})

	// creates perspective
	fov := float32(60.0)
	projection := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100)

	// Textures
	particlTexture, err := gfx.NewTextureFromFile("textures/particle.png",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// Settings
	backgroundColor := mgl32.Vec3{0, 0, 0}
	objectColor := mgl32.Vec3{1.0, 1.0, 1.0}
	polygonMode := false

	particle_size := 0.3
	numParticles := 95
	particlesColor := &lightColors[0]
	particlesPos := &lightPositions[0]
	particlesVel := mgl32.Vec3{1, 0, 0}
	particlesMinLife := float32(0.5)
	particlesMaxLife := float32(3)
	particlesAmplitude := float32(.5)

	particles := NewParticles(numParticles, *particlesColor, *particlesPos, particlesVel, particlesMinLife, particlesMaxLife, particlesAmplitude)

	if polygonMode {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	}

	// Geometry
	particleVAO, particleVBO := createParticleVAO(particles.points)

	xLightSegments, yLighteSegments := 30, 30
	lightVAO := createVAO(Sphere(xLightSegments, yLighteSegments))

	xPlaneSegments, yPlaneSegments := 15, 15
	planeVAO := createVAO(Square(xPlaneSegments, yPlaneSegments, 1))

	// Scene and animation always needs to be after the model and buffers initialization
	animationCtl := gfx.NewAnimationManager()
	animationCtl.AddContunuousAnimation(func() {
		r := float32(3)
		*particlesPos = mgl32.Vec3{r * math32.Cos(float32(animationCtl.GetAngle())), particlesPos.Y(), r * math32.Sin(float32(animationCtl.GetAngle()))}
		particlesVel = mgl32.Vec3{r * -math32.Sin(float32(animationCtl.GetAngle())), particlesPos.Y(), 0 * math32.Cos(float32(animationCtl.GetAngle()))}.Normalize()

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
		particles.Update(float32(animationCtl.GetElapsed()), *particlesPos, particlesVel)

		// You shall draw here
		program.Use()

		gl.UniformMatrix4fv(viewUL, 1, false, &camera[0])
		gl.UniformMatrix4fv(projectUL, 1, false, &projection[0])

		gl.Uniform3fv(viewPosUL, 1, &eye[0])
		gl.Uniform3f(objectColorUL, objectColor.X(), objectColor.Y(), objectColor.Z())
		gl.Uniform1i(numLightsUL, int32(len(lightPositions)))

		//Lights
		for index, pointLightPosition := range lightPositions {
			gl.Uniform3fv(pointLightsUL[index][0], 1, &pointLightPosition[0])
			gl.Uniform3f(pointLightsUL[index][1], 0.1, 0.1, 0.1)
			gl.Uniform3f(pointLightsUL[index][2], 0.8, 0.8, 0.8)
			gl.Uniform3f(pointLightsUL[index][3], 0.8, 0.8, 0.8)
			gl.Uniform1f(pointLightsUL[index][4], 1.)
			gl.Uniform1f(pointLightsUL[index][5], 0.09)
			gl.Uniform1f(pointLightsUL[index][6], 0.032)
			gl.Uniform3f(pointLightsUL[index][7], lightColors[index].X(), lightColors[index].Y(), lightColors[index].Z())
		}

		// render models

		//Plane
		gl.BindVertexArray(planeVAO)
		gl.UniformMatrix4fv(modelUL, 1, false, &model[0])
		gl.DrawElements(gl.TRIANGLES, int32(xPlaneSegments*yPlaneSegments)*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		gl.BindVertexArray(0)

		//Source program
		sourceProgram.Use()
		gl.UniformMatrix4fv(sourceProjectUL, 1, false, &projection[0])
		gl.UniformMatrix4fv(sourceViewUL, 1, false, &camera[0])

		//Light objects
		gl.BindVertexArray(lightVAO)
		for i, lp := range lightPositions {
			lightTransform := model
			if i > 0 {
				lightTransform = model.Mul4(mgl32.Translate3D(lp.Elem())).Mul4(mgl32.Scale3D(0.2, 0.2, 0.2))
			} else {
				lightTransform = model.Mul4(mgl32.Translate3D(lp.Elem())).Mul4(mgl32.Scale3D(0.05, 0.05, 0.05))
			}
			gl.Uniform3f(sourceObjectColorUL, lightColors[i].X(), lightColors[i].Y(), lightColors[i].Z())
			gl.UniformMatrix4fv(sourceModelUL, 1, false, &lightTransform[0])
			gl.DrawElements(gl.TRIANGLES, int32(xLightSegments*yLighteSegments)*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		}
		gl.BindVertexArray(0)

		//Particles
		particlesProgram.Use()
		gl.UniformMatrix4fv(particlesProjectUL, 1, false, &projection[0])
		gl.UniformMatrix4fv(particlesViewUL, 1, false, &camera[0])
		gl.UniformMatrix4fv(particlesModelUL, 1, false, &model[0])

		gl.BindVertexArray(particleVAO)
		gl.BindBuffer(gl.ARRAY_BUFFER, particleVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(particles.points)*4, gl.Ptr(particles.points), gl.STATIC_DRAW)

		gl.Uniform1f(particlesSizeUL, float32(particle_size))
		particlTexture.Bind(gl.TEXTURE0)
		particlTexture.SetUniform(particlesTextureUL)
		gl.Disable(gl.DEPTH_TEST)
		gl.DrawArrays(gl.POINTS, 0, int32(numParticles))
		gl.Enable(gl.DEPTH_TEST)
		particlTexture.UnBind()
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
