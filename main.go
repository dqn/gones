package main

import (
	"log"

	"github.com/dqn/gones/nes"
)

func run() error {
	n, err := nes.New("./sample1/sample1.nes")
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
