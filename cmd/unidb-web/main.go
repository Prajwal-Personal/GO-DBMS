package main

import (
	"log"
	
	"github.com/unidb/unidb-go/web"
)

func main() {
	// Start the web dashboard on port 8080
	if err := web.StartServer(8080); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
