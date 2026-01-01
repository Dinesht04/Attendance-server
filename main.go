package main

import (
	"github.com/dinesht04/ws-attendance/data"
	"github.com/dinesht04/ws-attendance/server"
)

func main() {
	db := data.ConnectToMongo()
	server.StartServer(db)
}
