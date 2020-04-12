package main

import (
	"log"
	"os"

	"github.com/dqn/gones/nes"
)

func run() error {
	n, err := nes.New(os.Args[1])
	if err != nil {
		return err
	}
	return n.Run()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
