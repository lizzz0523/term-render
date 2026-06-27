package model

import (
	"encoding/binary"
	"fmt"
	"math"

	"test-term/internal/geo"

	"github.com/qmuntal/gltf"
)

type Model struct {
	root   *bvhNode
	radius float64
}

type bvhNode struct {
	min, max  geo.Vec3
	left      *bvhNode
	right     *bvhNode
	triangles []triangle
}

type triangle struct {
	v0, v1, v2 geo.Vec3
	n0, n1, n2 geo.Vec3
}

func LoadGLB(path string) (*Model, error) {
	doc, err := gltf.Open(path)
	if err != nil {
		return nil, err
	}

	if len(doc.Meshes) == 0 {
		return nil, fmt.Errorf("no meshes found in GLB")
	}
	mesh := doc.Meshes[0]

	var positions []float64
	var normals []float64
	var indices []uint32

	for _, prim := range mesh.Primitives {
		posAcc := doc.Accessors[prim.Attributes["POSITION"]]
		positions = readFloats(doc, posAcc)

		if idx, ok := prim.Attributes["NORMAL"]; ok {
			nrmAcc := doc.Accessors[idx]
			normals = readFloats(doc, nrmAcc)
		}

		indices = readIndices(doc, doc.Accessors[*prim.Indices])
	}

	triangles := buildTriangles(positions, normals, indices)
	if len(triangles) == 0 {
		return nil, fmt.Errorf("no triangles generated")
	}

	center := computeCenter(triangles)
	centerTriangles(triangles, center)

	targetRadius := 2.0
	radius := computeRadius(triangles)
	scale := targetRadius / radius
	scaleTriangles(triangles, scale)

	root := buildBVH(triangles)

	return &Model{root: root, radius: targetRadius}, nil
}

func readFloats(doc *gltf.Document, acc *gltf.Accessor) []float64 {
	bv := doc.BufferViews[*acc.BufferView]
	bufData := doc.Buffers[bv.Buffer].Data
	start := bv.ByteOffset + acc.ByteOffset
	comps := acc.Type.Components()
	count := acc.Count
	byteSize := acc.ComponentType.ByteSize()

	result := make([]float64, count*comps)
	for i := range result {
		offset := start + i*byteSize
		result[i] = float64(math.Float32frombits(binary.LittleEndian.Uint32(bufData[offset:])))
	}
	return result
}

func readIndices(doc *gltf.Document, acc *gltf.Accessor) []uint32 {
	bv := doc.BufferViews[*acc.BufferView]
	bufData := doc.Buffers[bv.Buffer].Data
	start := bv.ByteOffset + acc.ByteOffset
	count := acc.Count
	byteSize := acc.ComponentType.ByteSize()

	result := make([]uint32, count)
	switch byteSize {
	case 2:
		for i := range result {
			result[i] = uint32(binary.LittleEndian.Uint16(bufData[start+i*2:]))
		}
	case 4:
		for i := range result {
			result[i] = binary.LittleEndian.Uint32(bufData[start+i*4:])
		}
	}
	return result
}

func buildTriangles(positions, normals []float64, indices []uint32) []triangle {
	tris := make([]triangle, 0, len(indices)/3)
	hasNormals := len(normals) == len(positions)

	for i := 0; i < len(indices); i += 3 {
		i0, i1, i2 := indices[i], indices[i+1], indices[i+2]

		v0 := geo.NewVec3(positions[i0*3], positions[i0*3+1], positions[i0*3+2])
		v1 := geo.NewVec3(positions[i1*3], positions[i1*3+1], positions[i1*3+2])
		v2 := geo.NewVec3(positions[i2*3], positions[i2*3+1], positions[i2*3+2])

		var n0, n1, n2 geo.Vec3
		if hasNormals {
			n0 = geo.NewVec3(normals[i0*3], normals[i0*3+1], normals[i0*3+2])
			n1 = geo.NewVec3(normals[i1*3], normals[i1*3+1], normals[i1*3+2])
			n2 = geo.NewVec3(normals[i2*3], normals[i2*3+1], normals[i2*3+2])
		} else {
			e1 := v1.Sub(v0)
			e2 := v2.Sub(v0)
			fn := e1.Cross(e2).Norm()
			n0, n1, n2 = fn, fn, fn
		}

		tris = append(tris, triangle{v0, v1, v2, n0, n1, n2})
	}
	return tris
}

func computeCenter(tris []triangle) geo.Vec3 {
	min := tris[0].v0
	max := tris[0].v0
	for _, t := range tris {
		for _, v := range []geo.Vec3{t.v0, t.v1, t.v2} {
			min = geo.MinVec(min, v)
			max = geo.MaxVec(max, v)
		}
	}
	return min.Add(max).Mul(0.5)
}

func centerTriangles(tris []triangle, center geo.Vec3) {
	for i := range tris {
		tris[i].v0 = tris[i].v0.Sub(center)
		tris[i].v1 = tris[i].v1.Sub(center)
		tris[i].v2 = tris[i].v2.Sub(center)
	}
}

func computeRadius(tris []triangle) float64 {
	r := 0.0
	for _, t := range tris {
		for _, v := range []geo.Vec3{t.v0, t.v1, t.v2} {
			if d := v.Len(); d > r {
				r = d
			}
		}
	}
	return r
}

func scaleTriangles(tris []triangle, s float64) {
	for i := range tris {
		tris[i].v0 = tris[i].v0.Mul(s)
		tris[i].v1 = tris[i].v1.Mul(s)
		tris[i].v2 = tris[i].v2.Mul(s)
	}
}

func buildBVH(tris []triangle) *bvhNode {
	if len(tris) == 0 {
		return nil
	}
	node := &bvhNode{}
	node.min = tris[0].v0
	node.max = tris[0].v0
	for _, t := range tris {
		for _, v := range []geo.Vec3{t.v0, t.v1, t.v2} {
			node.min = geo.MinVec(node.min, v)
			node.max = geo.MaxVec(node.max, v)
		}
	}

	if len(tris) <= 4 {
		node.triangles = tris
		return node
	}

	size := node.max.Sub(node.min)
	var axis int
	if size.X >= size.Y && size.X >= size.Z {
		axis = 0
	} else if size.Y >= size.X && size.Y >= size.Z {
		axis = 1
	} else {
		axis = 2
	}

	mid := 0.5 * (node.min.Comp(axis) + node.max.Comp(axis))

	leftTris := make([]triangle, 0, len(tris)/2)
	rightTris := make([]triangle, 0, len(tris)/2)

	for _, t := range tris {
		center := t.v0.Add(t.v1).Add(t.v2).Mul(1.0 / 3)
		if center.Comp(axis) < mid {
			leftTris = append(leftTris, t)
		} else {
			rightTris = append(rightTris, t)
		}
	}

	if len(leftTris) == 0 || len(rightTris) == 0 {
		node.triangles = tris
		return node
	}

	node.left = buildBVH(leftTris)
	node.right = buildBVH(rightTris)
	return node
}
