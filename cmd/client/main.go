package main

import (
	"github.com/jeffbaumes/govox/pkg/client"
)

const (
	address = "localhost:50051"
)

func main() {
	client.Start("andrew", "default", 50051, nil)
}
