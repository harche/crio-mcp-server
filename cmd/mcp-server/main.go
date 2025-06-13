package main

import (
	"log"

	"github.com/harche/crio-mcp-server/pkg/sdkserver"
)

func main() {
	srv := sdkserver.New()
	if err := sdkserver.StartStdio(srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
