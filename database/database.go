package database

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	database, err := gorm.Open(sqlite.Open("database/database.db"), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	DB = database
}

func Close() {
	db, err := DB.DB()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	db.Close()
}