package main

import (
	"flag"
	"log"

	"github.com/harche/crio-mcp-server/pkg/server"
)

func main() {
	addr := flag.String("listen", ":50051", "address to listen on")
	config := flag.String("config", "/etc/crio/crio.conf", "path to CRI-O config")
	flag.Parse()

	srv := server.New(*config)
	if err := srv.Start(*addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
