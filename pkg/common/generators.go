package common

import pb "github.com/jeffbaumes/govox/pkg/govox"

var (
	generators map[string](func(*Planet, pb.CellLoc) pb.Cell)
	systems    map[string](func() []*pb.PlanetSpec)
)

func init() {
	generators = make(map[string](func(*Planet, pb.CellLoc) pb.Cell))

	generators["sphere"] = func(p *Planet, loc pb.CellLoc) pb.Cell {
		if float64(loc.Alt)/float64(p.Spec.AltCells) < 0.5 {
			return pb.Cell{Material: pb.Material_STONE}
		}
		return pb.Cell{Material: pb.Material_AIR}
	}
	generators["moon"] = func(p *Planet, loc pb.CellLoc) pb.Cell {
		if float64(loc.Alt)/float64(p.Spec.AltCells) < 0.5 {
			return pb.Cell{Material: pb.Material_MOON}
		}
		return pb.Cell{Material: pb.Material_AIR}
	}
	generators["sun"] = func(p *Planet, loc pb.CellLoc) pb.Cell {
		if float64(loc.Alt)/float64(p.Spec.AltCells) < 0.5 {
			return pb.Cell{Material: pb.Material_SUN}
		}
		return pb.Cell{Material: pb.Material_AIR}
	}

	generators["rings"] = func(p *Planet, loc pb.CellLoc) pb.Cell {
		scale := 1.0
		n := p.noise.Eval2(float64(loc.Alt)*scale, 0)
		fracHeight := float64(loc.Alt) / float64(p.Spec.AltCells)
		if fracHeight < 0.5 {
			return pb.Cell{Material: pb.Material_GRASS}
		}
		if fracHeight > 0.6 && int64(loc.Lat) == p.LatCells/2 {
			if n > 0.1 {
				return pb.Cell{Material: pb.Material_YELLOW_BLOCK}
			}
			return pb.Cell{Material: pb.Material_RED_BLOCK}
		}
		return pb.Cell{Material: pb.Material_AIR}
	}

	generators["bumpy"] = func(p *Planet, loc pb.CellLoc) pb.Cell {
		pos := p.CellLocToCartesian(loc).Normalize().Mul(float32(p.Spec.AltCells / 2))
		scale := 0.1
		height := float64(p.Spec.AltCells)/2 + p.noise.Eval3(float64(pos[0])*scale, float64(pos[1])*scale, float64(pos[2])*scale)*8
		if float64(loc.Alt) <= height {
			if float64(loc.Alt) > float64(p.Spec.AltCells)/2+2 {
				return pb.Cell{Material: pb.Material_DIRT}
			}
			return pb.Cell{Material: pb.Material_GRASS}
		}
		if float64(loc.Alt) < float64(p.Spec.AltCells)/2+1 {
			return pb.Cell{Material: pb.Material_BLUE_BLOCK}
		}
		return pb.Cell{Material: pb.Material_AIR}
	}

	generators["caves"] = func(p *Planet, loc pb.CellLoc) pb.Cell {
		pos := p.CellLocToCartesian(loc)
		const scale = 0.05
		height := (p.noise.Eval3(float64(pos[0])*scale, float64(pos[1])*scale, float64(pos[2])*scale) + 1.0) * float64(p.Spec.AltCells) / 2.0
		if height > float64(p.Spec.AltCells)/2 {
			return pb.Cell{Material: pb.Material_STONE}
		}
		return pb.Cell{Material: pb.Material_AIR}
	}

	generators["rocks"] = func(p *Planet, loc pb.CellLoc) pb.Cell {
		pos := p.CellLocToCartesian(loc)
		const scale = 0.05
		noise := p.noise.Eval3(float64(pos[0])*scale, float64(pos[1])*scale, float64(pos[2])*scale)
		if noise > 0.5 {
			return pb.Cell{Material: pb.Material_STONE}
		}
		return pb.Cell{Material: pb.Material_AIR}
	}

	systems = make(map[string](func() []*pb.PlanetSpec))

	systems["planet"] = func() []*pb.PlanetSpec {
		return []*pb.PlanetSpec{
			&pb.PlanetSpec{
				Id:              0,
				Name:            "Spawn",
				GeneratorType:   "bumpy",
				Radius:          64.0,
				AltCells:        64,
				RotationSeconds: 10,
			},
		}
	}

	systems["moon"] = func() []*pb.PlanetSpec {
		return []*pb.PlanetSpec{
			&pb.PlanetSpec{
				Id:              0,
				Name:            "Spawn",
				GeneratorType:   "bumpy",
				Radius:          64.0,
				AltCells:        64,
				RotationSeconds: 10,
			},
			&pb.PlanetSpec{
				Id:              1,
				Name:            "Moon",
				GeneratorType:   "moon",
				Radius:          32.0,
				AltCells:        32,
				OrbitPlanet:     0,
				OrbitDistance:   100,
				OrbitSeconds:    5,
				RotationSeconds: 10,
			},
		}
	}

	systems["sun-moon"] = func() []*pb.PlanetSpec {
		return []*pb.PlanetSpec{
			&pb.PlanetSpec{
				Id:              0,
				Name:            "Spawn",
				GeneratorType:   "bumpy",
				Radius:          64.0,
				AltCells:        64,
				OrbitPlanet:     2,
				OrbitDistance:   300,
				OrbitSeconds:    1095,
				RotationSeconds: 180,
			},
			&pb.PlanetSpec{
				Id:              1,
				Name:            "Moon",
				GeneratorType:   "moon",
				Radius:          32.0,
				AltCells:        32,
				OrbitPlanet:     0,
				OrbitDistance:   100,
				OrbitSeconds:    90,
				RotationSeconds: -90,
			},
			&pb.PlanetSpec{
				Id:              2,
				Name:            "Sun",
				GeneratorType:   "sun",
				Radius:          64.0,
				AltCells:        64,
				OrbitPlanet:     2,
				RotationSeconds: 1e10,
			},
		}
	}

	systems["many"] = func() []*pb.PlanetSpec {
		planets := []*pb.PlanetSpec{
			&pb.PlanetSpec{
				Id:              0,
				Name:            "Sun",
				GeneratorType:   "sun",
				Radius:          64.0,
				AltCells:        64,
				OrbitPlanet:     0,
				RotationSeconds: 1e10,
			},
		}
		for i := 0; i < 100; i++ {
			planets = append(planets, &pb.PlanetSpec{
				Id:              int64(2*i + 1),
				Name:            "Spawn",
				GeneratorType:   "sphere",
				Radius:          32.0,
				AltCells:        32,
				OrbitPlanet:     0,
				OrbitDistance:   70 * float64(i+1),
				OrbitSeconds:    10 + float64(i),
				RotationSeconds: 1e10,
			})
			planets = append(planets, &pb.PlanetSpec{
				Id:              int64(2*i + 2),
				Name:            "Spawn",
				GeneratorType:   "sphere",
				Radius:          16.0,
				AltCells:        16,
				OrbitPlanet:     int64(2*i + 1),
				OrbitDistance:   30,
				OrbitSeconds:    5,
				RotationSeconds: 1e10,
			})
		}
		return planets
	}

}
