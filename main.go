package main

import (
	"github.com/elastic/beats/libbeat/beat"
	"github.com/urso/govarbeat/beater"
)

func main() {
	beat.Run("govarbeat", "", beater.New())
}
