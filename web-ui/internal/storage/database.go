package storage

import (
	"bytes"
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/core"
)

type MySQLStorage struct {
	db *sql.DB
}

func NewMySQLDB(dsn string) *MySQLStorage {
	db := connectToDatabase(dsn)
	err := migrateDatabase(dsn, "up")
	if err != nil {
		log.Error().Err(err).Msg("Error applying migrations")
	}

	return &MySQLStorage{db: db}
}

func (mysql *MySQLStorage) GetComicByID(id int) (core.Comic, error) {
	var comic core.Comic
	err := mysql.db.QueryRow("SELECT id, url, keywords FROM comics WHERE id = ?", id).Scan(&comic.ID, &comic.URL, &comic.Keywords)
	if err != nil {
		return core.Comic{}, err
	}
	return comic, nil
}

func (mysql *MySQLStorage) GetAllComics() ([]core.Comic, error) {
	rows, err := mysql.db.Query("SELECT id, url, keywords FROM comics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comics []core.Comic
	for rows.Next() {
		var comic core.Comic
		err := rows.Scan(&comic.ID, &comic.URL, &comic.Keywords)
		if err != nil {
			return nil, err
		}
		comics = append(comics, comic)
	}
	return comics, nil
}

func (mysql *MySQLStorage) SaveComicToDatabase(comic core.Comic) error {
	_, err := mysql.db.Exec("INSERT INTO comics (url, keywords) VALUES (?, ?)", comic.URL, comic.Keywords)
	if err != nil {
		return err
	}
	return nil
}

func connectToDatabase(dsn string) *sql.DB {
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

func (mysql *MySQLStorage) GetCount() (int, error) {
	var count int
	err := mysql.db.QueryRow("SELECT COUNT(*) FROM comics").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func migrateDatabase(dsn string, direction string) error {
	// создание абсолютного пути к папке с миграционными файлами
	migrationsPath := filepath.Join("internal", "storage", "migrations")

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

func (mysql *MySQLStorage) CreateUser(username, password, role string) error {
	_, err := mysql.db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", username, password, role)
	return err
}

func (mysql *MySQLStorage) GetUserByUsername(username string) (core.User, error) {
	var user core.User
	err := mysql.db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return core.User{}, err
	}
	return user, nil
}

func (mysql *MySQLStorage) PrettyPrint(v []core.Comic) bytes.Buffer {
	var responseBuffer bytes.Buffer
	for i, comic := range v {
		if i >= 10 {
			break
		}
		responseBuffer.WriteString(fmt.Sprintf("\n\nComic %d: %s", i+1, comic.URL))
	}
	return responseBuffer
}
