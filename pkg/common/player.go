package common

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	pb "github.com/jeffbaumes/govox/pkg/govox"
)

// Game modes
const (
	Normal           = iota
	Flying           = iota
	NumMovementModes = iota
)

// Creative or survival modes
const (
	Creative = iota
	Survival = iota
)

// Maximum health
const (
	MaxHealth = 10
)

// Player represents a player of the game
type Player struct {
	Planet           *Planet
	UpVel            float32
	DownVel          float32
	ForwardVel       float32
	BackVel          float32
	RightVel         float32
	LeftVel          float32
	FallVel          float32
	WalkVel          float32
	loc              mgl32.Vec3
	lookHeading      mgl32.Vec3
	lookAltitude     float64
	height           float64
	radius           float64
	MovementMode     int
	GameMode         int
	HoldingJump      bool
	inJump           bool
	Name             string
	ActiveHotBarSlot int
	FocusCellIndex   pb.CellIndex
	HotbarOn         bool
	Hotbar           [12]Slot
	Inventory        [48]Slot
	renderDistance   int
	Health           int
	Text             string
	DrawText         string
	Mode             string
}

// Slot is an inventory slot
type Slot struct {
	Material pb.Material
	Amount   int
}

// HitPlayerArgs are the arguments for the HitPlayer API call
type HitPlayerArgs struct {
	From   string
	Target string
	Amount int
}

// NewPlayer creates a new player
func NewPlayer(name string) *Player {
	p := Player{}
	p.WalkVel = 5.0
	p.height = 2
	p.radius = 0.25
	p.MovementMode = Normal
	p.GameMode = Survival
	p.Name = name
	p.ActiveHotBarSlot = 0
	p.HotbarOn = true
	p.renderDistance = 4
	return &p
}

// Spawn the player on their current planet spawn
func (player *Player) Spawn() {
	player.lookHeading = mgl32.Vec3{0, 1, 0}
	player.Health = MaxHealth
	player.UpVel = 0
	player.DownVel = 0
	player.ForwardVel = 0
	player.BackVel = 0
	player.RightVel = 0
	player.LeftVel = 0
	player.FallVel = 0
	loc := mgl32.Vec3{float32(player.Planet.Spec.Radius) + 5, 0, 0}
	player.loc = loc

	// Make sure the spawn area is ready (not async)
	player.LoadNearbyChunks(false)

	// Find a non-air place to land
	c := player.Planet.CartesianToCell(loc)
	for (c == nil || c.Material == pb.Material_AIR) && loc[0] > 0 {
		loc[0]--
		c = player.Planet.CartesianToCell(loc)
	}
	loc[0] += 5
	player.loc = loc
}

// Location returns the location of the player.
func (player *Player) Location() mgl32.Vec3 {
	if player.Mode == "Apex" {
		return mgl32.Vec3{0, 0, 750}
	}
	return player.loc
}

// SetLocation sets the location of the player.
func (player *Player) SetLocation(loc mgl32.Vec3) {
	if !(player.Mode == "Apex") {
		player.loc = loc
	}
}

// UpdateHealth updates a player health by a certain amount
func (player *Player) UpdateHealth(amount int) {
	player.Health += amount
	if player.Health <= 0 {
		player.Spawn()
	}
	if player.Health > MaxHealth {
		player.Health = MaxHealth
	}
}

// LookDir returns the player's look direction
func (player *Player) LookDir() mgl32.Vec3 {
	up := player.Location().Normalize()
	player.lookHeading = ProjectToPlane(player.lookHeading, up).Normalize()
	right := player.lookHeading.Cross(up)
	return mgl32.QuatRotate(float32((player.lookAltitude-90.0)*math.Pi/180.0), right).Rotate(up)
}

// Swivel swivels the player's direction based on mouse movement
func (player *Player) Swivel(deltaX float64, deltaY float64) {
	lookHeadingDelta := -0.1 * deltaX
	normalDir := player.Location().Normalize()
	player.lookHeading = mgl32.QuatRotate(float32(lookHeadingDelta*math.Pi/180.0), normalDir).Rotate(player.lookHeading)
	player.lookAltitude = player.lookAltitude - 0.1*deltaY
	player.lookAltitude = math.Max(math.Min(player.lookAltitude, 89.9), -89.9)
}

// LoadNearbyChunks loads the chunks around the player, either synchronously or asynchronously
func (player *Player) LoadNearbyChunks(async bool) {
	planet := player.Planet
	up := player.Location().Normalize()
	feet := player.Location().Sub(up.Mul(float32(player.height)))
	ind := planet.CartesianToChunkIndex(feet)
	for lon := int(ind.Lon) - player.renderDistance; lon <= int(ind.Lon)+player.renderDistance; lon++ {
		validLon := lon
		for validLon < 0 {
			validLon += int(planet.LonCells) / ChunkSize
		}
		for validLon >= int(planet.LonCells)/ChunkSize {
			validLon -= int(planet.LonCells) / ChunkSize
		}
		latMin := Max(int(ind.Lat)-player.renderDistance, 0)
		latMax := Min(int(ind.Lat)+player.renderDistance, int(planet.LatCells)/ChunkSize-1)
		for lat := latMin; lat <= latMax; lat++ {
			for alt := 0; alt < int(planet.Spec.AltCells)/ChunkSize; alt++ {
				planet.GetChunk(pb.ChunkIndex{Lon: int64(validLon), Lat: int64(lat), Alt: int64(alt)}, async)
			}
		}
	}
}

// UpdatePosition updates the player position
func (player *Player) UpdatePosition(h float32) {
	planet := player.Planet
	player.LoadNearbyChunks(true)
	if h > 0.05 {
		h = 0.05
	}

	up := player.Location().Normalize()
	right := player.lookHeading.Cross(up)
	if player.MovementMode == Normal {
		feet := player.Location().Sub(up.Mul(float32(player.height)))
		feetCell := planet.CartesianToCell(feet)
		falling := feetCell == nil || feetCell.Material == pb.Material_AIR
		if falling {
			player.FallVel -= 20 * h
		} else if player.HoldingJump && !player.inJump {
			player.FallVel = 7
			player.inJump = true
		} else {
			player.FallVel = 0
			player.inJump = false
		}

		playerVel := mgl32.Vec3{}
		playerVel = playerVel.Add(up.Mul(player.FallVel))
		playerVel = playerVel.Add(player.lookHeading.Mul((player.ForwardVel - player.BackVel)))
		playerVel = playerVel.Add(right.Mul((player.RightVel - player.LeftVel)))

		player.SetLocation(player.Location().Add(playerVel.Mul(h)))
		for height := planet.AltDelta / 2; height < player.height; height += planet.AltDelta {
			player.collide(planet, float32(height), pb.CellLoc{Lon: 0, Lat: 0, Alt: -1})
			player.collide(planet, float32(height), pb.CellLoc{Lon: 1, Lat: 0, Alt: 0})
			player.collide(planet, float32(height), pb.CellLoc{Lon: -1, Lat: 0, Alt: 0})
			player.collide(planet, float32(height), pb.CellLoc{Lon: 0, Lat: 1, Alt: 0})
			player.collide(planet, float32(height), pb.CellLoc{Lon: 0, Lat: -1, Alt: 0})
		}
	} else if player.MovementMode == Flying {
		LookDir := player.LookDir()
		player.SetLocation(player.Location().Add(up.Mul((player.UpVel - player.DownVel) * h)))
		player.SetLocation(player.Location().Add(LookDir.Mul((player.ForwardVel - player.BackVel) * h)))
		player.SetLocation(player.Location().Add(right.Mul((player.RightVel - player.LeftVel) * h)))
	}

	// Update focused cell
	increment := player.LookDir().Mul(0.05)
	pos := player.Location()
	player.FocusCellIndex = pb.CellIndex{Lat: 0, Lon: 0, Alt: 0}
	for i := 0; i < 100; i++ {
		pos = pos.Add(increment)
		cell := planet.CartesianToCell(pos)
		if cell != nil && cell.Material != pb.Material_AIR {
			cellIndex := planet.CartesianToCellIndex(pos)
			player.FocusCellIndex = cellIndex
			break
		}
	}
}

func (player *Player) collide(p *Planet, height float32, d pb.CellLoc) {
	up := player.Location().Normalize()
	pos := player.Location().Sub(up.Mul(float32(player.height) - height))
	l := p.CartesianToCellLoc(pos)
	c := p.CellLocToNearestCellCenter(l)
	adjCell := p.CellLocToCell(pb.CellLoc{
		Lon: c.Lon + d.Lon,
		Lat: c.Lat + d.Lat,
		Alt: c.Alt + d.Alt,
	})
	if adjCell != nil && adjCell.Material != pb.Material_AIR {
		if d.Alt != 0 {
			nLoc := p.CellLocToCartesian(pb.CellLoc{
				Lon: c.Lon + d.Lon/2,
				Lat: c.Lat + d.Lat/2,
				Alt: c.Alt + d.Alt/2,
			})
			distToPlane := up.Dot(pos.Sub(nLoc))
			if distToPlane < 0 {
				move := -distToPlane
				player.SetLocation(player.Location().Add(up.Mul(move)))
			}
		} else {
			nLoc := p.CellLocToCartesian(pb.CellLoc{
				Lon: c.Lon + d.Lon/2,
				Lat: c.Lat + d.Lat/2,
				Alt: c.Alt + d.Alt/2,
			})
			aLoc := p.CellLocToCartesian(pb.CellLoc{
				Lon: c.Lon + d.Lon,
				Lat: c.Lat + d.Lat,
				Alt: c.Alt + d.Alt,
			})
			cNorm := nLoc.Sub(aLoc).Normalize()
			cNorm = cNorm.Sub(Project(cNorm, up)).Normalize()
			distToPlane := cNorm.Dot(pos.Sub(nLoc))
			if distToPlane < float32(player.radius) {
				move := float32(player.radius) - distToPlane
				player.SetLocation(player.Location().Add(cNorm.Mul(move)))
			}
		}
	}
}
