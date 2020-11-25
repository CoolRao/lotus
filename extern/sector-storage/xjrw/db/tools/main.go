package main

import (
	"github.com/filecoin-project/lotus/extern/sector-storage/xjrw/db"
	"log"
)

func main() {
	err := db.CreateTable()
	if err != nil {
		log.Fatal(err)
	}
}
