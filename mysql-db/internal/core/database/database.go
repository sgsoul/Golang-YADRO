package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Comic represents the structure of a comic
type Comic struct {
	ID       int
	URL      string
	Keywords string
}

// DB represents the database interface
type DB struct {
	db *sql.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := ConnectToDatabase(dsn)
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}
// GetComicByID retrieves a comic by its ID from the database
func (db *DB) GetComicByID(id int) (Comic, error) {
	var comic Comic
	err := db.db.QueryRow("SELECT id, url, keywords FROM comics WHERE id = ?", id).Scan(&comic.ID, &comic.URL, &comic.Keywords)
	if err != nil {
		return Comic{}, err
	}
	return comic, nil
}

// GetIndex retrieves the index from the database
func (db *DB) GetIndex() (map[string][]int, error) {
	var indexJSON string
	err := db.db.QueryRow("SELECT index_json FROM index_table WHERE id = 1").Scan(&indexJSON)
	if err != nil {
		return nil, err
	}
	var index map[string][]int
	err = json.Unmarshal([]byte(indexJSON), &index)
	if err != nil {
		return nil, err
	}
	return index, nil
}

// GetAllComics retrieves all comics from the database
func (db *DB) GetAllComics() ([]Comic, error) {
	rows, err := db.db.Query("SELECT id, url, keywords FROM comics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comics []Comic
	for rows.Next() {
		var comic Comic
		err := rows.Scan(&comic.ID, &comic.URL, &comic.Keywords)
		if err != nil {
			return nil, err
		}
		comics = append(comics, comic)
	}
	return comics, nil
}

func (db *DB) SaveComicToDatabase(comic Comic) error {
	_, err := db.db.Exec("INSERT INTO comics (url, keywords) VALUES (?, ?)", comic.URL, comic.Keywords)
	if err != nil {
		return err
	}
	return nil
}

// CreateDatabase creates the database based on the DSN provided
func createDatabase(dsn string) error {
	// Extract database name from DSN
	parts := strings.Split(dsn, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid DSN format")
	}
	dbName := strings.Split(parts[len(parts)-1], "?")[0]

	// Connect to MySQL server
	db, err := sql.Open("mysql", dsn[:strings.LastIndex(dsn, "/")])
	if err != nil {
		return err
	}
	defer db.Close()

	fmt.Print("\n\n damn it still not workin\n\n")

	// Create the database if it doesn't exist
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + ";")
	if err != nil {
		return err
	}

	fmt.Print("\n\n yaaay \n\n")
	return nil
}

// CreateTable creates the comics table
func CreateTable(dsn string) error {
	// Connect to MySQL server
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create the comics table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS comics (
		id INT AUTO_INCREMENT PRIMARY KEY,
		url VARCHAR(255) NOT NULL,
		keywords TEXT
	)`)
	if err != nil {
		return err
	}

	return nil
}

func ConnectToDatabase(dsn string) (*sql.DB, error) {
	createDatabase(dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	// Create tables if they don't exist
	if err := CreateTable(dsn); err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return db, nil
}

func (db *DB) GetCount() (int, error) {
	var count int
	err := db.db.QueryRow("SELECT COUNT(*) FROM comics").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}