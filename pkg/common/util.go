package common

import "github.com/go-gl/mathgl/mgl32"

// Project projects one vector onto another
func Project(a mgl32.Vec3, b mgl32.Vec3) mgl32.Vec3 {
	bn := b.Normalize()
	return bn.Mul(a.Dot(bn))
}

// ProjectToPlane projects a vector onto a plane with a given normal
func ProjectToPlane(v mgl32.Vec3, n mgl32.Vec3) mgl32.Vec3 {
	if v[0] == 0 && v[1] == 0 && v[2] == 0 {
		return v
	}
	// To project vector to plane, subtract vector projected to normal
	return v.Sub(Project(v, n))
}

// Min returns the min of two integers
func Min(val, a int) int {
	if val < a {
		return val
	}
	return a
}

// Max returns the max of two integers
func Max(val, a int) int {
	if val > a {
		return val
	}
	return a
}
