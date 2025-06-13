package main

import (
	"flag"
	"log"

	"github.com/harche/crio-mcp-server/pkg/sdkserver"
)

func main() {
	config := flag.String("config", "/etc/crio/crio.conf", "path to CRI-O config")
	flag.Parse()

	srv := sdkserver.New(*config)
	if err := sdkserver.StartStdio(srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
