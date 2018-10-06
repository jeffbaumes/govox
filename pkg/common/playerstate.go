package common

import "github.com/go-gl/mathgl/mgl32"

// PlayerState holds the state of a person
type PlayerState struct {
	Name     string
	Position mgl32.Vec3
	LookDir  mgl32.Vec3
	SendText string
}
