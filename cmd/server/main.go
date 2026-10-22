package main

import (
	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/app/services/database"
)

func main() {
	startup()
}

func startup() {
	kernel.Init()

	database.GetDB()
}
