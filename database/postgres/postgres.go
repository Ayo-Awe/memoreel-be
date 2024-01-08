package postgres

import (
	"github.com/ayo-awe/memoreel-be/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	dbx *sqlx.DB
}

func NewDB(config config.Configuration) (*PostgresDB, error) {
	db, err := sqlx.Connect("postgres", config.Database.BuildDSN())
	if err != nil {
		return nil, err
	}

	pgDB := &PostgresDB{dbx: db}

	return pgDB, nil
}

func (p *PostgresDB) GetDB() *sqlx.DB {
	return p.dbx
}

func (p *PostgresDB) Close() error {
	return p.dbx.Close()
}
