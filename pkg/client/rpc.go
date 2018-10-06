package client

// // API is the RPC tag for client calls
// type API int

// // SetCellMaterial sets the material for a particular cell
// func (api *API) SetCellMaterial(args *common.RPCSetCellMaterialArgs, ret *bool) error {
// 	universe.PlanetMap[args.Planet].SetCellMaterial(args.Index, args.Material, false)
// 	*ret = true
// 	return nil
// }

// // GetPersonState returns this client's logged in user state
// func (api *API) GetPersonState(args *int, ret *common.PlayerState) error {
// 	ret.Name = universe.Player.Name
// 	ret.Position = universe.Player.Location()
// 	return nil
// }

// // PersonDisconnected notifies a client that a player has disconnected
// func (api *API) PersonDisconnected(name *string, ret *bool) error {
// 	var validPeople []*common.PlayerState
// 	*ret = false
// 	for _, p := range universe.ConnectedPeople {
// 		if p.Name != *name {
// 			validPeople = append(validPeople, p)
// 			*ret = true
// 		}
// 	}
// 	universe.ConnectedPeople = validPeople
// 	return nil
// }

// // UpdatePersonState updates another person's state
// func (api *API) UpdatePersonState(state *common.PlayerState, ret *bool) error {
// 	if state.Name == universe.Player.Name {
// 		return nil
// 	}
// 	found := false
// 	for _, c := range universe.ConnectedPeople {
// 		if c.Name == state.Name {
// 			c.Position = state.Position
// 			c.LookDir = state.LookDir
// 			found = true
// 			break
// 		}
// 	}
// 	if !found {
// 		universe.ConnectedPeople = append(universe.ConnectedPeople, state)
// 	}
// 	return nil
// }

// // SendText sends a player text
// func (api *API) SendText(text *string, ret *bool) error {
// 	universe.Player.DrawText = *text
// 	*ret = true
// 	return nil
// }

// // HitPlayer damages a playerv
// func (api *API) HitPlayer(args *common.HitPlayerArgs, ret *bool) error {
// 	log.Println(fmt.Sprintf("Hit by %v", args.From))
// 	universe.Player.UpdateHealth(-args.Amount)
// 	*ret = true
// 	return nil
// }
