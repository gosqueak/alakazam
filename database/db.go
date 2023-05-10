package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const DbFileName = "data.sqlite"

type Envelope struct {
	ToUserId string `json:"toUserId"`
	Event    string `json:"event"`
}

func Load(fp string) *sql.DB {
	db, err := sql.Open("sqlite3", fp)

	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS envelope (
			toUserId TEXT NOT NULL,
			event TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS indexEnvelopeToUserId ON envelope(toUserId);
	`)

	if err != nil {
		panic(err)
	}

	return db
}

func GetStoredEnvelopes(db *sql.DB, toUserId string) ([]Envelope, error) {
	stmt := "SELECT toUserId, event FROM envelope WHERE toUserId=?"
	rows, err := db.Query(stmt, toUserId)
	
	var envelopes []Envelope

	if err != nil {
		return envelopes, err
	}

	for {
		if ok := rows.Next(); !ok {
			break
		}

		env := Envelope{}
		err := rows.Scan(&env.ToUserId, &env.Event)

		if err != nil {
			break
		}

		envelopes = append(envelopes, env)
	}

	return envelopes, nil
}

func StoreEnvelope(db *sql.DB, toUserId string, event string) error {
	// TODO need to delete envelopes here
	stmt := "INSERT INTO envelope (toUserId, event) VALUES (?, ?)"
	_, err := db.Exec(stmt, toUserId, event)

	return err
}
