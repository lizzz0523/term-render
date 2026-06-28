package model

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"

	"term-render/internal/color"
	"term-render/internal/geo"

	"github.com/qmuntal/gltf"
)

type Model struct {
	root     *bvhNode
	radius   float64
	textures []texture
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
	c0, c1, c2 color.Color
	t0, t1, t2 geo.Vec2
	texIdx     int
}

type texture struct {
	Image  *image.RGBA
	Width  int
	Height int
}

func NewCube(w, h, d float64) *Model {
	hw, hh, hd := w/2, h/2, d/2

	V := func(x, y, z float64) geo.Vec3 { return geo.NewVec3(x, y, z) }
	N := func(x, y, z float64) geo.Vec3 { return geo.NewVec3(x, y, z) }

	mkTris := func(v0, v1, v2 geo.Vec3, n geo.Vec3, c color.Color, t0, t1, t2 geo.Vec2, texIdx int) []triangle {
		return []triangle{
			{v0: v0, v1: v1, v2: v2, n0: n, n1: n, n2: n, c0: c, c1: c, c2: c, t0: t0, t1: t1, t2: t2, texIdx: texIdx},
		}
	}

	white := color.New(1, 1, 1)
	uvZero := geo.NewVec2(0, 0)

	tris := make([]triangle, 0, 12)
	// +X
	tris = append(tris, mkTris(V(hw, -hh, -hd), V(hw, -hh, hd), V(hw, hh, hd), N(1, 0, 0), white, uvZero, uvZero, uvZero, -1)...)
	tris = append(tris, mkTris(V(hw, -hh, -hd), V(hw, hh, hd), V(hw, hh, -hd), N(1, 0, 0), white, uvZero, uvZero, uvZero, -1)...)
	// -X
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(-hw, hh, hd), V(-hw, -hh, hd), N(-1, 0, 0), white, uvZero, uvZero, uvZero, -1)...)
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(-hw, hh, -hd), V(-hw, hh, hd), N(-1, 0, 0), white, uvZero, uvZero, uvZero, -1)...)
	// +Y
	tris = append(tris, mkTris(V(-hw, hh, -hd), V(hw, hh, -hd), V(hw, hh, hd), N(0, 1, 0), white, uvZero, uvZero, uvZero, -1)...)
	tris = append(tris, mkTris(V(-hw, hh, -hd), V(hw, hh, hd), V(-hw, hh, hd), N(0, 1, 0), white, uvZero, uvZero, uvZero, -1)...)
	// -Y
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(hw, -hh, hd), V(hw, -hh, -hd), N(0, -1, 0), white, uvZero, uvZero, uvZero, -1)...)
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(-hw, -hh, hd), V(hw, -hh, hd), N(0, -1, 0), white, uvZero, uvZero, uvZero, -1)...)
	// +Z
	tris = append(tris, mkTris(V(-hw, -hh, hd), V(hw, -hh, hd), V(hw, hh, hd), N(0, 0, 1), white, uvZero, uvZero, uvZero, -1)...)
	tris = append(tris, mkTris(V(-hw, -hh, hd), V(hw, hh, hd), V(-hw, hh, hd), N(0, 0, 1), white, uvZero, uvZero, uvZero, -1)...)
	// -Z
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(hw, hh, -hd), V(hw, -hh, -hd), N(0, 0, -1), white, uvZero, uvZero, uvZero, -1)...)
	tris = append(tris, mkTris(V(-hw, -hh, -hd), V(-hw, hh, -hd), V(hw, hh, -hd), N(0, 0, -1), white, uvZero, uvZero, uvZero, -1)...)

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

	textures, err := loadTextures(doc)
	if err != nil {
		return nil, err
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

			var colors []float64
			if idx, ok := prim.Attributes["COLOR_0"]; ok {
				colAcc := doc.Accessors[idx]
				colors = readColors(doc, colAcc)
			}

			var uvs []float64
			if idx, ok := prim.Attributes["TEXCOORD_0"]; ok {
				uvAcc := doc.Accessors[idx]
				uvs = readVec2s(doc, uvAcc)
			}

			materialColor := color.New(1, 1, 1)
			if prim.Material != nil {
				mat := doc.Materials[*prim.Material]
				if mat.PBRMetallicRoughness != nil {
					bcf := mat.PBRMetallicRoughness.BaseColorFactorOrDefault()
					materialColor = color.New(bcf[0], bcf[1], bcf[2])
				}
			}

			texIdx := -1
			if prim.Material != nil {
				mat := doc.Materials[*prim.Material]
				if mat.PBRMetallicRoughness != nil && mat.PBRMetallicRoughness.BaseColorTexture != nil {
					texInfo := mat.PBRMetallicRoughness.BaseColorTexture
					if texInfo.Index < len(doc.Textures) {
						tex := doc.Textures[texInfo.Index]
						if tex.Source != nil && *tex.Source < len(textures) {
							texIdx = *tex.Source
						}
					}
				}
			}
			// Fallback: if the primitive has UV coordinates and textures exist
			// but no BaseColorTexture, assume the first texture should be used.
			if texIdx < 0 && len(uvs) > 0 && len(textures) > 0 {
				texIdx = 0
			}

			indices := readIndices(doc, doc.Accessors[*prim.Indices])

			tris := buildTriangles(positions, normals, colors, uvs, materialColor, texIdx, indices, worldMat)
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

	return &Model{root: root, radius: targetRadius, textures: textures}, nil
}

func buildParentMap(doc *gltf.Document) map[int]int {
	parent := make(map[int]int)
	for i, n := range doc.Nodes {
		for _, child := range n.Children {
			parent[child] = i
		}
	}
	return parent
}

func buildMeshMap(doc *gltf.Document) map[int]int {
	result := make(map[int]int)
	for i, n := range doc.Nodes {
		if n.Mesh != nil {
			result[*n.Mesh] = i
		}
	}
	return result
}

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

func nodeLocalMatrix(n *gltf.Node) geo.Mat4 {
	// Decompose from T * R * S when no explicit matrix is provided.
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

func readColors(doc *gltf.Document, acc *gltf.Accessor) []float64 {
	bufView := doc.BufferViews[*acc.BufferView]
	bufData := doc.Buffers[bufView.Buffer].Data
	start := bufView.ByteOffset + acc.ByteOffset
	count := acc.Count
	totalComps := acc.Type.Components()
	rgbComps := 3

	result := make([]float64, count*rgbComps)

	switch acc.ComponentType {
	case gltf.ComponentFloat:
		for i := 0; i < count; i++ {
			src := start + i*totalComps*4
			dst := i * rgbComps
			for j := 0; j < rgbComps; j++ {
				result[dst+j] = float64(math.Float32frombits(
					binary.LittleEndian.Uint32(bufData[src+j*4:])))
			}
		}
	case gltf.ComponentByte:
		for i := 0; i < count; i++ {
			src := start + i*totalComps
			dst := i * rgbComps
			for j := 0; j < rgbComps; j++ {
				v := float64(int8(bufData[src+j]))
				if v < 0 {
					v = 0
				}
				result[dst+j] = v / 127.0
			}
		}
	case gltf.ComponentUbyte:
		for i := 0; i < count; i++ {
			src := start + i*totalComps
			dst := i * rgbComps
			for j := 0; j < rgbComps; j++ {
				result[dst+j] = float64(bufData[src+j]) / 255.0
			}
		}
	case gltf.ComponentShort:
		for i := 0; i < count; i++ {
			src := start + i*totalComps*2
			dst := i * rgbComps
			for j := 0; j < rgbComps; j++ {
				v := float64(int16(binary.LittleEndian.Uint16(bufData[src+j*2:])))
				if v < 0 {
					v = 0
				}
				result[dst+j] = v / 32767.0
			}
		}
	case gltf.ComponentUshort:
		for i := 0; i < count; i++ {
			src := start + i*totalComps*2
			dst := i * rgbComps
			for j := 0; j < rgbComps; j++ {
				result[dst+j] = float64(binary.LittleEndian.Uint16(
					bufData[src+j*2:])) / 65535.0
			}
		}
	}
	return result
}

func readVec2s(doc *gltf.Document, acc *gltf.Accessor) []float64 {
	bufView := doc.BufferViews[*acc.BufferView]
	bufData := doc.Buffers[bufView.Buffer].Data
	start := bufView.ByteOffset + acc.ByteOffset
	count := acc.Count
	byteSize := acc.ComponentType.ByteSize()

	result := make([]float64, count*2)
	for i := 0; i < count*2; i++ {
		offset := start + i*byteSize
		result[i] = float64(math.Float32frombits(binary.LittleEndian.Uint32(bufData[offset:])))
	}
	return result
}

func loadTextures(doc *gltf.Document) ([]texture, error) {
	var textures []texture
	for _, img := range doc.Images {
		var raw []byte
		if img.BufferView != nil {
			bv := doc.BufferViews[*img.BufferView]
			buf := doc.Buffers[bv.Buffer].Data
			raw = buf[bv.ByteOffset : bv.ByteOffset+bv.ByteLength]
		} else if img.URI != "" {
			return nil, fmt.Errorf("external texture references not supported: %s", img.URI)
		}
		if len(raw) == 0 {
			continue
		}

		decoded, _, err := image.Decode(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("failed to decode texture: %w", err)
		}

		bounds := decoded.Bounds()
		rgba := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rgba.Set(x, y, decoded.At(x, y))
			}
		}
		textures = append(textures, texture{
			Image:  rgba,
			Width:  bounds.Dx(),
			Height: bounds.Dy(),
		})
	}
	return textures, nil
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

func buildTriangles(positions, normals, colors, uvs []float64, materialColor color.Color, texIdx int, indices []uint32, mat geo.Mat4) []triangle {
	tris := make([]triangle, 0, len(indices)/3)
	hasNormals := len(normals) == len(positions)
	hasColors := len(colors) > 0
	hasUVs := len(uvs) > 0

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

		var c0, c1, c2 color.Color
		white := color.New(1, 1, 1)
		if hasColors {
			c0 = color.New(colors[i0*3], colors[i0*3+1], colors[i0*3+2])
			c1 = color.New(colors[i1*3], colors[i1*3+1], colors[i1*3+2])
			c2 = color.New(colors[i2*3], colors[i2*3+1], colors[i2*3+2])
		} else {
			c0, c1, c2 = white, white, white
		}
		// Only multiply materialColor when there is no texture (texIdx < 0).
		// When a texture is present, color is sampled from it in intersect.
		if texIdx < 0 {
			c0 = c0.Mul(materialColor)
			c1 = c1.Mul(materialColor)
			c2 = c2.Mul(materialColor)
		}

		var t0, t1, t2 geo.Vec2
		if hasUVs {
			t0 = geo.NewVec2(uvs[i0*2], uvs[i0*2+1])
			t1 = geo.NewVec2(uvs[i1*2], uvs[i1*2+1])
			t2 = geo.NewVec2(uvs[i2*2], uvs[i2*2+1])
		}

		tris = append(tris, triangle{v0, v1, v2, n0, n1, n2, c0, c1, c2, t0, t1, t2, texIdx})
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
