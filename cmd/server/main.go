package main

import (
	"github.com/jeffbaumes/govox/pkg/server"
)

func main() {
	server.Start("default", 0, 50051, "default")
}
