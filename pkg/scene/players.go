package scene

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/govox/pkg/common"
)

// Players draws the other players in the game
type Players struct {
	connectedPlayers  *[]*common.PlayerState
	program           uint32
	drawableVAO       uint32
	pointsVBO         uint32
	normalsVBO        uint32
	projectionUniform int32
}

// NewPlayers creates a new Players object
func NewPlayers(cp *[]*common.PlayerState) *Players {
	const vertexShader = `
		#version 410
		in vec3 vp;
		in vec3 n;
		uniform mat4 proj;
		out vec3 color;
		out vec3 light;
		void main() {
			color = n;
			gl_Position = proj * vec4(vp, 1.0);

			// Apply lighting effect
			highp vec3 ambientLight = vec3(0.0, 0.0, 0.2);
			highp vec3 light1Color = vec3(0.5, 0.5, 0.4);
			highp vec3 light1Dir = normalize(vec3(0.85, 0.8, 0.75));
			highp float light1 = max(dot(n, light1Dir), 0.0);
			highp vec3 light2Color = vec3(0.1, 0.1, 0.2);
			highp vec3 light2Dir = normalize(vec3(-0.85, -0.8, -0.75));
			highp float light2 = max(dot(n, light2Dir), 0.0);
			light = ambientLight + (light1Color * light1) + (light2Color * light2);
		}
	`

	const fragmentShader = `
		#version 410
		in vec3 color;
		in vec3 light;
		out vec4 frag_color;
		void main() {
			frag_color = vec4(light, 1.0);
		}
	`

	peopleRen := Players{}
	peopleRen.connectedPlayers = cp
	peopleRen.program = createProgram(vertexShader, fragmentShader)
	bindAttribute(peopleRen.program, 0, "vp")
	bindAttribute(peopleRen.program, 1, "n")
	peopleRen.projectionUniform = uniformLocation(peopleRen.program, "proj")
	peopleRen.pointsVBO = newVBO()
	fillVBO(peopleRen.pointsVBO, cube)
	peopleRen.normalsVBO = newVBO()
	peopleRen.drawableVAO = newPointsNormalsVAO(peopleRen.pointsVBO, peopleRen.normalsVBO)
	return &peopleRen
}

// Draw draws the other players
func (peopleRen *Players) Draw(player *common.Player, w *glfw.Window) {
	gl.UseProgram(peopleRen.program)
	for _, p := range *peopleRen.connectedPlayers {
		pts := make([]float32, len(cube))
		for i := 0; i < len(cube); i += 3 {
			pts[i] = p.Position[0] + cube[i]
			pts[i+1] = p.Position[1] + cube[i+1]
			pts[i+2] = p.Position[2] + cube[i+2]
		}

		nms := make([]float32, len(cube))
		for i := 0; i < len(cube); i += 9 {
			p1 := mgl32.Vec3{pts[i+0], pts[i+1], pts[i+2]}
			p2 := mgl32.Vec3{pts[i+3], pts[i+4], pts[i+5]}
			p3 := mgl32.Vec3{pts[i+6], pts[i+7], pts[i+8]}
			v1 := p1.Sub(p2)
			v2 := p1.Sub(p3)
			n := v1.Cross(v2).Normalize()
			for j := 0; j < 3; j++ {
				nms[i+3*j+0] = n[0]
				nms[i+3*j+1] = n[1]
				nms[i+3*j+2] = n[2]
			}
		}

		fillVBO(peopleRen.normalsVBO, nms)

		lookDir := player.LookDir()
		view := mgl32.LookAtV(player.Location(), player.Location().Add(lookDir), player.Location().Normalize())
		width, height := FramebufferSize(w)
		perspective := mgl32.Perspective(45, float32(width)/float32(height), 0.01, 1000)
		proj := perspective.Mul4(view)
		proj = proj.Mul4(mgl32.Translate3D(p.Position[0], p.Position[1], p.Position[2]))
		right := p.LookDir.Cross(p.Position.Normalize()).Normalize()
		up := right.Cross(p.LookDir).Normalize()
		proj = proj.Mul4(mgl32.Mat4FromCols(p.LookDir.Vec4(0), up.Vec4(0), right.Vec4(0), mgl32.Vec4{0, 0, 0, 1}))
		gl.UniformMatrix4fv(peopleRen.projectionUniform, 1, false, &proj[0])

		gl.BindVertexArray(peopleRen.drawableVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(cube)/3))
	}
}
