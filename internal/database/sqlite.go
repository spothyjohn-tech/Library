package database

import (
	"database/sql"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	var count int
	query := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('Users', 'Books', 'Issue');`
	err2 := db.QueryRow(query).Scan(&count)
	if err2 != nil {
		return nil, err
	}
	allExist := (count == 3)

	if !allExist {

		file, err := os.ReadFile("../migrations/models.sql")
		if err != nil {
			return nil, err
		}

		requests := strings.Split(string(file), ";")

		for _, request := range requests {
			request = strings.TrimSpace(request)
			if request == "" {
				continue
			}
			_, err = db.Exec(request)
			if err != nil {
				return nil, err
			}
		}

		id, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
        hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
		query := `INSERT INTO Users(ID, "Name", Email, Password, Role, RegistrationDate) VALUES(?, ?, ?, ?, ?, ?)`
        db.Exec(query, id, "System Admin", "admin@admin.com", string(hash), "admin", time.Now())
        slog.Info("Создан первый админ: email: admin@admin.com / passw: admin123")

	}

	return db, nil
}
