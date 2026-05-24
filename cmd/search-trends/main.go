package main

import (
	"os"

	"github.com/HAL-X9/search-trends-service/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
