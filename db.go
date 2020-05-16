package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// A DB read / writes scheduler's data from / to disk.
type DB struct {
	Todos    map[string]Todo
	Triggers map[string]Trigger
	filename string `json:"-"`
}

// NewDB returns a DB located in filename.
func NewDB(filename string) (*DB, error) {
	db := DB{
		filename: filename,
		Todos:    make(map[string]Todo),
		Triggers: make(map[string]Trigger),
	}
	err := db.Read()
	if err != nil {
		return nil, err
	}
	return &db, nil
}

// Save stores the database onto disk. DB file will be created if doesn't exist.
func (db *DB) Write() error {
	encoded, err := json.MarshalIndent(*db, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(db.filename, encoded, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Reads loads the database from disk. If file doesn't exit then DB will be empty.
func (db *DB) Read() error {
	encoded, err := ioutil.ReadFile(db.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	err = json.Unmarshal(encoded, db)
	if err != nil {
		return err
	}
	return nil
}
