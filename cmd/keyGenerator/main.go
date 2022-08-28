package main

import (
	"log"

	"github.com/zklevsha/go-musthave-devops/internal/rsaencrypt"
)

func main() {
	err := rsaencrypt.Generate(".")
	if err != nil {
		log.Fatalf(err.Error())
	}
}
