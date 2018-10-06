package client

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/jeffbaumes/govox/pkg/common"
	pb "github.com/jeffbaumes/govox/pkg/govox"
	"github.com/jeffbaumes/govox/pkg/scene"
)

func cursorGrabbed(w *glfw.Window) bool {
	return w.GetInputMode(glfw.CursorMode) == glfw.CursorDisabled
}

func cursorPosCallback() func(w *glfw.Window, xpos, ypos float64) {
	lastCursor := mgl64.Vec2{math.NaN(), math.NaN()}
	return func(w *glfw.Window, xpos, ypos float64) {
		guicusorposcallback := screen.CursorPosCallback()
		guicusorposcallback(w, xpos, ypos)
		if !cursorGrabbed(w) {
			lastCursor = mgl64.Vec2{math.NaN(), math.NaN()}
			return
		}
		curCursor := mgl64.Vec2{xpos, ypos}
		if math.IsNaN(lastCursor[0]) {
			lastCursor = curCursor
		}
		delta := curCursor.Sub(lastCursor)
		universe.Player.Swivel(delta[0], delta[1])
		lastCursor = curCursor
	}
}

func glToPixel(w *glfw.Window, xpos, ypos float64) (xpix, ypix float64) {
	winw, winh := w.GetSize()
	return float64(winw) * (xpos + 1) / 2, float64(winh) * (-ypos + 1) / 2
}

// func sendText(text string) {
// 	if text == "enter" {
// 		universe.Player.Intext = false
// 		// NewText()
// 		var ret bool
// 		universe.RPC.Go("API.SendText", universe.Player.Text, &ret, nil)
// 		universe.Player.Text = ""
// 		// need to send text to server
// 	} else if text == "delete" {
// 		if len(universe.Player.Text) != 0 {
// 			universe.Player.Text = universe.Player.Text[0:(len(universe.Player.Text) - 1)]
// 		}
// 	} else {
// 		universe.Player.Text = universe.Player.Text + text
// 	}
// }

var (
	oelapsed   = time.Duration(800000)
	ostart     = time.Now()
	guielapsed = time.Duration(800000)
	guistart   = time.Now()
)

func keyCallbackInventory(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Press {
		return
	}
	player := universe.Player
	slot := -1
	m := op.OptionMap
	switch key {
	case m["Slot1"].Key:
		slot = 0
	case m["Slot2"].Key:
		slot = 1
	case m["Slot3"].Key:
		slot = 2
	case m["Slot4"].Key:
		slot = 3
	case m["Slot5"].Key:
		slot = 4
	case m["Slot6"].Key:
		slot = 5
	case m["Slot7"].Key:
		slot = 6
	case m["Slot8"].Key:
		slot = 7
	case m["Slot9"].Key:
		slot = 8
	case m["Slot10"].Key:
		slot = 9
	case m["Slot11"].Key:
		slot = 10
	case m["Slot12"].Key:
		slot = 11
	case m["Inventory"].Key:
		player.Mode = "play"
	}
	if slot < 0 {
		return
	}
	xpos, ypos := w.GetCursorPos()
	winw, winh := w.GetSize()
	aspect := float32(winw) / float32(winh)
	if player.GameMode == common.Creative {
		for m := range common.Materials {
			sz := float32(0.03)
			px := 1.25 * 2 * sz * (float32(m) - float32(len(common.Materials))/2)
			py := 1 - 0.25*aspect
			scale := sz
			xMin, yMin := glToPixel(w, float64(px-scale), float64(py+scale*aspect))
			xMax, yMax := glToPixel(w, float64(px+scale), float64(py-scale*aspect))
			if float64(xpos) >= xMin && float64(xpos) <= xMax && float64(ypos) >= yMin && float64(ypos) <= yMax {
				player.Hotbar[slot] = common.Slot{Material: pb.Material(m), Amount: 10}
			}
		}
	} else if player.GameMode == common.Survival {
		for row := 0; row < 1; row++ {
			for col := 0; col < 12; col++ {
				slotInd := row*12 + col
				sz := float32(0.03)
				px := 1.25 * 2 * sz * (float32(col) - float32(12)/2)
				py := 1 - 0.25*aspect
				scale := sz
				xMin, yMin := glToPixel(w, float64(px-scale), float64(py+scale*aspect))
				xMax, yMax := glToPixel(w, float64(px+scale), float64(py-scale*aspect))
				if float64(xpos) >= xMin && float64(xpos) <= xMax && float64(ypos) >= yMin && float64(ypos) <= yMax {
					tmp := player.Hotbar[slot]
					player.Hotbar[slot] = player.Inventory[slotInd]
					player.Inventory[slotInd] = tmp
				}
			}
		}
	}
}

func keyCallbackPlay(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	player := universe.Player
	planet := player.Planet
	planetRen := universe.PlanetMap[planet.Spec.Id]
	m := op.OptionMap
	switch action {
	case glfw.Press:
		switch key {
		case glfw.KeyEscape:
			player.Mode = "Options"
			w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		case m["Apex"].Key:
			player.Mode = "Apex"
		case m["Text"].Key:
			player.Mode = "Text"
		case m["Inventory"].Key:
			player.Mode = "Inventory"
			w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		case m["Forward"].Key:
			player.ForwardVel = player.WalkVel
		case m["Backward"].Key:
			player.BackVel = player.WalkVel
		case m["Right"].Key:
			player.RightVel = player.WalkVel
		case m["Left"].Key:
			player.LeftVel = player.WalkVel
		case m["Mode"].Key:
			player.MovementMode++
			if player.MovementMode >= common.NumMovementModes {
				player.MovementMode = 0
			}
			if player.MovementMode == common.Flying {
				player.FallVel = 0
			}
		case m["Slot1"].Key:
			player.ActiveHotBarSlot = 0
		case m["Slot2"].Key:
			player.ActiveHotBarSlot = 1
		case m["Slot3"].Key:
			player.ActiveHotBarSlot = 2
		case m["Slot4"].Key:
			player.ActiveHotBarSlot = 3
		case m["Slot5"].Key:
			player.ActiveHotBarSlot = 4
		case m["Slot6"].Key:
			player.ActiveHotBarSlot = 5
		case m["Slot7"].Key:
			player.ActiveHotBarSlot = 6
		case m["Slot8"].Key:
			player.ActiveHotBarSlot = 7
		case m["Slot9"].Key:
			player.ActiveHotBarSlot = 8
		case m["Slot10"].Key:
			player.ActiveHotBarSlot = 9
		case m["Slot11"].Key:
			player.ActiveHotBarSlot = 10
		case m["Slot12"].Key:
			player.ActiveHotBarSlot = 11
		case m["Up"].Key:
			if cursorGrabbed(w) {
				if player.MovementMode == common.Normal {
					player.HoldingJump = true
				} else {
					player.UpVel = player.WalkVel
				}
			}
		case m["Down"].Key:
			if cursorGrabbed(w) && player.MovementMode == common.Flying {
				player.DownVel = player.WalkVel
			}
		case m["PlanetR"].Key:
			id := player.Planet.Spec.Id + 1
			if universe.PlanetMap[id] == nil {
				id = 0
			}
			player.Planet = universe.PlanetMap[id].Planet
			player.Spawn()
		case m["PlanetL"].Key:
			id := player.Planet.Spec.Id - 1
			if universe.PlanetMap[id] == nil {
				if id < 0 {
					for i := range universe.PlanetMap {
						if i > id {
							id = i
						}
					}
				}
			}
			player.Planet = universe.PlanetMap[id].Planet
			player.Spawn()
		case m["Destroy"].Key:
			increment := player.LookDir().Mul(0.05)
			pos := player.Location()
			for i := 0; i < 100; i++ {
				pos = pos.Add(increment)
				cell := planet.CartesianToCell(pos)
				hitPlayer := false
				for _, otherPlayer := range universe.ConnectedPeople {
					if pos.Sub(otherPlayer.Position).Len() < 0.6 {
						log.Println(fmt.Sprintf("Hit %v", otherPlayer.Name))
						ctx, cancel := context.WithTimeout(context.Background(), time.Second)
						defer cancel()
						universe.GRPCClient.HitPlayer(ctx, &pb.HitPlayerRequest{
							From:   player.Name,
							Target: otherPlayer.Name,
							Amount: 1,
						})
						hitPlayer = true
						break
					}
				}
				if hitPlayer {
					break
				}
				if cell != nil && cell.Material != pb.Material_AIR {
					bestSpot := -1
					for i, slot := range player.Hotbar {
						if slot.Material == cell.Material {
							bestSpot = i
							break
						}
						if slot.Material == pb.Material_AIR && bestSpot < 0 {
							bestSpot = i
						}
					}
					if bestSpot >= 0 {
						player.Hotbar[bestSpot] = common.Slot{
							Material: cell.Material,
							Amount:   player.Hotbar[bestSpot].Amount + 1,
						}
					}
					cellIndex := planet.CartesianToCellIndex(pos)
					planetRen.SetCellMaterial(cellIndex, pb.Material_AIR, true)
					break
				}
			}
		case m["Build"].Key:
			increment := player.LookDir().Mul(0.05)
			pos := player.Location()
			prevCellIndex := pb.CellIndex{Lon: -1, Lat: -1, Alt: -1}
			cellIndex := pb.CellIndex{}
			for i := 0; i < 100; i++ {
				pos = pos.Add(increment)
				nextCellIndex := planet.CartesianToCellIndex(pos)
				if nextCellIndex.Lon != cellIndex.Lon || nextCellIndex.Lat != cellIndex.Lat || nextCellIndex.Alt != cellIndex.Alt {
					prevCellIndex = cellIndex
					cellIndex = nextCellIndex
					cell := planet.CellIndexToCell(cellIndex)
					if cell != nil && cell.Material != pb.Material_AIR {
						if prevCellIndex.Lon != -1 {
							hotbarslot := player.Hotbar[player.ActiveHotBarSlot]
							player.Hotbar[player.ActiveHotBarSlot].Amount--
							if hotbarslot.Amount == 1 {
								player.Hotbar[player.ActiveHotBarSlot] = common.Slot{}
							}
							planetRen.SetCellMaterial(prevCellIndex, hotbarslot.Material, true)
						}
						break
					}
				}
			}
		}
	case glfw.Release:
		switch key {
		case m["Up"].Key:
			player.HoldingJump = false
			player.UpVel = 0
		case m["Down"].Key:
			player.DownVel = 0
		case m["Forward"].Key:
			player.ForwardVel = 0
		case m["Backward"].Key:
			player.BackVel = 0
		case m["Right"].Key:
			player.RightVel = 0
		case m["Left"].Key:
			player.LeftVel = 0
		}
	}
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	player := universe.Player
	guikeycallback := screen.KeyCallBack()
	m := player.Mode
	if m == "Play" {
		keyCallbackPlay(w, key, scancode, action, mods)
	} else {
		if action == glfw.Press && key == glfw.KeyEscape {
			player.Mode = "Play"
			w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		} else if m == "Inventory" {
			keyCallbackInventory(w, key, scancode, action, mods)
		} else if m == "Text" || m == "Options" {
			guikeycallback(w, key, scancode, action, mods)
		}
	}
}

func windowSizeCallback(w *glfw.Window, wd, ht int) {
	fwidth, fheight := scene.FramebufferSize(w)
	gl.Viewport(0, 0, int32(fwidth), int32(fheight))
}

func mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	player := universe.Player
	guimousebuttoncallback := screen.MouseButtonCallback()
	m := player.Mode
	if m == "Play" {
		keyCallbackPlay(w, glfw.Key(button), 0, action, mods)
	} else {
		if action == glfw.Press && player.Mode != "Options" {
			player.Mode = "Play"
			w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		} else if m == "Inventory" || m == "Text" || m == "Options" {
			guimousebuttoncallback(w, button, action, mods)
		}
	}
}

// 	player := universe.Player
// 	planet := player.Planet

// 	guimousebuttoncallback := screen.MouseButtonCallback()
// 	planetRen := universe.PlanetMap[planet.ID]
// 	// keyCallbackPlay(w, glfw.Key(button), 0, action, mods)
// 	guimousebuttoncallback(w, button, action, mods)
// 	if cursorGrabbed(w) {
// 		if action == glfw.Press && button == glfw.MouseButtonLeft {
// 			increment := player.LookDir().Mul(0.05)
// 			pos := player.Location()
// 			for i := 0; i < 100; i++ {
// 				pos = pos.Add(increment)
// 				cell := planet.CartesianToCell(pos)
// 				hitPlayer := false
// 				for _, otherPlayer := range universe.ConnectedPeople {
// 					if pos.Sub(otherPlayer.Position).Len() < 0.6 {
// 						log.Println(fmt.Sprintf("Hit %v", otherPlayer.Name))
// 						var ret bool
// 						universe.RPC.Go("API.HitPlayer", common.HitPlayerArgs{From: player.Name, Target: otherPlayer.Name, Amount: 1}, &ret, nil)
// 						hitPlayer = true
// 						break
// 					}
// 				}
// 				if hitPlayer {
// 					break
// 				}
// 				if cell != nil && cell.Material != common.Air {
// 					cellIndex := planet.CartesianToCellIndex(pos)
// 					planetRen.SetCellMaterial(cellIndex, common.Air, true)
// 					break
// 				}
// 			}
// 		} else if action == glfw.Press && button == glfw.MouseButtonRight {
// 			increment := player.LookDir().Mul(0.05)
// 			pos := player.Location()
// 			prevCellIndex := common.CellIndex{Lon: -1, Lat: -1, Alt: -1}
// 			cellIndex := common.CellIndex{}
// 			for i := 0; i < 100; i++ {
// 				pos = pos.Add(increment)
// 				nextCellIndex := planet.CartesianToCellIndex(pos)
// 				if nextCellIndex != cellIndex {
// 					prevCellIndex = cellIndex
// 					cellIndex = nextCellIndex
// 					cell := planet.CellIndexToCell(cellIndex)
// 					if cell != nil && cell.Material != common.Air {
// 						if prevCellIndex.Lon != -1 {
// 							planetRen.SetCellMaterial(prevCellIndex, player.Hotbar[player.ActiveHotBarSlot], true)
// 						}
// 						break
// 					}
// 				}
// 			}
// 		}
// 	} else {
// 		if action == glfw.Press && player.Mode != "Options" {
// 			player.Mode = "Play"
// 			w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
// 		}
// 	}
// }
