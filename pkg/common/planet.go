package common

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"log"
	"math"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	pb "github.com/jeffbaumes/govox/pkg/govox"
	opensimplex "github.com/ojrac/opensimplex-go"
)

// ChunkSize is the number of cells per side of a chunk
const (
	ChunkSize = 16
)

// Planet represents all the cells in a spherical planet
type Planet struct {
	grpcClient    pb.GovoxClient
	db            *sql.DB
	Geometry      *pb.PlanetGeometry
	GeometryMutex *sync.Mutex
	Chunks        map[ChunkKey]*pb.Chunk
	databaseMutex *sync.Mutex
	ChunksMutex   *sync.Mutex
	noise         *opensimplex.Noise
	Generator     func(*Planet, pb.CellLoc) pb.Cell
	AltMin        float64
	AltDelta      float64
	LatMax        float64
	LonCells      int64
	LatCells      int64
	Spec          pb.PlanetSpec
}

// NewPlanet constructs a Planet instance
func NewPlanet(grpcClient pb.GovoxClient, db *sql.DB, spec pb.PlanetSpec) *Planet {
	p := Planet{}
	p.Spec = spec
	p.grpcClient = grpcClient
	p.noise = opensimplex.NewWithSeed(int64(p.Spec.Seed))
	p.Spec.AltCells = p.Spec.AltCells / ChunkSize * ChunkSize
	p.AltMin = p.Spec.Radius - float64(p.Spec.AltCells)
	p.AltDelta = 1.0
	p.LatMax = 90.0
	p.LonCells = int64(2.0*math.Pi*3.0/4.0*(0.5*p.Spec.Radius)+0.5) / ChunkSize * ChunkSize
	p.LatCells = int64(p.LatMax/90.0*math.Pi*(0.5*p.Spec.Radius)) / ChunkSize * ChunkSize
	p.Chunks = make(map[ChunkKey]*pb.Chunk)
	p.db = db
	p.databaseMutex = &sync.Mutex{}
	p.ChunksMutex = &sync.Mutex{}
	p.GeometryMutex = &sync.Mutex{}
	p.Generator = generators[p.Spec.GeneratorType]
	if p.Generator == nil {
		p.Generator = generators["sphere"]
	}
	return &p
}

// ChunkKey stores the latitude, longitude, and altitude index of a chunk
type ChunkKey struct {
	Lon, Lat, Alt int64
}

// GetChunk retrieves the chunk of a planet from chunk indices, either synchronously or asynchronously
func (p *Planet) GetChunk(ind pb.ChunkIndex, async bool) *pb.Chunk {
	if ind.Lon < 0 || ind.Lon >= p.LonCells/ChunkSize {
		return nil
	}
	if ind.Lat < 0 || ind.Lat >= p.LatCells/ChunkSize {
		return nil
	}
	if ind.Alt < 0 || ind.Alt >= p.Spec.AltCells/ChunkSize {
		return nil
	}

	p.ChunksMutex.Lock()
	key := ChunkKey{Lon: ind.Lon, Lat: ind.Lat, Alt: ind.Alt}
	chunk := p.Chunks[key]
	p.ChunksMutex.Unlock()

	if chunk != nil && chunk.WaitingForData {
		return nil
	}
	if chunk == nil {
		if p.grpcClient == nil {
			if p.db != nil {
				p.databaseMutex.Lock()
				rows, e := p.db.Query("SELECT data FROM chunk WHERE planet = ? AND lon = ? AND lat = ? AND alt = ?", p.Spec.Id, ind.Lon, ind.Lat, ind.Alt)
				if e != nil {
					panic(e)
				}
				if rows.Next() {
					var data []byte
					e = rows.Scan(&data)
					if e != nil {
						panic(e)
					}
					var dbuf bytes.Buffer
					dbuf.Write(data)
					dec := gob.NewDecoder(&dbuf)
					var ch pb.Chunk
					e = dec.Decode(&ch)
					if e != nil {
						panic(e)
					}
					chunk = &ch
				}
				rows.Close()
				p.databaseMutex.Unlock()
				if chunk == nil {
					chunk = newChunk(ind, p)
					p.databaseMutex.Lock()
					stmt, e := p.db.Prepare("INSERT INTO chunk VALUES (?, ?, ?, ?, ?)")
					if e != nil {
						panic(e)
					}
					var buf bytes.Buffer
					enc := gob.NewEncoder(&buf)
					e = enc.Encode(chunk)
					if e != nil {
						panic(e)
					}
					_, e = stmt.Exec(p.Spec.Id, ind.Lon, ind.Lat, ind.Alt, buf.Bytes())
					if e != nil {
						panic(e)
					}
					p.databaseMutex.Unlock()
				}
				p.ChunksMutex.Lock()
				p.Chunks[key] = chunk
				p.ChunksMutex.Unlock()
			} else {
				chunk = newChunk(ind, p)
				p.ChunksMutex.Lock()
				p.Chunks[key] = chunk
				p.ChunksMutex.Unlock()
			}
		} else {
			request := pb.GetChunkRequest{Planet: p.Spec.Id, Index: &ind}
			if async {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					response, err := p.grpcClient.GetChunk(ctx, &request)
					if err != nil {
						log.Fatalf("get chunk failed: %v", err)
					}
					p.ChunksMutex.Lock()
					p.Chunks[key] = response.Chunk
					p.ChunksMutex.Unlock()
				}()
				p.ChunksMutex.Lock()
				p.Chunks[key] = &pb.Chunk{WaitingForData: true}
				p.ChunksMutex.Unlock()
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				response, err := p.grpcClient.GetChunk(ctx, &request)
				if err != nil {
					log.Fatalf("get chunk failed: %v", err)
				}
				p.ChunksMutex.Lock()
				p.Chunks[key] = response.Chunk
				p.ChunksMutex.Unlock()
			}
		}
	}
	return chunk
}

// SetCellMaterial sets the contents of a cell
func (p *Planet) SetCellMaterial(ind pb.CellIndex, material pb.Material, updateServer bool) bool {
	cell := p.CellIndexToCell(ind)
	if cell == nil {
		return false
	}
	if cell.Material == material {
		return false
	}
	cell.Material = material
	if p.grpcClient != nil && updateServer {
		go func() {
			request := pb.SetCellMaterialRequest{Planet: p.Spec.Id, Index: &ind, Cell: &pb.Cell{Material: material}}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := p.grpcClient.SetCellMaterial(ctx, &request)
			if err != nil {
				log.Fatalf("set cell material failed: %v", err)
			}
		}()
	}
	if p.db != nil {
		chunkInd := p.CellIndexToChunkIndex(ind)
		chunk := p.CellIndexToChunk(ind)
		p.databaseMutex.Lock()
		stmt, e := p.db.Prepare("UPDATE chunk SET data = ? WHERE planet = 0 AND lon = ? AND lat = ? AND alt = ?")
		if e != nil {
			panic(e)
		}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		e = enc.Encode(chunk)
		if e != nil {
			panic(e)
		}
		_, e = stmt.Exec(buf.Bytes(), chunkInd.Lon, chunkInd.Lat, chunkInd.Alt)
		if e != nil {
			panic(e)
		}
		p.databaseMutex.Unlock()
	}

	return true
}

func (p *Planet) validateCellLoc(l pb.CellLoc) pb.CellLoc {
	if l.Lon < 0 {
		l.Lon += float64(p.LonCells)
	}
	for l.Lon >= float64(p.LonCells) {
		l.Lon -= float64(p.LonCells)
	}
	return l
}

// CellLocToChunk converts floating-point cell indices to a chunk
func (p *Planet) CellLocToChunk(l pb.CellLoc) *pb.Chunk {
	l = p.validateCellLoc(l)
	return p.CellIndexToChunk(p.CellLocToCellIndex(l))
}

// CellIndexToChunk converts a cell index to its containing chunk
func (p *Planet) CellIndexToChunk(cellIndex pb.CellIndex) *pb.Chunk {
	ind := p.CellIndexToChunkIndex(cellIndex)
	if ind.Lon < 0 || ind.Lon >= p.LonCells/ChunkSize {
		return nil
	}
	if ind.Lat < 0 || ind.Lat >= p.LatCells/ChunkSize {
		return nil
	}
	if ind.Alt < 0 || ind.Alt >= p.Spec.AltCells/ChunkSize {
		return nil
	}
	return p.GetChunk(ind, true)
}

// CellLocToChunkIndex converts floating-point cell indices to a chunk index
func (p *Planet) CellLocToChunkIndex(l pb.CellLoc) pb.ChunkIndex {
	l = p.validateCellLoc(l)
	return p.CellIndexToChunkIndex(p.CellLocToCellIndex(l))
}

// CellIndexToChunkIndex converts a cell index to its containing chunk index
func (p *Planet) CellIndexToChunkIndex(cellInd pb.CellIndex) pb.ChunkIndex {
	cs := float64(ChunkSize)
	return pb.ChunkIndex{
		Lon: int64(math.Floor(float64(cellInd.Lon) / cs)),
		Lat: int64(math.Floor(float64(cellInd.Lat) / cs)),
		Alt: int64(math.Floor(float64(cellInd.Alt) / cs)),
	}
}

// CellLocToCellIndex converts floating-point cell indices to a cell index
func (p *Planet) CellLocToCellIndex(l pb.CellLoc) pb.CellIndex {
	l = p.validateCellLoc(l)
	l = p.CellLocToNearestCellCenter(l)
	l = p.validateCellLoc(l)
	return pb.CellIndex{Lon: int64(l.Lon), Lat: int64(l.Lat), Alt: int64(l.Alt)}
}

// CellIndexToCellLoc converts a cell index to floating-point cell indices
func (p *Planet) CellIndexToCellLoc(l pb.CellIndex) pb.CellLoc {
	return pb.CellLoc{Lon: float64(l.Lon), Lat: float64(l.Lat), Alt: float64(l.Alt)}
}

// CartesianToChunkIndex converts world coordinates to a chunk index
func (p *Planet) CartesianToChunkIndex(cart mgl32.Vec3) pb.ChunkIndex {
	l := p.CartesianToCellLoc(cart)
	return p.CellLocToChunkIndex(l)
}

// CartesianToCellIndex converts world coordinates to a cell index
func (p *Planet) CartesianToCellIndex(cart mgl32.Vec3) pb.CellIndex {
	return p.CellLocToCellIndex(p.CartesianToCellLoc(cart))
}

// CartesianToChunk converts world coordinates to a chunk
func (p *Planet) CartesianToChunk(cart mgl32.Vec3) *pb.Chunk {
	return p.CellLocToChunk(p.CartesianToCellLoc(cart))
}

// CellLocToNearestCellCenter converts floating-point cell indices to the nearest integral indices
func (p *Planet) CellLocToNearestCellCenter(l pb.CellLoc) pb.CellLoc {
	l = p.validateCellLoc(l)
	return pb.CellLoc{
		Lon: float64(math.Floor(float64(l.Lon) + 0.5)),
		Lat: float64(math.Floor(float64(l.Lat) + 0.5)),
		Alt: float64(math.Floor(float64(l.Alt) + 0.5)),
	}
}

// CellLocToCell converts floating-point chunk indices to a cell
func (p *Planet) CellLocToCell(l pb.CellLoc) *pb.Cell {
	l = p.validateCellLoc(l)
	return p.CellIndexToCell(p.CellLocToCellIndex(l))
}

// CellIndexToCell converts a cell index to a cell
func (p *Planet) CellIndexToCell(cellIndex pb.CellIndex) *pb.Cell {
	chunkIndex := p.CellIndexToChunkIndex(cellIndex)
	lonCells, latCells := p.LonLatCellsInChunkIndex(chunkIndex)
	lonWidth := ChunkSize / lonCells
	latWidth := ChunkSize / latCells
	chunk := p.CellIndexToChunk(cellIndex)
	if chunk == nil {
		return nil
	}
	lonInd := (cellIndex.Lon % ChunkSize) / int64(lonWidth)
	latInd := (cellIndex.Lat % ChunkSize) / int64(latWidth)
	altInd := cellIndex.Alt % ChunkSize
	return chunk.Cell[lonInd].Cell[latInd].Cell[altInd]
}

// SphericalToCellLoc converts spherical coordinates to floating-point cell indices
func (p *Planet) SphericalToCellLoc(r, theta, phi float32) pb.CellLoc {
	alt := (r - float32(p.AltMin)) / float32(p.AltDelta)
	lat := (180*theta/math.Pi-90+float32(p.LatMax))*float32(p.LatCells)/(2*float32(p.LatMax)) - 0.5
	if phi < 0 {
		phi += 2 * math.Pi
	}
	lon := phi * float32(p.LonCells) / (2 * math.Pi)
	return pb.CellLoc{Lon: float64(lon), Lat: float64(lat), Alt: float64(alt)}
}

// CartesianToCell returns the cell contianing a set of world coordinates
func (p *Planet) CartesianToCell(cart mgl32.Vec3) *pb.Cell {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	l := p.SphericalToCellLoc(r, theta, phi)
	return p.CellLocToCell(l)
}

// CartesianToCellLoc converts world coordinates to floating-point cell indices
func (p *Planet) CartesianToCellLoc(cart mgl32.Vec3) pb.CellLoc {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	return p.SphericalToCellLoc(r, theta, phi)
}

// CellIndexToCartesian converts a cell index to world coordinates
func (p *Planet) CellIndexToCartesian(ind pb.CellIndex) mgl32.Vec3 {
	loc := p.CellIndexToCellLoc(ind)
	return p.CellLocToCartesian(loc)
}

// CellLocToCartesian converts floating-point cell indices to world coordinates
func (p *Planet) CellLocToCartesian(l pb.CellLoc) mgl32.Vec3 {
	l = p.validateCellLoc(l)
	r, theta, phi := p.CellLocToSpherical(l)
	return mgl32.SphericalToCartesian(r, theta, phi)
}

// CellLocToSpherical converts floating-point cell indices to spherical coordinates
func (p *Planet) CellLocToSpherical(l pb.CellLoc) (r, theta, phi float32) {
	l = p.validateCellLoc(l)
	r = float32(l.Alt)*float32(p.AltDelta) + float32(p.AltMin)
	theta = (math.Pi / 180) * ((90.0 - float32(p.LatMax)) + ((float32(l.Lat)+0.5)/float32(p.LatCells))*(2.0*float32(p.LatMax)))
	phi = 2 * math.Pi * float32(l.Lon) / float32(p.LonCells)
	return
}

// LonLatCellsInChunkIndex returns the number of longitude and latitude cells in a chunk, which changes based on latitude and altitude
func (p *Planet) LonLatCellsInChunkIndex(ind pb.ChunkIndex) (lonCells, latCells int) {
	lonCells = ChunkSize
	latCells = ChunkSize

	// If chunk is too close to the poles, lower the longitude cells per chunk
	theta := (90.0 - float32(p.LatMax) + (float32(ind.Lat)+0.5)*float32(ChunkSize)/float32(p.LatCells)) * (2.0 * float32(p.LatMax))
	if math.Abs(float64(theta-90)) >= 60 {
		lonCells /= 2
	}
	if math.Abs(float64(theta-90)) >= 80 {
		lonCells /= 2
	}

	// If chunk is too close to the center of the planet, lower both lon and lat cells per chunk
	if (float64(ind.Alt)+0.5)*ChunkSize < p.Spec.Radius/4 {
		lonCells /= 2
		latCells /= 2
	}
	if (float64(ind.Alt)+0.5)*ChunkSize < p.Spec.Radius/8 {
		lonCells /= 2
		latCells /= 2
	}

	return
}

func newChunk(ind pb.ChunkIndex, p *Planet) *pb.Chunk {
	chunk := pb.Chunk{}
	lonCells, latCells := p.LonLatCellsInChunkIndex(ind)
	lonWidth := ChunkSize / lonCells
	latWidth := ChunkSize / latCells
	chunk.Cell = make([]*pb.Chunk_CellLat, lonCells)
	for lonIndex := 0; lonIndex < lonCells; lonIndex++ {
		chunk.Cell[lonIndex] = &pb.Chunk_CellLat{}
		chunk.Cell[lonIndex].Cell = make([]*pb.Chunk_CellAlt, latCells)
		for latIndex := 0; latIndex < latCells; latIndex++ {
			chunk.Cell[lonIndex].Cell[latIndex] = &pb.Chunk_CellAlt{}
			chunk.Cell[lonIndex].Cell[latIndex].Cell = make([]*pb.Cell, ChunkSize)
			for altIndex := 0; altIndex < ChunkSize; altIndex++ {
				l := pb.CellLoc{
					Lon: float64(int(ChunkSize*ind.Lon) + lonIndex*lonWidth),
					Lat: float64(int(ChunkSize*ind.Lat) + latIndex*latWidth),
					Alt: float64(int(ChunkSize*ind.Alt) + altIndex),
				}
				c := p.Generator(p, l)

				// Always give the planet a solid core
				if l.Alt < 2 {
					c.Material = pb.Material_STONE
				}

				chunk.Cell[lonIndex].Cell[latIndex].Cell[altIndex] = &c
			}
		}
	}
	return &chunk
}

// List of materials
var (
	Materials = []string{
		"air",
		"grass",
		"dirt",
		"stone",
		"moon",
		"asteroid",
		"sun",
		"blue_block",
		"blue_sand",
		"purple_block",
		"purple_sand",
		"red_block",
		"red_sand",
		"yellow_block",
		"yellow_sand",
		"water",
	}
	MaterialColors = []mgl32.Vec3{
		{0.0, 0.0, 0.0},
		{0.5, 1.0, 0.5},
		{0.5, 0.3, 0.0},
		{0.5, 0.5, 0.5},
		{0.7, 0.7, 0.7},
		{0.4, 0.4, 0.4},
		{1.0, 0.9, 0.5},
		{0.5, 0.5, 1.0},
		{0.5, 0.5, 1.0},
		{1.0, 0.0, 1.0},
		{1.0, 0.0, 1.0},
		{1.0, 0.5, 0.5},
		{1.0, 0.5, 0.5},
		{1.0, 1.0, 0.0},
		{1.0, 1.0, 0.0},
		{0.0, 0.0, 0.0},
	}
)

func (p *Planet) generateGeometry() *pb.PlanetGeometry {
	geom := pb.PlanetGeometry{}
	lonCells := 64
	latCells := 32 + 1
	geom.Material = make([]*pb.PlanetGeometry_MaterialRow, lonCells)
	geom.Altitude = make([]*pb.PlanetGeometry_AltitudeRow, lonCells)
	for lon := 0; lon < lonCells; lon++ {
		geom.Material[lon] = &pb.PlanetGeometry_MaterialRow{}
		geom.Material[lon].Material = make([]pb.Material, latCells)
		geom.Altitude[lon] = &pb.PlanetGeometry_AltitudeRow{}
		geom.Altitude[lon].Altitude = make([]int64, latCells)
		for lat := 0; lat < latCells; lat++ {
			lonInd := math.Floor(float64(p.LonCells) * float64(lon) / float64(lonCells))

			// Make sure latitude hits both poles, hence the need for division by (latCells - 1)
			latInd := math.Floor(float64(p.LatCells) * float64(lat) / float64(latCells-1))

			loc := pb.CellLoc{Lon: float64(lonInd), Lat: float64(latInd), Alt: float64(p.Spec.AltCells - 1)}
			cell := p.Generator(p, loc)
			for cell.Material == pb.Material_AIR && loc.Alt > 0 {
				loc.Alt--
				cell = p.Generator(p, loc)
			}
			geom.Material[lon].Material[lat] = cell.Material
			geom.Altitude[lon].Altitude[lat] = int64(loc.Alt)
		}
	}
	return &geom
}

// GetGeometry returns the low-resultion geometry for the planet.
func (p *Planet) GetGeometry(async bool) *pb.PlanetGeometry {
	if p.Geometry != nil && p.Geometry.IsLoading {
		return nil
	}
	if p.Geometry != nil {
		return p.Geometry
	}
	if p.grpcClient != nil {
		request := pb.GetPlanetGeometryRequest{Planet: p.Spec.Id}
		if async {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				response, err := p.grpcClient.GetPlanetGeometry(ctx, &request)
				if err != nil {
					log.Fatalf("failed to get geometry: %v", err)
				}
				p.GeometryMutex.Lock()
				p.Geometry = response.Geometry
				p.GeometryMutex.Unlock()
			}()
			p.GeometryMutex.Lock()
			p.Geometry = &pb.PlanetGeometry{IsLoading: true}
			p.GeometryMutex.Unlock()
			return p.Geometry
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		response, err := p.grpcClient.GetPlanetGeometry(ctx, &request)
		if err != nil {
			log.Fatalf("failed to get geometry: %v", err)
		}
		p.GeometryMutex.Lock()
		p.Geometry = response.Geometry
		p.GeometryMutex.Unlock()
		return p.Geometry
	}
	p.Geometry = p.generateGeometry()
	return p.Geometry
}
