package main

import "log"

func main() {
	loader := &JSONConfigLoader{}
	cfg, err := loader.Load("config/config.json")
	if err != nil {
		log.Fatal(err)
	}

	logger := NewFileWarningLogger(cfg.WarningLogDir)

	mysqlDB, err := NewMySQLDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer mysqlDB.Close()

	tokenManager, err := NewMySQLTokenManager(cfg, mysqlDB.db, logger)
	if err != nil {
		log.Fatal(err)
	}

	authService := NewAuthService(mysqlDB, tokenManager, cfg)

	server := NewHTTPServer(authService, cfg)
	log.Printf("Server starting on port %d", cfg.Port)
	log.Fatal(server.Run())
}
