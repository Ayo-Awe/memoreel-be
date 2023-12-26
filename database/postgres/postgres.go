package postgres

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	dbx *sqlx.DB
}

// TODO: replace dsn param with config param
func NewDB(dsn string) (*PostgresDB, error) {
	db, err := sqlx.Connect("postgres", dsn)
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
