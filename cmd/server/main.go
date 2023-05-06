package main

import (
	"github.com/amidvn/go-metrics/internal/apiserver"
)

func main() {
	s := apiserver.New()
	if err := s.Start(); err != nil {
		panic(err)
	}
}
