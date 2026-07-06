package main

import (
	"log"
	"net/http"
	"url-checker/internal/api"
)

func main() {
	s := api.NewServer()

	log.Fatal(http.ListenAndServe(":8080", s.Routes()))
}
