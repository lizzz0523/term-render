package geo

import "math"

// Mat4 is a 4x4 matrix stored in column-major order.
type Mat4 [16]float64

func Mat4Identity() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func Mat4Translation(x, y, z float64) Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

func Mat4Scale(x, y, z float64) Mat4 {
	return Mat4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

func Mat4Rotation(qx, qy, qz, qw float64) Mat4 {
	return Mat4{
		1 - 2*qy*qy - 2*qz*qz, 2*qx*qy + 2*qz*qw, 2*qx*qz - 2*qy*qw, 0,
		2*qx*qy - 2*qz*qw, 1 - 2*qx*qx - 2*qz*qz, 2*qy*qz + 2*qx*qw, 0,
		2*qx*qz + 2*qy*qw, 2*qy*qz - 2*qx*qw, 1 - 2*qx*qx - 2*qy*qy, 0,
		0, 0, 0, 1,
	}
}

func Mat4From64(m [16]float64) Mat4 {
	return Mat4{
		m[0], m[1], m[2], m[3],
		m[4], m[5], m[6], m[7],
		m[8], m[9], m[10], m[11],
		m[12], m[13], m[14], m[15],
	}
}

func (m Mat4) Mul(n Mat4) Mat4 {
	var out Mat4
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			sum := 0.0
			for k := 0; k < 4; k++ {
				sum += m[k*4+row] * n[col*4+k]
			}
			out[col*4+row] = sum
		}
	}
	return out
}

// TransformPoint applies the full 4x4 transform to a point.
// Column-major: col[0]=m[0..3], col[1]=m[4..7], col[2]=m[8..11], col[3]=m[12..15].
func (m Mat4) TransformPoint(v Vec3) Vec3 {
	x := m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]
	y := m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]
	z := m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]
	w := m[3]*v.X + m[7]*v.Y + m[11]*v.Z + m[15]
	if math.Abs(w) > 1e-10 {
		return Vec3{x / w, y / w, z / w}
	}
	return Vec3{x, y, z}
}

// TransformDirection applies only the 3x3 portion (w=0, no translation).
func (m Mat4) TransformDirection(v Vec3) Vec3 {
	x := m[0]*v.X + m[4]*v.Y + m[8]*v.Z
	y := m[1]*v.X + m[5]*v.Y + m[9]*v.Z
	z := m[2]*v.X + m[6]*v.Y + m[10]*v.Z
	return Vec3{x, y, z}
}
