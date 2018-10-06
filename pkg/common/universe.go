package common

import (
	"bytes"
	"database/sql"
	"encoding/gob"

	pb "github.com/jeffbaumes/govox/pkg/govox"
	opensimplex "github.com/ojrac/opensimplex-go"
)

// Universe stores the set of planets in a universe
type Universe struct {
	seed      int64
	noise     *opensimplex.Noise
	PlanetMap map[int64]*Planet
}

// NewUniverse creates a universe with a given seed
func NewUniverse(db *sql.DB, systemType string) *Universe {
	u := Universe{}
	u.noise = opensimplex.NewWithSeed(0)
	u.PlanetMap = make(map[int64]*Planet)
	planetSpecs := queryPlanetSpecs(db)

	// If no planets in the database, generate a planetary system
	if len(planetSpecs) == 0 {
		systemGen := systems[systemType]
		if systemGen == nil {
			systemGen = systems["planet"]
		}
		planetSpecs = systemGen()
		for _, spec := range planetSpecs {
			savePlanetSpec(db, *spec)
		}
	}

	// Put the planets in the universe
	for _, spec := range planetSpecs {
		planet := NewPlanet(nil, db, *spec)
		u.PlanetMap[planet.Spec.Id] = planet
	}

	return &u
}

func queryPlanetSpecs(db *sql.DB) []*pb.PlanetSpec {
	states := []*pb.PlanetSpec{}
	rows, err := db.Query("SELECT data FROM planet")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var val pb.PlanetSpec
		var data []byte
		err = rows.Scan(&data)
		if err != nil {
			panic(err)
		}
		var dbuf bytes.Buffer
		dbuf.Write(data)
		dec := gob.NewDecoder(&dbuf)
		err = dec.Decode(&val)
		if err != nil {
			panic(err)
		}
		states = append(states, &val)
	}
	return states
}

func savePlanetSpec(db *sql.DB, spec pb.PlanetSpec) {
	stmt, err := db.Prepare("INSERT INTO planet VALUES (?, ?)")
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(spec)
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec(spec.Id, buf.Bytes())
	if err != nil {
		panic(err)
	}
}
