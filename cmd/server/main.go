package main

import (
	"log"

	"github.com/amidvn/go-metrics/internal/apiserver"
)

func main() {
	s := apiserver.New()
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
