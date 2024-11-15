package main

import (
	"RestAPI/app"
	"RestAPI/core"
	"RestAPI/db"
	"RestAPI/docs"
	"log"
	"os"
)

func main() {
	Production := os.Getenv("PRODUCTION")
	if Production == "true" {
		core.PRODUCTION = true
	}

	er := core.InitEnv()
	if er != nil {
		log.Println("Error initializing environment", er)
		return
	}

	er = db.ConnectToDB(core.DB_CREDENTIALS)
	if er != nil {
		log.Println("Error connecting to DB", er)
		return
	}

	app.InitHandlers()

	er = docs.GenerateDocs()
	if er != nil {
		log.Println("Error generating docs", er)
		return
	}

	serv, er := core.CreateServer(app.MainApplication)
	if er != nil {
		log.Println("Error creating server", er)
		return
	}

	er = serv.Start()
	if er != nil {
		log.Println("Error starting server", er)
		return
	}

	select {}
}
