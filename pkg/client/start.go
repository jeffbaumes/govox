package client

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jeffbaumes/govox/pkg/common"
	pb "github.com/jeffbaumes/govox/pkg/govox"
	"github.com/jeffbaumes/govox/pkg/gui"
	"github.com/jeffbaumes/govox/pkg/scene"
	"google.golang.org/grpc"
)

const (
	targetFPS = 60
	gravity   = 9.8
)

var (
	universe *scene.Universe
	screen   *gui.Screen
	op       *scene.Options
)

// Start starts a client with the given username, host, and port
func Start(username, host string, port int, scr *gui.Screen) {
	screen = scr
	screen.Clear()
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 5555
	}
	window := screen.Window
	if screen.Window == nil {
		runtime.LockOSThread()

		window = initGlfw()
		initOpenGL()
	}
	defer glfw.Terminate()

	address := fmt.Sprintf("%v:%v", host, port)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := pb.NewGovoxClient(conn)

	player := common.NewPlayer(username)
	universe = scene.NewUniverse(grpcClient, player)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	planetsResult, err := grpcClient.GetPlanets(ctx, &pb.GetPlanetsRequest{})
	if err != nil {
		log.Fatalf("could not get planets: %v", err)
	}
	log.Printf("Planets: %v", planetsResult.Planets)

	for _, spec := range planetsResult.Planets {
		planet := common.NewPlanet(grpcClient, nil, *spec)
		planetRen := scene.NewPlanet(planet)
		universe.AddPlanet(planetRen)
	}

	op = scene.NewOptions(screen)
	player.Planet = universe.PlanetMap[0].Planet
	player.Spawn()

	over := scene.NewCrosshair()
	text := &scene.Text{}
	bar := scene.NewHotbar()
	health := scene.NewHealth()
	player.Mode = "Play"

	// // Setup server connection
	// smuxConn, e := cmux.Accept()
	// if e != nil {
	// 	panic(e)
	// }
	// s := rpc.NewServer()
	// clientAPI := new(API)
	// s.Register(clientAPI)
	// go s.ServeConn(smuxConn)

	peopleRen := scene.NewPlayers(&universe.ConnectedPeople)
	focusRen := scene.NewFocusCell()

	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetKeyCallback(keyCallback)
	window.SetCursorPosCallback(cursorPosCallback())
	window.SetSizeCallback(windowSizeCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)

	startTime := time.Now()
	t := startTime
	syncT := t
	for !window.ShouldClose() {
		h := float32(time.Since(t)) / float32(time.Second)
		t = time.Now()
		elapsedSeconds := float64(time.Since(startTime)) / float64(time.Second)

		drawFrame(h, player, text, over, peopleRen, focusRen, bar, health, screen, elapsedSeconds, op)

		player.UpdatePosition(h)

		if float64(time.Since(syncT))/float64(time.Second) > 0.05 {
			syncT = time.Now()
			request := pb.UpdatePlayerStateRequest{
				Name: player.Name,
				Position: []float64{
					float64(player.Location().X()),
					float64(player.Location().Y()),
					float64(player.Location().Z()),
				},
				LookDir: []float64{
					float64(player.LookDir().X()),
					float64(player.LookDir().Y()),
					float64(player.LookDir().Z()),
				},
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := grpcClient.UpdatePlayerState(ctx, &request)
			if err != nil {
				log.Fatalf("update person state failed: %v", err)
			}
		}
		time.Sleep(time.Second/time.Duration(targetFPS) - time.Since(t))
	}
}
