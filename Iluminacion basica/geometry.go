package main

import (
	"git.maze.io/go/math32"
	"github.com/go-gl/mathgl/mgl32"
)

// DRAW WITH: gl.DrawElements(gl.TRIANGLES, X_SEGMENTS*Y_SEGMENTS*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
func Sphere(Y_SEGMENTS, X_SEGMENTS int) (vertices, normals []float32, indices []uint32) {

	for y := 0; y <= Y_SEGMENTS; y++ {
		for x := 0; x <= X_SEGMENTS; x++ {
			xSegment := float32(x) / float32(X_SEGMENTS)
			ySegment := float32(y) / float32(Y_SEGMENTS)

			xPos := float32(math32.Cos(xSegment*math32.Pi*2.0) * math32.Sin(ySegment*math32.Pi))
			yPos := float32(math32.Cos(ySegment * math32.Pi))
			zPos := float32(math32.Sin(xSegment*math32.Pi*2.0) * math32.Sin(ySegment*math32.Pi))

			vertices = append(vertices, xPos, yPos, zPos)
			xPos, yPos, zPos = mgl32.Vec3{xPos, yPos, zPos}.Normalize().Elem()
			normals = append(normals, xPos, yPos, zPos)
		}
	}
	for i := 0; i < Y_SEGMENTS; i++ {
		for j := 0; j < X_SEGMENTS; j++ {
			a1 := uint32(i*(X_SEGMENTS+1) + j)
			a2 := uint32((i+1)*(X_SEGMENTS+1) + j)
			a3 := uint32((i+1)*(X_SEGMENTS+1) + j + 1)
			b1 := uint32(i*(X_SEGMENTS+1) + j)
			b2 := uint32((i+1)*(X_SEGMENTS+1) + j + 1)
			b3 := uint32(i*(X_SEGMENTS+1) + j + 1)
			indices = append(indices, a1, a2, a3, b1, b2, b3)
		}
	}
	return
}
