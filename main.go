package main

import (
	"fmt"
	"log"
)

func main() {
	loader := &JSONConfigLoader{}
	cfg, err := loader.Load("config/config.json")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Config loaded: db=%s@%s:%d/%s\n", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
}
