package main

import (
	"log"
	"os"

	"kv-storage/internal/app"
	_ "kv-storage/docs"
)

// @title           KV Storage API
// @version         1.0
// @description     Modern key-value storage with HTTP API built on Tarantool
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth
func main() {
	application, err := app.Bootstrap()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
		os.Exit(1)
	}
	
	application.Run()
} 