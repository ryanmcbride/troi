package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	gorm "gorm.io/gorm"
)

func initDB() *gorm.DB {
	dbSpec := os.Getenv("DATABASE_URL")

	if dbSpec == "" {
		dbSpec = "host=127.0.0.1 port=5432 user=ryanmcbride dbname=postgres sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dbSpec), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Migrate the schema
	db.AutoMigrate(&User{})

	return db
}
