package server

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jeffbaumes/govox/pkg/common"
	pb "github.com/jeffbaumes/govox/pkg/govox"
	_ "github.com/mattn/go-sqlite3" // Needed to use sqlite
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	universe *common.Universe
)

// Start takes a name, seed, and port and starts the universe server
func Start(name string, seed, port int, system string) {
	_ = os.Mkdir("worlds/", os.ModePerm)
	dbName := "worlds/" + name + ".db"

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS chunk (planet INT, lon INT, lat INT, alt INT, data BLOB, PRIMARY KEY (planet, lat, lon, alt))")
	if err != nil {
		log.Fatalf("failed to create chunk table statement: %v", err)
	}
	if _, err := stmt.Exec(); err != nil {
		log.Fatalf("failed to create chunk table: %v", err)
	}
	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS planet (id INT PRIMARY KEY, data BLOB)")
	if err != nil {
		log.Fatalf("failed to create planet table statement: %v", err)
	}
	if _, err := stmt.Exec(); err != nil {
		log.Fatalf("failed to create planet table: %v", err)
	}
	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS entity (name TEXT PRIMARY KEY, data BLOB)")
	if err != nil {
		log.Fatalf("failed to create entity table statement: %v", err)
	}
	if _, err := stmt.Exec(); err != nil {
		log.Fatalf("failed to create entity table: %v", err)
	}

	universe = common.NewUniverse(db, system)

	// // Connect to generator
	// conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	// if err != nil {
	// 	log.Fatalf("did not connect: %v", err)
	// }
	// defer conn.Close()
	// generatorClient = pb.NewGeneratorClient(conn)
	// var cancel context.CancelFunc
	// generatorContext, cancel = context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// response, err := generatorClient.CellMaterial(generatorContext, &pb.CellMaterialRequest{})
	// log.Printf("response: %v", response)

	// Start the server
	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGovoxServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// // Start takes a name, seed, and port and starts the universe server
// func Start(name string, seed, port int) {
// 	if port == 0 {
// 		port = 5555
// 	}
// 	_ = os.Mkdir("worlds/", os.ModePerm)
// 	dbName := "worlds/" + name + ".db"

// 	db, err := sql.Open("sqlite3", dbName)
// 	checkErr(err)
// 	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS chunk (planet INT, lon INT, lat INT, alt INT, data BLOB, PRIMARY KEY (planet, lat, lon, alt))")
// 	checkErr(err)
// 	_, err = stmt.Exec()
// 	checkErr(err)
// 	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS planet (id INT PRIMARY KEY, data BLOB)")
// 	checkErr(err)
// 	_, err = stmt.Exec()
// 	checkErr(err)
// 	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS player (name TEXT PRIMARY KEY, data BLOB)")
// 	checkErr(err)
// 	_, err = stmt.Exec()
// 	checkErr(err)

// 	universe = common.NewUniverse(db, getsystem())

// 	api := new(API)
// 	listener, e := net.Listen("tcp", fmt.Sprintf(":%v", port))
// 	if e != nil {
// 		log.Fatal("listen error:", e)
// 	}
// 	log.Printf("Server listening on port %v...\n", port)
// 	for {
// 		conn, e := listener.Accept()
// 		if e != nil {
// 			panic(e)
// 		}

// 		// Set up server side of yamux
// 		mux, e := yamux.Server(conn, nil)
// 		if e != nil {
// 			panic(e)
// 		}
// 		muxConn, e := mux.Accept()
// 		if e != nil {
// 			panic(e)
// 		}
// 		srpc := rpc.NewServer()
// 		srpc.Register(api)
// 		go srpc.ServeConn(muxConn)

// 		// Set up stream back to client
// 		stream, e := mux.Open()
// 		if e != nil {
// 			panic(e)
// 		}
// 		crpc := rpc.NewClient(stream)

// 		// Ask client for player name
// 		var state common.PlayerState
// 		e = crpc.Call("API.GetPersonState", 0, &state)
// 		if e != nil {
// 			log.Fatal("GetPersonState error:", e)
// 		}
// 		p := connectedPerson{state: state, rpc: crpc}
// 		log.Println(p.state.Name)
// 		api.connectedPeople = append(api.connectedPeople, &p)
// 	}
// }

// type connectedPerson struct {
// 	rpc   *rpc.Client
// 	state common.PlayerState
// }
