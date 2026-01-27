package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"terminal_store/pkg/api"
	"terminal_store/pkg/db"
	"terminal_store/pkg/env"
)

func main() {
	_ = env.Load(".env")
	conn, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}
    defer conn.Close()

    if err := db.RunMigrations(conn, "migrations"); err != nil {
        log.Fatal(err)
    }

    r := gin.Default()
    api.Register(r, conn)

    addr := os.Getenv("SERVER_ADDR")
    if addr == "" {
        addr = ":8080"
    }
    log.Println("server listening on", addr)
    log.Fatal(r.Run(addr))
}
