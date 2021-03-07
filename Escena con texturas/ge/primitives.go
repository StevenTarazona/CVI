package ge

import (
	"git.maze.io/go/math32"
	"github.com/go-gl/mathgl/mgl32"
)

//GetCircleVertices3 ...
func GetCircleVertices3(r float32, vertices int) (unitCircleVertices []mgl32.Vec3) {
	var sectorStep = 2 * math32.Pi / float32(vertices)
	var sectorAngle float32 // radian
	unitCircleVertices = append(unitCircleVertices, mgl32.Vec3{0, 0, 0})
	for i := 0; i <= vertices; i++ {
		sectorAngle = float32(i) * sectorStep
		unitCircleVertices = append(unitCircleVertices, mgl32.Vec3{r * math32.Cos(sectorAngle), 0, r * math32.Sin(sectorAngle)})
	}
	return
}

//GetRingVerticies3 ...
func GetRingVerticies3(rIn float32, rOut float32, vertices int) (ring []mgl32.Vec3) {
	in := GetCircleVertices3(rIn, vertices)
	out := GetCircleVertices3(rOut, vertices)
	for i := 1; i <= vertices+1; i++ {
		ring = append(ring, out[i])
		ring = append(ring, in[i])
	}
	return
}

//GetCylinderVertices3 ...
func GetCylinderVertices3(h float32, rBottom float32, rTop float32, vertices int) (side, top, bottom []mgl32.Vec3) {
	var slices int
	var sign float32
	if slices = int(h / (math32.Sin(2*math32.Pi/float32(vertices)) * math32.Max(rBottom, rTop))); slices == 0 {
		slices = 1
	}
	auxR := rBottom
	if rBottom == rTop {
		sign = 1
	} else {
		sign = (rBottom - rTop) / math32.Abs(rBottom-rTop)
	}
	tan := sign * (rBottom - rTop) / h
	circle := GetCircleVertices3(auxR, vertices)
	nextCircle := circle
	bottom = circle
	for slice := 1; slice <= slices; slice++ {
		auxH := float32(slice) * (h / float32(slices))
		auxR = sign*(h-auxH)*tan + rTop
		circle = nextCircle
		nextCircle = Translate(GetCircleVertices3(auxR, vertices), mgl32.Vec3{0, auxH, 0})
		for i := 1; i <= vertices+1; i++ {
			side = append(side, circle[i])
			side = append(side, nextCircle[i])
		}
	}
	top = Translate(GetCircleVertices3(rTop, vertices), mgl32.Vec3{0, h, 0})
	return
}

//GetPipeVertices3 ...
func GetPipeVertices3(h float32, rIn float32, rOut float32, vertices int) (sideIn, sideOut, top, bottom []mgl32.Vec3) {
	var slices int
	if slices = int(h / (math32.Sin(2*math32.Pi/float32(vertices)) * rOut)); slices == 0 {
		slices = 1
	}

	bottomIn := GetCircleVertices3(rIn, vertices)
	bottomOut := GetCircleVertices3(rOut, vertices)

	for slice := 0; slice < slices; slice++ {
		for i := 1; i < len(bottomOut); i++ {
			sideIn = append(sideIn, bottomIn[i].Add(mgl32.Vec3{0, float32(slice) * (h / float32(slices)), 0}))
			sideIn = append(sideIn, bottomIn[i].Add(mgl32.Vec3{0, float32(slice+1) * (h / float32(slices)), 0}))
			sideOut = append(sideOut, bottomOut[i].Add(mgl32.Vec3{0, float32(slice) * (h / float32(slices)), 0}))
			sideOut = append(sideOut, bottomOut[i].Add(mgl32.Vec3{0, float32(slice+1) * (h / float32(slices)), 0}))
		}
	}
	bottom = GetRingVerticies3(rIn, rOut, vertices)
	top = Translate(bottom, mgl32.Vec3{0, h, 0})
	return
}

//GetSemiSphereVertices3 ...
func GetSemiSphereVertices3(r float32, vertices int) (side, top, bottom []mgl32.Vec3) {
	auxR := r
	auxH := float32(0)
	circle := GetCircleVertices3(r, vertices)
	nextCircle := circle
	bottom = circle
	for true {
		auxH += nextCircle[2].Z()
		if auxH >= r {
			break
		}
		auxR = math32.Sqrt(math32.Pow(r, 2) - math32.Pow(auxH, 2))
		circle = nextCircle
		nextCircle = Translate(GetCircleVertices3(auxR, vertices), mgl32.Vec3{0, auxH, 0})
		for i := 1; i <= vertices+1; i++ {
			side = append(side, circle[i])
			side = append(side, nextCircle[i])
		}
	}
	top = nextCircle
	top[0] = mgl32.Vec3{0, r, 0}
	return
}

//GetSphereVertices3 ...
func GetSphereVertices3(r float32, numVertex int) (side, top, bottom []mgl32.Vec3) {
	semiSphere, top, _ := GetSemiSphereVertices3(r, numVertex)
	for i := len(semiSphere) - 1; i >= 0; i-- {
		side = append(side, Mul(semiSphere[i], mgl32.Vec3{1, -1, 1}))
	}
	side = append(side, semiSphere...)
	for _, v := range top {
		bottom = append(bottom, Mul(v, mgl32.Vec3{1, -1, 1}))
	}
	side = Translate(side, mgl32.Vec3{0, r, 0})
	top = Translate(top, mgl32.Vec3{0, r, 0})
	bottom = Translate(bottom, mgl32.Vec3{0, r, 0})
	return
}

//GetCapsuleVertices3 ...
func GetCapsuleVertices3(h float32, rBottom float32, rTop float32, vertices int) (side, top, bottom []mgl32.Vec3) {
	sideTemp, _, _ := GetCylinderVertices3(h-rBottom-rTop, rBottom, rTop, vertices)
	sideTemp = Translate(sideTemp, mgl32.Vec3{0, rBottom, 0})
	topSide, top, _ := GetSemiSphereVertices3(rTop, vertices)
	topSide = Translate(topSide, mgl32.Vec3{0, h - rTop, 0})
	top = Translate(top, mgl32.Vec3{0, h - rTop, 0})
	bottomSide, bottomTemp, _ := GetSemiSphereVertices3(rBottom, vertices)
	for _, v := range bottomTemp {
		bottom = append(bottom, mgl32.Vec3{v.X(), rBottom - v.Y(), v.Z()})
	}
	for i := len(bottomSide) - 1; i >= 0; i-- {
		side = append(side, mgl32.Vec3{bottomSide[i].X(), rBottom - bottomSide[i].Y(), bottomSide[i].Z()})
	}
	side = append(side, sideTemp...)
	side = append(side, topSide...)
	return
}

//GetCubicHexahedronVertices3 ...
func GetCubicHexahedronVertices3(X, Y, Z float32) []mgl32.Vec3 {
	var vertices = []mgl32.Vec3{
		{-X / 2, 0, -Z / 2}, {-X / 2, Y, -Z / 2}, {X / 2, 0, -Z / 2},
		{X / 2, 0, -Z / 2}, {-X / 2, Y, -Z / 2}, {X / 2, Y, -Z / 2},
		{X / 2, 0, -Z / 2}, {X / 2, Y, -Z / 2}, {X / 2, Y, Z / 2},
		{X / 2, Y, Z / 2}, {X / 2, Y, -Z / 2}, {-X / 2, Y, -Z / 2},
		{X / 2, Y, Z / 2}, {-X / 2, Y, -Z / 2}, {-X / 2, Y, Z / 2},
		{-X / 2, Y, Z / 2}, {-X / 2, Y, -Z / 2}, {-X / 2, 0, Z / 2},
		{-X / 2, Y, Z / 2}, {-X / 2, 0, Z / 2}, {X / 2, Y, Z / 2},
		{X / 2, Y, Z / 2}, {-X / 2, 0, Z / 2}, {X / 2, 0, Z / 2},
		{X / 2, Y, Z / 2}, {X / 2, 0, Z / 2}, {X / 2, 0, -Z / 2},
		{X / 2, 0, -Z / 2}, {X / 2, 0, Z / 2}, {-X / 2, 0, Z / 2},
		{X / 2, 0, -Z / 2}, {-X / 2, 0, Z / 2}, {-X / 2, 0, -Z / 2},
		{-X / 2, 0, -Z / 2}, {-X / 2, 0, Z / 2}, {-X / 2, Y, -Z / 2},
	}
	return vertices
}

//GetCubicHexahedronTextureCoords ...
func GetCubicHexahedronTextureCoords(X, Y, Z float32) []mgl32.Vec2 {
	var vertices = []mgl32.Vec2{
		{0, 0}, {0, Y}, {X, 0},
		{X, 0}, {0, Y}, {X, Y},
		{0, 0}, {Y, 0}, {Y, Z},
		{X, Z}, {X, 0}, {0, 0},
		{X, Z}, {0, 0}, {0, Z},
		{Y, Z}, {Y, 0}, {0, Z},
		{0, Y}, {0, 0}, {X, Y},
		{X, Y}, {0, 0}, {X, 0},
		{Y, Z}, {0, Z}, {0, 0},
		{X, 0}, {X, Z}, {0, Z},
		{X, 0}, {0, Z}, {0, 0},
		{0, 0}, {0, Z}, {Y, 0},
	}
	return vertices
}

//GetPlaneVertices3 ...
func GetPlaneVertices3(h int, w int, l int) []mgl32.Vec3 {
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

//GetPlaneTextureCoords ...
func GetPlaneTextureCoords(h int, w int, l int) (vertices []mgl32.Vec2) {
	planeVertices := GetPlaneVertices3(h, w, l)
	for _, v := range planeVertices {
		//x := (i / 2) % (w + 1)
		//y := (i / (2 * (h + 1))) + i%2
		vertices = append(vertices, mgl32.Vec2{(v.X() + float32(w)/2) / float32(w), (v.Z() + float32(h)/2) / float32(h)})
	}
	return
}
