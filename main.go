package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/urso/govarbeat/beater"
)

func main() {
	err := beat.Run("govarbeat", "", beater.New())
	if err != nil {
		os.Exit(1)
	}
}
