package scene

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/govox/pkg/common"
	pb "github.com/jeffbaumes/govox/pkg/govox"
)

// FocusCell draws an outline around the focused cell
type FocusCell struct {
	program           uint32
	drawableVAO       uint32
	pointsVBO         uint32
	projectionUniform int32
}

// NewFocusCell creates a new focus cell object
func NewFocusCell() *FocusCell {
	focusRen := FocusCell{}

	const vertexShader = `
		#version 410
		uniform mat4 proj;
		in vec3 position;
		void main() {
			gl_Position = proj * vec4(position, 1.0);
		}
	`

	const fragmentShader = `
		#version 410
		out vec4 frag_color;
		void main() {
			frag_color = vec4(0,0,0,1);
		}
	`

	focusRen.program = createProgram(vertexShader, fragmentShader)
	bindAttribute(focusRen.program, 0, "position")
	focusRen.projectionUniform = uniformLocation(focusRen.program, "proj")
	if focusRen.projectionUniform < 0 {
		panic("Could not find projection uniform")
	}
	focusRen.pointsVBO = newVBO()
	focusRen.drawableVAO = newPointsVAO(focusRen.pointsVBO, 3)
	return &focusRen
}

// Draw draws an outline around the focused cell
func (focusRen *FocusCell) Draw(player *common.Player, planet *common.Planet, w *glfw.Window) {
	gl.UseProgram(focusRen.program)
	lonCells, latCells := planet.LonLatCellsInChunkIndex(planet.CellIndexToChunkIndex(player.FocusCellIndex))
	lonWidth := common.ChunkSize / lonCells
	latWidth := common.ChunkSize / latCells

	pts := make([]float32, len(box))
	for i := 0; i < len(box); i += 3 {
		ind := pb.CellLoc{
			Lon: float64(int(player.FocusCellIndex.Lon)/lonWidth*lonWidth) + float64(lonWidth-1)/2 + float64(lonWidth)*float64(box[i+0]*1.01),
			Lat: float64(int(player.FocusCellIndex.Lat)/latWidth*latWidth) + float64(latWidth-1)/2 + float64(latWidth)*float64(box[i+1]*1.01),
			Alt: float64(player.FocusCellIndex.Alt) + float64(box[i+2]*1.01),
		}
		pt := planet.CellLocToCartesian(ind)
		pts[i+0] = pt.X()
		pts[i+1] = pt.Y()
		pts[i+2] = pt.Z()
	}

	fillVBO(focusRen.pointsVBO, pts)

	lookDir := player.LookDir()
	view := mgl32.LookAtV(player.Location(), player.Location().Add(lookDir), player.Location().Normalize())
	width, height := FramebufferSize(w)
	perspective := mgl32.Perspective(float32(60*math.Pi/180), float32(width)/float32(height), 0.01, 1000)
	proj := perspective.Mul4(view)
	gl.UniformMatrix4fv(focusRen.projectionUniform, 1, false, &proj[0])

	gl.BindVertexArray(focusRen.drawableVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(pts)/3))
}
