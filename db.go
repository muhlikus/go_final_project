package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func dbConnect(dbFileName string) (*sql.DB, error) {

	dbMustBeCreated := !dbExists(dbFileName)

	db, err := sql.Open("sqlite", dbFileName)
	if err != nil {
		return nil, err
	}

	if dbMustBeCreated {
		err = dbCreate(db)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func dbExists(dbFileName string) bool {

	appPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	return fileExists(filepath.Join(appPath, dbFileName))

}

func dbCreate(db *sql.DB) error {

	_, err := db.Exec(`
		CREATE TABLE scheduler ( 
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			date CHAR(8) NOT NULL DEFAULT "", 
			title TEXT NOT NULL DEFAULT "",
			comment TEXT NOT NULL DEFAULT "",
			repeat CHAR(128) NOT NULL DEFAULT "")`)

	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE INDEX idx_date ON scheduler (date)`)

	if err != nil {
		return err
	}

	return nil
}

func fileExists(fileFullName string) bool {
	_, err := os.Stat(fileFullName)
	return err == nil
}
