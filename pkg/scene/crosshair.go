package scene

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Crosshair draws the center crosshair
type Crosshair struct {
	program           uint32
	drawableVAO       uint32
	pointsVBO         uint32
	projectionUniform int32
}

// NewCrosshair creates a new crosshair
func NewCrosshair() *Crosshair {
	const vertexShader = `
		#version 410
		in vec3 vp;
		uniform mat4 proj;
		void main() {
			gl_Position = proj * vec4(vp, 1.0);
		}
	`

	const fragmentShader = `
		#version 410
		out vec4 frag_color;
		void main() {
			frag_color = vec4(1.0, 1.0, 1.0, 1.0);
		}
	`

	over := Crosshair{}
	points := []float32{
		-20.0, 0.0, 0.0,
		19.0, 0.0, 0.0,

		0.0, -20.0, 0.0,
		0.0, 19.0, 0.0,
	}
	over.pointsVBO = newVBO()
	fillVBO(over.pointsVBO, points)
	over.drawableVAO = newPointsVAO(over.pointsVBO, 3)

	over.program = createProgram(vertexShader, fragmentShader)
	bindAttribute(over.program, 0, "vp")

	over.projectionUniform = uniformLocation(over.program, "proj")
	return &over
}

// Draw renders the crosshair
func (over *Crosshair) Draw(w *glfw.Window) {
	gl.UseProgram(over.program)
	width, height := w.GetSize()
	proj := mgl32.Scale3D(1/float32(width), 1/float32(height), 1.0)
	gl.UniformMatrix4fv(over.projectionUniform, 1, false, &proj[0])

	gl.BindVertexArray(over.drawableVAO)
	gl.DrawArrays(gl.LINES, 0, 4)
}
