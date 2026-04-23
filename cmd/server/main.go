package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"kikneip.com/akirawhats/internal/api"
	"kikneip.com/akirawhats/internal/db"
)

func main() {
	if err := godotenv.Load(); err != nil {
		return
	}
	db.Open()
	defer db.Close()

	engine := gin.Default()

	api.RegistrarRotas(engine)

	err := engine.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		log.Fatalf("error running http server: %v", err)
		return
	}
}
