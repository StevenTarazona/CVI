package ge

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

//CreateVAO ...
func CreateVAO(vertices, tCoords []float32, indices []uint32) uint32 {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	var VBO uint32
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	if len(tCoords) > 0 {
		var TBO uint32
		gl.GenBuffers(1, &TBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, TBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(tCoords)*4, gl.Ptr(tCoords), gl.STATIC_DRAW)
		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(1)
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	}
	if len(indices) > 0 {
		var EBO uint32
		gl.GenBuffers(1, &EBO)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	}
	gl.BindVertexArray(0)

	return VAO
}

//Mul defines multiplication of 2 vert3
func Mul(v1 mgl32.Vec3, v2 mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{v1.X() * v2.X(), v1.Y() * v2.Y(), v1.Z() * v2.Z()}
}

func Mul2(v1 mgl32.Vec2, v2 mgl32.Vec2) mgl32.Vec2 {
	return mgl32.Vec2{v1.X() * v2.X(), v1.Y() * v2.Y()}
}

//Translate defines sum of vert3 array and a ver3
func Translate(vertices []mgl32.Vec3, vertex mgl32.Vec3) (translated []mgl32.Vec3) {
	for _, ver := range vertices {
		translated = append(translated, ver.Add(vertex))
	}
	return
}

//Transform defines multiplication of vert3 array and a ver3
func Transform(vertices []mgl32.Vec3, vertex mgl32.Vec3) (translated []mgl32.Vec3) {
	for _, ver := range vertices {
		translated = append(translated, Mul(ver, vertex))
	}
	return
}

func Transform2(vertices []mgl32.Vec2, vertex mgl32.Vec2) (translated []mgl32.Vec2) {
	for _, ver := range vertices {
		translated = append(translated, Mul2(ver, vertex))
	}
	return
}
