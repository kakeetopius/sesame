package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prashantgupta24/firewalld-rest/server"
)

func main() {
	fmt.Println("starting application")

	// TODO:
	// load config
	// init firewall

	server := server.New(nil)
	log.Fatal(http.ListenAndServe(":8080", server))
}
