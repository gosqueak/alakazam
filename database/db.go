package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const DbFileName = "data.sqlite"

func Load(fp string) *sql.DB {
	db, err := sql.Open("sqlite3", fp)

	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS socketEvent (
			toUserId TEXT NOT NULL,
			json TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS indexSocketEventToUserId on socketEvent(toUserId);
	`)

	if err != nil {
		panic(err)
	}

	return db
}

func GetStoredSocketEvents(db *sql.DB, toUserId string) (eventJSON []string, err error) {
	stmt := "SELECT json FROM socketEvent WHERE toUserId=?"
	rows, err := db.Query(stmt, toUserId)

	if err != nil {
		return nil, err
	}

	for {
		if ok := rows.Next(); !ok {
			break
		}

		var json string
		err = rows.Scan(&json)

		if err != nil {
			break
		}

		eventJSON = append(eventJSON, json)
	}

	return eventJSON, nil
}

func StoreSocketEvent(db *sql.DB, toUserId string, eventJSON string) error {
	stmt := "INSERT INTO socketEvent (toUserId, json) VALUES (?, ?)"
	_, err := db.Exec(stmt, toUserId, eventJSON)

	return err
}
