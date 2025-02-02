package data

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sqlx.DB
	authRepo
}

func CreatePostgres(connectionString string) (*Postgres, error) {
	db, err := sqlx.Connect("postgres", "user=postgres password=server dbname=tranquility sslmode=disable")
	if err != nil {
		return nil, err
	}

	return &Postgres{
		db:       db,
		authRepo: authRepo{db},
	}, nil
}
