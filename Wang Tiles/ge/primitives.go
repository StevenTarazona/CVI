package ge

import (
	"math/rand"
	"time"

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

//GetSquare ...
func GetSquare(hTiles int, vTiles int, tileLengths float32) (vertices []mgl32.Vec3, tCoords []mgl32.Vec2, indices []uint32) {
	vOfset := (float32(vTiles) * tileLengths) / 2
	hOfset := (float32(hTiles) * tileLengths) / 2
	for v := 0; v <= vTiles; v++ {
		for h := 0; h <= hTiles; h++ {
			vertices = append(vertices, mgl32.Vec3{(float32(h) * tileLengths) - hOfset, 0, (float32(v) * tileLengths) - vOfset})
			tCoords = append(tCoords, mgl32.Vec2{(float32(h) * (1 / float32(hTiles))), (float32(v) * (1 / float32(vTiles)))})
			if h < hTiles && v < vTiles {
				indices = append(indices, []uint32{uint32(h + (hTiles+1)*v), uint32(h + (hTiles+1)*v + (hTiles + 1)), uint32(h + (hTiles+1)*v + 1), uint32(h + (hTiles+1)*v + (hTiles + 1) + 1), uint32(h + (hTiles+1)*v + 1), uint32(h + (hTiles+1)*v + (hTiles + 1))}...)
			}
		}
	}
	return
}

//GetSquareRepeat ...
func GetSquareRepeat(hTiles int, vTiles int, tileLengths float32) (vertices []mgl32.Vec3, tCoords []mgl32.Vec2, indices []uint32) {
	vOfset := (float32(vTiles) * tileLengths) / 2
	hOfset := (float32(hTiles) * tileLengths) / 2
	for v := 0; v < vTiles; v++ {
		for h := 0; h < hTiles; h++ {
			vertices = append(vertices, []mgl32.Vec3{
				{(float32(h) * tileLengths) - hOfset, 0, (float32(v) * tileLengths) - vOfset},
				{(float32(h) * tileLengths) - hOfset, 0, (float32(v+1) * tileLengths) - vOfset},
				{(float32(h+1) * tileLengths) - hOfset, 0, (float32(v) * tileLengths) - vOfset},
				{(float32(h+1) * tileLengths) - hOfset, 0, (float32(v+1) * tileLengths) - vOfset},
			}...)
			tCoords = append(tCoords, []mgl32.Vec2{{0, 0}, {0, 1}, {1, 0}, {1, 1}}...)
			indices = append(indices, []uint32{uint32(4*h + hTiles*4*v), uint32(4*h + hTiles*4*v + 1), uint32(4*h + hTiles*4*v + 2), uint32(4*h + hTiles*4*v + 3), uint32(4*h + hTiles*4*v + 2), uint32(4*h + hTiles*4*v + 1)}...)
		}
	}
	return
}

//GetSquareWangTiles ...
func GetSquareWangTiles(hTiles int, vTiles int, tileLengths float32, tileCords [][]mgl32.Vec2, adjacencyList [][][]int) (vertices []mgl32.Vec3, tCoords []mgl32.Vec2, indices []uint32) {
	rand.Seed(time.Now().UnixNano())
	vOfset := (float32(vTiles) * tileLengths) / 2
	hOfset := (float32(hTiles) * tileLengths) / 2

	possibilities := [][][]int{}
	for v := 0; v < vTiles; v++ {
		possibilities = append(possibilities, [][]int{})
		for h := 0; h < hTiles; h++ {
			possibilities[v] = append(possibilities[v], []int{})
		}
	}

	all := []int{}
	for i := range tileCords {
		all = append(all, i)
	}
	possibilities[0][0] = all

	for v := 0; v < vTiles; v++ {
		for h := 0; h < hTiles; h++ {
			vertices = append(vertices, []mgl32.Vec3{
				{(float32(h) * tileLengths) - hOfset, 0, (float32(v) * tileLengths) - vOfset},
				{(float32(h) * tileLengths) - hOfset, 0, (float32(v+1) * tileLengths) - vOfset},
				{(float32(h+1) * tileLengths) - hOfset, 0, (float32(v) * tileLengths) - vOfset},
				{(float32(h+1) * tileLengths) - hOfset, 0, (float32(v+1) * tileLengths) - vOfset},
			}...)
			firstIndex := 4 * (h + v*hTiles)
			indices = append(indices, []uint32{
				uint32(firstIndex),
				uint32(firstIndex + 1),
				uint32(firstIndex + 2),
				uint32(firstIndex + 3),
				uint32(firstIndex + 2),
				uint32(firstIndex + 1),
			}...)

			currentPossibilities := possibilities[h][v]
			currentTile := currentPossibilities[rand.Intn(len(currentPossibilities))]
			tCoords = append(tCoords, tileCords[currentTile]...)

			if (h < hTiles-1 && v < vTiles-1) || (h == hTiles-1 && v < vTiles-1) {
				//Bottom [1]
				if len(possibilities[h][v+1]) == 0 {
					possibilities[h][v+1] = adjacencyList[currentTile][1]
				} else {
					possibilities[h][v+1] = andArray(possibilities[h+1][v], adjacencyList[currentTile][1])
				}
			}
			if (h < hTiles-1 && v < vTiles-1) || (v == vTiles-1 && h < hTiles-1) {
				//Right [0]
				if len(possibilities[h+1][v]) == 0 {
					possibilities[h+1][v] = adjacencyList[currentTile][0]
				} else {
					possibilities[h+1][v] = andArray(possibilities[h+1][v], adjacencyList[currentTile][0])
				}
			}
		}
	}
	return
}

func andArray(a1, a2 []int) (a []int) {
	for _, i := range a1 {
		for _, j := range a2 {
			if i == j {
				a = append(a, i)
			}
		}
	}
	if len(a) == 0 {
		return append(a1, a2...)
	}
	return
}
