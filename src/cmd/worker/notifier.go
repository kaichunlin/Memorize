package main

import (
	"db"
	"log"
	"notify"
)

func main() {
	err := db.Init()
	if err != nil {
		log.Println("db.InitDb error:", err)
	}
	_, err = notify.Notify()
	if err != nil {
		log.Println(err)
	}
}
