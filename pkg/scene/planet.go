package scene

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/govox/pkg/common"
	pb "github.com/jeffbaumes/govox/pkg/govox"
)

// Planet draws visible chunks for a planet
type Planet struct {
	Planet            *common.Planet
	chunkRenderers    map[common.ChunkKey]*chunkRenderer
	program           uint32
	texture           uint32
	textureUnit       int32
	projectionUniform int32
	planetLocUniform  int32
	planetRotUniform  int32

	chunkProgram           uint32
	chunkProjectionUniform int32
	chunkPlanetLocUniform  int32
	chunkPlanetRotUniform  int32
	chunkTextureUniform    int32

	drawableVAO     uint32
	pointsVBO       uint32
	colorsVBO       uint32
	numTriangles    int32
	geometryUpdated bool
}

// NewPlanet creates a new planet renderer
func NewPlanet(planet *common.Planet) *Planet {
	const vertexShader = `
		#version 410
		in vec3 vp;
		in vec4 c;
		uniform mat4 proj;
		uniform mat3 planetrot;
		uniform vec3 planetloc;
		out vec4 color;
		out vec3 light;
		void main() {
			color = c;
			gl_Position = proj * vec4(vp, 1.0);

			highp vec3 rotated = planetrot * vp;
			highp vec3 sundir = normalize(planetloc + rotated);

			// Apply lighting effect
			highp vec3 ambientLight = vec3(0, 0, 0);
			highp vec3 vpn = normalize(rotated);
			highp vec3 light1Color = vec3(0.9, 0.9, 0.9);
			highp float light1 = max(sqrt(dot(vpn, sundir)), 0.0);
			highp vec3 light2Color = vec3(0.2, 0.2, 0.2);
			highp float light2 = max(sqrt(1 - dot(vpn, sundir)), 0.0);
			highp vec3 light3Color = vec3(1.0, 0.5, 0.1);
			highp float light3 = max(0.4 - sqrt(abs(dot(vpn, sundir))), 0.0);
			light = ambientLight + (light1Color * light1) + (light2Color * light2) + (light3Color * light3);
		}
	`

	const fragmentShader = `
		#version 410
		in vec4 color;
		in vec3 light;
		out vec4 frag_color;
		void main() {
			frag_color = color * vec4(light, 1.0);
		}
	`

	const vertexShaderChunk = `
		#version 410
		in vec3 vp;
		in vec3 n;
		in vec2 t;
		uniform mat4 proj;
		uniform mat3 planetrot;
		uniform vec3 planetloc;
		out vec3 light;
		out vec2 texcoord;
		void main() {
			texcoord = t;
			gl_Position = proj * vec4(vp, 1.0);

			highp vec3 rotated = planetrot * vp;
			highp vec3 sundir = normalize(planetloc + rotated);

			// Apply lighting effect
			highp vec3 ambientLight = vec3(0, 0, 0);
			highp vec3 vpn = normalize(rotated);
			highp vec3 light1Color = vec3(0.9, 0.9, 0.9);
			highp float light1 = max(sqrt(dot(vpn, sundir)), 0.0);
			highp vec3 light2Color = vec3(0.2, 0.2, 0.2);
			highp float light2 = max(sqrt(1 - dot(vpn, sundir)), 0.0);
			highp vec3 light3Color = vec3(1.0, 0.5, 0.1);
			highp float light3 = max(0.4 - sqrt(abs(dot(vpn, sundir))), 0.0);
			light = ambientLight + (light1Color * light1) + (light2Color * light2) + (light3Color * light3);
		}
	`

	const fragmentShaderChunk = `
		#version 410
		in vec3 light;
		in vec2 texcoord;
		uniform sampler2D texBase;
		out vec4 frag_color;
		void main() {
			vec4 texel = texture(texBase, texcoord);
			frag_color = texel * vec4(light, 1.0);
		}
	`

	pr := Planet{}
	pr.Planet = planet
	pr.chunkProgram = createProgramNoLink(vertexShaderChunk, fragmentShaderChunk)
	pr.chunkRenderers = make(map[common.ChunkKey]*chunkRenderer)
	bindAttribute(pr.chunkProgram, 0, "vp")
	bindAttribute(pr.chunkProgram, 1, "n")
	bindAttribute(pr.chunkProgram, 2, "t")
	gl.LinkProgram(pr.chunkProgram)
	pr.chunkProjectionUniform = uniformLocation(pr.chunkProgram, "proj")
	pr.chunkPlanetLocUniform = uniformLocation(pr.chunkProgram, "planetloc")
	pr.chunkPlanetRotUniform = uniformLocation(pr.chunkProgram, "planetrot")
	pr.chunkTextureUniform = uniformLocation(pr.chunkProgram, "texBase")

	pr.program = createProgramNoLink(vertexShader, fragmentShader)
	bindAttribute(pr.program, 0, "vp")
	bindAttribute(pr.program, 1, "c")
	gl.LinkProgram(pr.program)

	pr.projectionUniform = uniformLocation(pr.program, "proj")
	pr.planetLocUniform = uniformLocation(pr.program, "planetloc")
	pr.planetRotUniform = uniformLocation(pr.program, "planetrot")

	rgba := LoadTextures()

	pr.textureUnit = 1
	gl.ActiveTexture(uint32(gl.TEXTURE0 + pr.textureUnit))
	gl.GenTextures(1, &pr.texture)
	gl.BindTexture(gl.TEXTURE_2D, pr.texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)

	gl.GenerateMipmap(gl.TEXTURE_2D)

	pr.pointsVBO = newVBO()
	pr.colorsVBO = newVBO()
	pr.drawableVAO = newPointsColorsVAO(pr.pointsVBO, pr.colorsVBO)

	pr.Planet.GetGeometry(true)
	return &pr
}

// SetCellMaterial sets the material at a particular cell and marks its chunk for redraw
func (planetRen *Planet) SetCellMaterial(ind pb.CellIndex, material pb.Material, updateServer bool) {
	planetRen.Planet.SetCellMaterial(ind, material, updateServer)
	chunkInd := planetRen.Planet.CellIndexToChunkIndex(ind)
	chunkRen := planetRen.chunkRenderers[common.ChunkKey{Lon: chunkInd.Lon, Lat: chunkInd.Lat, Alt: chunkInd.Alt}]
	if chunkRen == nil {
		return
	}

	// Mark the chunk's geometry to be recalculated
	chunkRen.geometryUpdated = false

	lonCells := len(chunkRen.chunk.Cell)
	latCells := len(chunkRen.chunk.Cell[0].Cell)
	lonFactor := common.ChunkSize / lonCells
	latFactor := common.ChunkSize / latCells

	// If the cell is along a chunk edge, also mark the adjacent chunk dirty
	if int(ind.Lon%common.ChunkSize)/lonFactor == 0 {
		lonChunks := planetRen.Planet.LonCells / common.ChunkSize
		lonNeg := chunkInd.Lon - 1
		if lonNeg < 0 {
			lonNeg = lonChunks - 1
		}
		cr := planetRen.chunkRenderers[common.ChunkKey{Lon: lonNeg, Lat: chunkInd.Lat, Alt: chunkInd.Alt}]
		if cr != nil {
			cr.geometryUpdated = false
		}
	}
	if int(ind.Lon%common.ChunkSize)/lonFactor == lonCells-1 {
		lonChunks := planetRen.Planet.LonCells / common.ChunkSize
		lonPos := (chunkInd.Lon + 1) % lonChunks
		cr := planetRen.chunkRenderers[common.ChunkKey{Lon: lonPos, Lat: chunkInd.Lat, Alt: chunkInd.Alt}]
		if cr != nil {
			cr.geometryUpdated = false
		}
	}
	if int(ind.Lat%common.ChunkSize)/latFactor == 0 {
		cr := planetRen.chunkRenderers[common.ChunkKey{Lon: chunkInd.Lon, Lat: chunkInd.Lat - 1, Alt: chunkInd.Alt}]
		if cr != nil {
			cr.geometryUpdated = false
		}
	}
	if int(ind.Lat%common.ChunkSize)/latFactor == latCells-1 {
		cr := planetRen.chunkRenderers[common.ChunkKey{Lon: chunkInd.Lon, Lat: chunkInd.Lat + 1, Alt: chunkInd.Alt}]
		if cr != nil {
			cr.geometryUpdated = false
		}
	}
	if ind.Alt%common.ChunkSize == 0 {
		cr := planetRen.chunkRenderers[common.ChunkKey{Lon: chunkInd.Lon, Lat: chunkInd.Lat, Alt: chunkInd.Alt - 1}]
		if cr != nil {
			cr.geometryUpdated = false
		}
	}
	if ind.Alt%common.ChunkSize == common.ChunkSize-1 {
		cr := planetRen.chunkRenderers[common.ChunkKey{Lon: chunkInd.Lon, Lat: chunkInd.Lat, Alt: chunkInd.Alt + 1}]
		if cr != nil {
			cr.geometryUpdated = false
		}
	}
}

func (planetRen *Planet) location(time float64, planetMap map[int64]*Planet) mgl32.Vec3 {
	planet := planetRen.Planet
	if planet.Spec.Id == planet.Spec.OrbitPlanet {
		return mgl32.Vec3{}
	}
	orbitLoc := planetMap[planet.Spec.OrbitPlanet].location(time, planetMap)
	_, timeOfOrbit := math.Modf(time / planet.Spec.OrbitSeconds)
	orbitAng := 2 * math.Pi * timeOfOrbit
	loc := orbitLoc.Add(
		mgl32.Vec3{
			float32(math.Cos(orbitAng)),
			float32(math.Sin(orbitAng)),
			0,
		}.Mul(float32(planet.Spec.OrbitDistance)),
	)
	return loc
}

// Draw draws the planet's visible chunks
func (planetRen *Planet) Draw(player *common.Player, planetMap map[int64]*Planet, w *glfw.Window, time float64) {
	loc := player.Location()
	planetRotation := time / planetRen.Planet.Spec.RotationSeconds
	planetRotation *= 2 * math.Pi
	orbitPosition := time / planetRen.Planet.Spec.OrbitSeconds
	orbitPosition *= 2 * math.Pi
	lookDir := player.LookDir()
	view := mgl32.LookAtV(loc, loc.Add(lookDir), loc.Normalize())
	planetLoc := planetRen.location(time, planetMap)
	planetRotate := mgl32.Rotate3DZ(float32(planetRotation))
	planetRotateNeg := mgl32.Rotate3DZ(-float32(planetRotation))
	farPlane := float32(1000)
	if player.Planet.Spec.Id != planetRen.Planet.Spec.Id {
		playerPlanetLoc := planetMap[player.Planet.Spec.Id].location(time, planetMap)
		planetLoc = planetRen.location(time, planetMap)
		relativeLoc := playerPlanetLoc.Sub(planetLoc)
		_, playerPlanetRotation := math.Modf(time / player.Planet.Spec.RotationSeconds)
		playerPlanetRotation *= 2 * math.Pi
		playerPlanetRotate := mgl32.HomogRotate3DZ(float32(playerPlanetRotation))
		translate := mgl32.Translate3D(relativeLoc[0], relativeLoc[1], relativeLoc[2])
		planetRotate4 := mgl32.HomogRotate3DZ(float32(planetRotation))
		view = view.Mul4(playerPlanetRotate).Mul4(translate).Mul4(planetRotate4)
		farPlane = 10000
	}
	width, height := FramebufferSize(w)
	perspective := mgl32.Perspective(float32(60*math.Pi/180), float32(width)/float32(height), 0.01, farPlane)
	proj := perspective.Mul4(view)

	if planetRen.Planet != player.Planet {
		gl.UseProgram(planetRen.program)
		gl.UniformMatrix4fv(planetRen.projectionUniform, 1, false, &proj[0])
		gl.UniformMatrix3fv(planetRen.planetRotUniform, 1, false, &planetRotate[0])
		gl.Uniform3f(planetRen.planetLocUniform, planetLoc[0], planetLoc[1], planetLoc[2])
		planetRen.drawGeometry()
		return
	}

	gl.UseProgram(planetRen.chunkProgram)
	gl.UniformMatrix4fv(planetRen.chunkProjectionUniform, 1, false, &proj[0])
	gl.UniformMatrix3fv(planetRen.chunkPlanetRotUniform, 1, false, &planetRotateNeg[0])
	gl.Uniform1i(planetRen.chunkTextureUniform, planetRen.textureUnit)
	gl.Uniform3f(planetRen.chunkPlanetLocUniform, planetLoc[0], planetLoc[1], planetLoc[2])
	planetRen.Planet.ChunksMutex.Lock()
	for key, chunk := range planetRen.Planet.Chunks {
		if chunk.WaitingForData {
			continue
		}
		cr := planetRen.chunkRenderers[key]
		if cr == nil {
			cr = newChunkRenderer(chunk)
			planetRen.chunkRenderers[key] = cr
		}
		if !cr.geometryUpdated {
			cr.updateGeometry(planetRen.Planet, key.Lon, key.Lat, key.Alt)
		}
		cr.draw()
	}
	planetRen.Planet.ChunksMutex.Unlock()
}

func (planetRen *Planet) updateGeometry() {
	points := []float32{}
	normals := []float32{}
	colors := []float32{}

	p := planetRen.Planet
	geom := planetRen.Planet.Geometry

	lonCells := len(geom.Altitude)
	latCells := len(geom.Altitude[0].Altitude)

	appendAttributesForIndex := func(cLat, cLon int) {
		cellIndex := pb.CellIndex{
			Lon: p.LonCells * int64(cLon) / int64(lonCells),
			Lat: p.LatCells * int64(cLat) / int64(latCells-1),
			Alt: geom.Altitude[cLon].Altitude[cLat],
		}
		pt := planetRen.Planet.CellIndexToCartesian(cellIndex)
		nm := pt.Normalize()
		c := common.MaterialColors[geom.Material[cLon].Material[cLat]]
		points = append(points, pt[0], pt[1], pt[2])
		normals = append(normals, nm[0], nm[1], nm[2])
		colors = append(colors, c[0], c[1], c[2], 1.0)
	}

	for cLat := 0; cLat < latCells-1; cLat++ {
		for cLon := 0; cLon < lonCells; cLon++ {
			appendAttributesForIndex(cLat, cLon)
			appendAttributesForIndex(cLat+1, cLon)
		}
	}

	planetRen.numTriangles = int32(len(points) / 3)
	if planetRen.numTriangles > 0 {
		fillVBO(planetRen.pointsVBO, points)
		fillVBO(planetRen.colorsVBO, colors)
	}
	planetRen.geometryUpdated = true
}

func (planetRen *Planet) drawGeometry() {
	if !planetRen.geometryUpdated {
		planetRen.updateGeometry()
	}
	planetRen.Planet.GeometryMutex.Lock()
	geom := planetRen.Planet.Geometry
	planetRen.Planet.GeometryMutex.Unlock()
	if geom == nil || geom.IsLoading {
		return
	}
	if planetRen.numTriangles > 0 {
		gl.BindVertexArray(planetRen.drawableVAO)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, planetRen.numTriangles)
	}
}
