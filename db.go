package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {

	var err error
	db, err = sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		// apparently this DOUBLE CHECKS that it's up?
		panic(err)
	}

	muteSchema := `
	CREATE TABLE IF NOT EXISTS mutes (
		id UUID PRIMARY KEY,
		discord_id TEXT NOT NULL,
		channel_id TEXT,
		expiration TIMESTAMP
	);
	`
	_, err = db.Exec(muteSchema)
	if err != nil {
		log.Printf("%q: %s\n", err, muteSchema)
		log.Println("Unable to create mutes table")
		panic(err)
	}
}

// I dont know why this panics, but it does and ill fix it later

// func schema() (err error) {

// 	_, err = db.Exec(`
// CREATE TABLE IF NOT EXISTS mutes (
// 	id UUID PRIMARY KEY,
// 	discord_id TEXT NOT NULL,
// 	channel_id TEXT,
// 	expiration TIMESTAMP
// );
// `)
// 	if err != nil {
// 		log.Println(err)
// 		log.Println("Unable to create mutes table")
// 		panic(err)
// 	}

// 	return
// }
