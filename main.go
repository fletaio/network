package main

import (
	"time"

	"fleta"
	"fleta/mock/network"
)

func main() {
	var fleta fleta.Fleta
	network.Fleta(&fleta)
	network.Run()

	for {
		time.Sleep(time.Hour)
	}

}
