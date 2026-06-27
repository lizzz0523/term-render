package model

import (
	"encoding/binary"
	"fmt"
	"math"

	"term-render/internal/geo"

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

func NewCube(w, h, d float64) *Model {
	hw, hh, hd := w/2, h/2, d/2

	V := func(x, y, z float64) geo.Vec3 { return geo.NewVec3(x, y, z) }
	N := func(x, y, z float64) geo.Vec3 { return geo.NewVec3(x, y, z) }

	mkTris := func(v0, v1, v2 geo.Vec3, n geo.Vec3) []triangle {
		return []triangle{
			{v0: v0, v1: v1, v2: v2, n0: n, n1: n, n2: n},
		}
	}

	tris := make([]triangle, 0, 12)
	// +X
	tris = append(tris, mkTris(V(hw, -hh, -hd), V(hw, -hh, hd), V(hw, hh, hd), N(1, 0, 0))...)
	tris = append(tris, mkTris(V(hw, -hh, -hd), V(hw, hh, hd), V(hw, hh, -hd), N(1, 0, 0))...)
	// -X
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(-hw, hh, hd), V(-hw, -hh, hd), N(-1, 0, 0))...)
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(-hw, hh, -hd), V(-hw, hh, hd), N(-1, 0, 0))...)
	// +Y
	tris = append(tris, mkTris(V(-hw, hh, -hd), V(hw, hh, -hd), V(hw, hh, hd), N(0, 1, 0))...)
	tris = append(tris, mkTris(V(-hw, hh, -hd), V(hw, hh, hd), V(-hw, hh, hd), N(0, 1, 0))...)
	// -Y
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(hw, -hh, hd), V(hw, -hh, -hd), N(0, -1, 0))...)
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(-hw, -hh, hd), V(hw, -hh, hd), N(0, -1, 0))...)
	// +Z
	tris = append(tris, mkTris(V(-hw, -hh, hd), V(hw, -hh, hd), V(hw, hh, hd), N(0, 0, 1))...)
	tris = append(tris, mkTris(V(-hw, -hh, hd), V(hw, hh, hd), V(-hw, hh, hd), N(0, 0, 1))...)
	// -Z
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(hw, hh, -hd), V(hw, -hh, -hd), N(0, 0, -1))...)
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(-hw, hh, -hd), V(hw, hh, -hd), N(0, 0, -1))...)

	root := buildBVH(tris)
	return &Model{root: root, radius: math.Sqrt(hw*hw + hh*hh + hd*hd)}
}

func LoadGLB(path string, targetRadius float64) (*Model, error) {
	doc, err := gltf.Open(path)
	if err != nil {
		return nil, err
	}

	if len(doc.Meshes) == 0 {
		return nil, fmt.Errorf("no meshes found in GLB")
	}

	parent := buildParentMap(doc)
	meshNode := buildMeshMap(doc)
	worldMatrices := computeWorldMatrices(doc, parent)

	var allTris []triangle

	for meshIdx, mesh := range doc.Meshes {
		worldMat := geo.Mat4Identity()
		if nodeIdx, ok := meshNode[meshIdx]; ok {
			worldMat = worldMatrices[nodeIdx]
		}

		for _, prim := range mesh.Primitives {
			posAcc := doc.Accessors[prim.Attributes["POSITION"]]
			positions := readFloats(doc, posAcc)

			var normals []float64
			if idx, ok := prim.Attributes["NORMAL"]; ok {
				nrmAcc := doc.Accessors[idx]
				normals = readFloats(doc, nrmAcc)
			}

			indices := readIndices(doc, doc.Accessors[*prim.Indices])

			tris := buildTriangles(positions, normals, indices, worldMat)
			allTris = append(allTris, tris...)
		}
	}

	if len(allTris) == 0 {
		return nil, fmt.Errorf("no triangles generated")
	}

	center := computeCenter(allTris)
	centerTriangles(allTris, center)

	radius := computeRadius(allTris)
	scale := targetRadius / radius
	scaleTriangles(allTris, scale)

	root := buildBVH(allTris)

	return &Model{root: root, radius: targetRadius}, nil
}

// buildParentMap maps each child node index to its parent node index.
func buildParentMap(doc *gltf.Document) map[int]int {
	parent := make(map[int]int)
	for i, n := range doc.Nodes {
		for _, child := range n.Children {
			parent[child] = i
		}
	}
	return parent
}

// buildMeshMap returns a map from mesh index to the node index that references it.
func buildMeshMap(doc *gltf.Document) map[int]int {
	result := make(map[int]int)
	for i, n := range doc.Nodes {
		if n.Mesh != nil {
			result[*n.Mesh] = i
		}
	}
	return result
}

// computeWorldMatrices computes the world-space transform for every node
// by traversing the scene graph from root nodes.
func computeWorldMatrices(doc *gltf.Document, parent map[int]int) []geo.Mat4 {
	result := make([]geo.Mat4, len(doc.Nodes))
	for i := range result {
		result[i] = geo.Mat4Identity()
	}

	var roots []int
	for i := range doc.Nodes {
		if _, hasParent := parent[i]; !hasParent {
			roots = append(roots, i)
		}
	}

	var traverse func(nodeIdx int, parentWorld geo.Mat4)
	traverse = func(nodeIdx int, parentWorld geo.Mat4) {
		if nodeIdx >= len(doc.Nodes) {
			return
		}
		n := doc.Nodes[nodeIdx]
		local := nodeLocalMatrix(n)
		world := parentWorld.Mul(local)
		result[nodeIdx] = world

		for _, child := range n.Children {
			traverse(child, world)
		}
	}

	for _, root := range roots {
		traverse(root, geo.Mat4Identity())
	}

	return result
}

// nodeLocalMatrix returns the local TRS matrix.
// Uses the explicit Matrix field if set, otherwise decomposes from T*R*S.
func nodeLocalMatrix(n *gltf.Node) geo.Mat4 {
	m := n.MatrixOrDefault()
	if m != gltf.DefaultMatrix {
		return geo.Mat4From64(m)
	}

	tx, ty, tz := n.Translation[0], n.Translation[1], n.Translation[2]
	T := geo.Mat4Translation(tx, ty, tz)

	r := n.RotationOrDefault()
	R := geo.Mat4Rotation(r[0], r[1], r[2], r[3])

	sx, sy, sz := n.Scale[0], n.Scale[1], n.Scale[2]
	S := geo.Mat4Identity()
	if sx != 1 || sy != 1 || sz != 1 {
		S = geo.Mat4Scale(sx, sy, sz)
	}

	return T.Mul(R).Mul(S)
}

func readFloats(doc *gltf.Document, acc *gltf.Accessor) []float64 {
	bufView := doc.BufferViews[*acc.BufferView]
	bufData := doc.Buffers[bufView.Buffer].Data
	start := bufView.ByteOffset + acc.ByteOffset
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
	bufView := doc.BufferViews[*acc.BufferView]
	bufData := doc.Buffers[bufView.Buffer].Data
	start := bufView.ByteOffset + acc.ByteOffset
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

func buildTriangles(positions, normals []float64, indices []uint32, mat geo.Mat4) []triangle {
	tris := make([]triangle, 0, len(indices)/3)
	hasNormals := len(normals) == len(positions)

	for i := 0; i < len(indices); i += 3 {
		i0, i1, i2 := indices[i], indices[i+1], indices[i+2]

		v0 := geo.NewVec3(positions[i0*3], positions[i0*3+1], positions[i0*3+2])
		v1 := geo.NewVec3(positions[i1*3], positions[i1*3+1], positions[i1*3+2])
		v2 := geo.NewVec3(positions[i2*3], positions[i2*3+1], positions[i2*3+2])

		v0 = mat.TransformPoint(v0)
		v1 = mat.TransformPoint(v1)
		v2 = mat.TransformPoint(v2)

		var n0, n1, n2 geo.Vec3
		if hasNormals {
			n0 = geo.NewVec3(normals[i0*3], normals[i0*3+1], normals[i0*3+2])
			n1 = geo.NewVec3(normals[i1*3], normals[i1*3+1], normals[i1*3+2])
			n2 = geo.NewVec3(normals[i2*3], normals[i2*3+1], normals[i2*3+2])
			n0 = mat.TransformDirection(n0).Norm()
			n1 = mat.TransformDirection(n1).Norm()
			n2 = mat.TransformDirection(n2).Norm()
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
