package database

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

type Comic struct {
	ID       int
	URL      string
	Keywords string
}

type DB struct {
	db *sql.DB
}

type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

func NewDB(dsn string) *DB {
	db := ConnectToDatabase(dsn)
	err := MigrateDatabase(dsn, "up")
	if err != nil {
		log.Error().Err(err).Msg("Error applying migrations")
	}

	return &DB{db: db}
}

func (db *DB) GetComicByID(id int) (Comic, error) {
	var comic Comic
	err := db.db.QueryRow("SELECT id, url, keywords FROM comics WHERE id = ?", id).Scan(&comic.ID, &comic.URL, &comic.Keywords)
	if err != nil {
		return Comic{}, err
	}
	return comic, nil
}

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

func ConnectToDatabase(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Error().Err(err).Msg("error connecting to database")
		return nil
	}

	err = db.Ping()
	if err != nil {
		log.Error().Err(err).Msg("error pinging database")
		return nil
	}

	return db
}

func (db *DB) GetCount() (int, error) {
	var count int
	err := db.db.QueryRow("SELECT COUNT(*) FROM comics").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func MigrateDatabase(dsn string, direction string) error {
	// создание абсолютного пути к папке с миграционными файлами
	migrationsPath := filepath.Join("internal", "core", "database", "migrations")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"mysql",
		driver,
	)
	if err != nil {
		return err
	}

	if direction == "up" {
		err = m.Up()
	} else if direction == "down" {
		err = m.Down()
	}

	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	if direction == "up" {
		log.Info().Msg("Migrations applied successfully")
	} else if direction == "down" {
		log.Info().Msg("Migrations reverted successfully")
	}

	return nil
}

func (db *DB) CreateUser(username, password, role string) error {
	_, err := db.db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", username, password, role)
	return err
}

func (db *DB) GetUserByUsername(username string) (User, error) {
	var user User
	err := db.db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func PrettyPrint(v []Comic) bytes.Buffer {
	var responseBuffer bytes.Buffer
	for i, comic := range v {
		if i >= 10 {
			break
		}
		responseBuffer.WriteString(fmt.Sprintf("\n\nComic %d: %s", i+1, comic.URL))
	}
	return responseBuffer
}
