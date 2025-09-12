package main

import (
	"log"

	"github.com/palo-verde-digital/codex/pkg/filegraph"
)

func main() {
	src := "vault"

	g, err := filegraph.Create(src)
	if err != nil {
		log.Panicf("Unable to create graph from vault: %s", err.Error())
	}

	log.Print(g.String())
}
