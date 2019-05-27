package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/georgemac/repositories/pkg/cached"
	"github.com/georgemac/repositories/pkg/repositories"
	"github.com/georgemac/repositories/pkg/server"
)

var (
	addr              = flag.String("addr", ":8080", "address on which to serve the repositories service")
	repositoryService = flag.String("repository-addr", "http://localhost:7080", "address on which repository service is found")
)

func main() {
	flag.Parse()

	service, err := repositories.New(*repositoryService)
	if err != nil {
		log.Fatal(err)
	}

	server := server.New(cached.New(service))

	http.Handle("/repositories", server)

	fmt.Printf("Listening on %q\n", *addr)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
