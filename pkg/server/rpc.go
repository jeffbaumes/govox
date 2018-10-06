package server

import (
	"context"
	"errors"

	pb "github.com/jeffbaumes/govox/pkg/govox"
)

type server struct{}

func (s *server) GetPlanets(ctx context.Context, in *pb.GetPlanetsRequest) (*pb.GetPlanetsResponse, error) {
	planetSpecs := []*pb.PlanetSpec{}
	for _, planet := range universe.PlanetMap {
		planetSpecs = append(planetSpecs, &planet.Spec)
	}
	ret := pb.GetPlanetsResponse{Planets: planetSpecs}
	return &ret, nil
}

func (s *server) GetChunk(ctx context.Context, in *pb.GetChunkRequest) (*pb.GetChunkResponse, error) {
	planet := universe.PlanetMap[in.Planet]
	if planet == nil {
		return nil, errors.New("unknown planet ID")
	}
	chunk := planet.GetChunk(*in.Index, false)
	return &pb.GetChunkResponse{Chunk: chunk}, nil
}

func (s *server) SetCellMaterial(ctx context.Context, in *pb.SetCellMaterialRequest) (*pb.SetCellMaterialResponse, error) {
	planet := universe.PlanetMap[in.Planet]
	if planet == nil {
		return nil, errors.New("unknown planet ID")
	}
	planet.SetCellMaterial(*in.Index, in.Cell.Material, false)
	// var validPeople []*connectedPerson
	// for _, c := range api.connectedPeople {
	// 	var ret bool
	// 	e := c.rpc.Call("API.SetCellMaterial", args, &ret)
	// 	if e != nil {
	// 		if e.Error() == "connection is shut down" {
	// 			api.personDisconnected(c.state.Name)
	// 			continue
	// 		}
	// 		log.Println("SetCellMaterial error:", e)
	// 	}
	// 	validPeople = append(validPeople, c)
	// }
	// api.connectedPeople = validPeople
	return &pb.SetCellMaterialResponse{}, nil
}

// GetPlanetGeometry returns the low resolution geometry for a planet
func (s *server) GetPlanetGeometry(ctx context.Context, in *pb.GetPlanetGeometryRequest) (*pb.GetPlanetGeometryResponse, error) {
	planet := universe.PlanetMap[in.Planet]
	if planet == nil {
		return nil, errors.New("Unknown planet ID")
	}
	g := planet.GetGeometry(false)
	return &pb.GetPlanetGeometryResponse{Geometry: g}, nil
}

// HitPlayer damages a person
func (s *server) HitPlayer(ctx context.Context, in *pb.HitPlayerRequest) (*pb.HitPlayerResponse, error) {
	// var validPeople []*connectedPerson
	// for _, c := range api.connectedPeople {
	// 	if c.state.Name == args.Target {
	// 		var r bool
	// 		e := c.rpc.Call("API.HitPlayer", args, &r)
	// 		if e != nil {
	// 			if e.Error() == "connection is shut down" {
	// 				api.personDisconnected(c.state.Name)
	// 				continue
	// 			}
	// 			log.Println("HitPlayer error:", e)
	// 		}
	// 	}
	// 	validPeople = append(validPeople, c)
	// }
	// api.connectedPeople = validPeople
	// *ret = true
	return &pb.HitPlayerResponse{}, nil
}

// SendText sends a text to all players
func (s *server) SendText(ctx context.Context, in *pb.SendTextRequest) (*pb.SendTextResponse, error) {
	// var validPeople []*connectedPerson
	// for _, c := range api.connectedPeople {
	// 	var r bool
	// 	e := c.rpc.Call("API.SendText", text, &r)
	// 	if e != nil {
	// 		if e.Error() == "connection is shut down" {
	// 			api.personDisconnected(c.state.Name)
	// 			continue
	// 		}
	// 		log.Println("UpdatePersonState error:", e)
	// 	}
	// 	validPeople = append(validPeople, c)
	// }
	// api.connectedPeople = validPeople
	// *ret = true
	return &pb.SendTextResponse{}, nil
}

// UpdatePlayerState updates a person's position
func (s *server) UpdatePlayerState(ctx context.Context, in *pb.UpdatePlayerStateRequest) (*pb.UpdatePlayerStateResponse, error) {
	// var validPeople []*connectedPerson
	// for _, c := range api.connectedPeople {
	// 	if c.state.Name == state.Name {
	// 		c.state = *state
	// 	} else {
	// 		var r bool
	// 		e := c.rpc.Call("API.UpdatePersonState", state, &r)
	// 		if e != nil {
	// 			if e.Error() == "connection is shut down" {
	// 				api.personDisconnected(c.state.Name)
	// 				continue
	// 			}
	// 			log.Println("UpdatePersonState error:", e)
	// 		}
	// 	}
	// 	validPeople = append(validPeople, c)
	// }
	// api.connectedPeople = validPeople
	// *ret = true
	return &pb.UpdatePlayerStateResponse{}, nil
}

// // SetCellMaterial sets the material for a particular cell
// func (api *API) SetCellMaterial(args *common.RPCSetCellMaterialArgs, ret *bool) error {
// 	planet := universe.PlanetMap[args.Planet]
// 	if planet == nil {
// 		return errors.New("Unknown planet ID")
// 	}
// 	*ret = planet.SetCellMaterial(args.Index, args.Material, false)
// 	var validPeople []*connectedPerson
// 	for _, c := range api.connectedPeople {
// 		var ret bool
// 		e := c.rpc.Call("API.SetCellMaterial", args, &ret)
// 		if e != nil {
// 			if e.Error() == "connection is shut down" {
// 				api.personDisconnected(c.state.Name)
// 				continue
// 			}
// 			log.Println("SetCellMaterial error:", e)
// 		}
// 		validPeople = append(validPeople, c)
// 	}
// 	api.connectedPeople = validPeople
// 	return nil
// }

// func (api *API) personDisconnected(name string) {
// 	log.Printf("%v disconnected", name)
// 	for _, c := range api.connectedPeople {
// 		var ret bool
// 		c.rpc.Call("API.PersonDisconnected", name, &ret)
// 	}
// }
