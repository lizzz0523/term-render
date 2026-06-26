package main

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/qmuntal/gltf"
)

type Triangle struct {
	V0, V1, V2 vec3
	N0, N1, N2 vec3
}

type BVHNode struct {
	Min, Max  vec3
	Left      *BVHNode
	Right     *BVHNode
	Triangles []Triangle
}

type Model struct {
	Root   *BVHNode
	Radius float64
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

	return &Model{Root: root, Radius: targetRadius}, nil
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

func buildTriangles(positions, normals []float64, indices []uint32) []Triangle {
	tris := make([]Triangle, 0, len(indices)/3)
	hasNormals := len(normals) == len(positions)

	for i := 0; i < len(indices); i += 3 {
		i0, i1, i2 := indices[i], indices[i+1], indices[i+2]

		v0 := vec3{positions[i0*3], positions[i0*3+1], positions[i0*3+2]}
		v1 := vec3{positions[i1*3], positions[i1*3+1], positions[i1*3+2]}
		v2 := vec3{positions[i2*3], positions[i2*3+1], positions[i2*3+2]}

		var n0, n1, n2 vec3
		if hasNormals {
			n0 = vec3{normals[i0*3], normals[i0*3+1], normals[i0*3+2]}
			n1 = vec3{normals[i1*3], normals[i1*3+1], normals[i1*3+2]}
			n2 = vec3{normals[i2*3], normals[i2*3+1], normals[i2*3+2]}
		} else {
			e1 := v1.sub(v0)
			e2 := v2.sub(v0)
			fn := e1.cross(e2).norm()
			n0, n1, n2 = fn, fn, fn
		}

		tris = append(tris, Triangle{V0: v0, V1: v1, V2: v2, N0: n0, N1: n1, N2: n2})
	}
	return tris
}

func computeCenter(tris []Triangle) vec3 {
	min := tris[0].V0
	max := tris[0].V0
	for _, t := range tris {
		for _, v := range []vec3{t.V0, t.V1, t.V2} {
			min = minVec(min, v)
			max = maxVec(max, v)
		}
	}
	return min.add(max).mul(0.5)
}

func centerTriangles(tris []Triangle, center vec3) {
	for i := range tris {
		tris[i].V0 = tris[i].V0.sub(center)
		tris[i].V1 = tris[i].V1.sub(center)
		tris[i].V2 = tris[i].V2.sub(center)
	}
}

func computeRadius(tris []Triangle) float64 {
	r := 0.0
	for _, t := range tris {
		for _, v := range []vec3{t.V0, t.V1, t.V2} {
			if d := v.len(); d > r {
				r = d
			}
		}
	}
	return r
}

func scaleTriangles(tris []Triangle, s float64) {
	for i := range tris {
		tris[i].V0 = tris[i].V0.mul(s)
		tris[i].V1 = tris[i].V1.mul(s)
		tris[i].V2 = tris[i].V2.mul(s)
	}
}

func buildBVH(tris []Triangle) *BVHNode {
	if len(tris) == 0 {
		return nil
	}
	node := &BVHNode{}
	node.Min = tris[0].V0
	node.Max = tris[0].V0
	for _, t := range tris {
		for _, v := range []vec3{t.V0, t.V1, t.V2} {
			node.Min = minVec(node.Min, v)
			node.Max = maxVec(node.Max, v)
		}
	}

	if len(tris) <= 4 {
		node.Triangles = tris
		return node
	}

	size := node.Max.sub(node.Min)
	var axis int
	if size.x >= size.y && size.x >= size.z {
		axis = 0
	} else if size.y >= size.x && size.y >= size.z {
		axis = 1
	} else {
		axis = 2
	}

	mid := 0.5 * (node.Min.comp(axis) + node.Max.comp(axis))

	leftTris := make([]Triangle, 0, len(tris)/2)
	rightTris := make([]Triangle, 0, len(tris)/2)

	for _, t := range tris {
		center := t.V0.add(t.V1).add(t.V2).mul(1.0 / 3)
		if center.comp(axis) < mid {
			leftTris = append(leftTris, t)
		} else {
			rightTris = append(rightTris, t)
		}
	}

	if len(leftTris) == 0 || len(rightTris) == 0 {
		node.Triangles = tris
		return node
	}

	node.Left = buildBVH(leftTris)
	node.Right = buildBVH(rightTris)
	return node
}
