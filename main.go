package main

import (
	"flag"
	"log"
)

func main() {
	addr := flag.String("addr", ":2222", "address to listen on")
	keyPath := flag.String("hostkey", "host_key", "path to SSH host key (created if missing)")
	flag.Parse()

	hub := NewHub()

	if err := Serve(*addr, *keyPath, hub); err != nil {
		log.Fatal(err)
	}
}
