package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/ayo-awe/memoreel-be/database"
	"github.com/stretchr/testify/require"
)

var (
	_db *PostgresDB
)

func getDB(t *testing.T) (database.Database, func()) {
	// TODO: replace hardcoded dsn with config
	db, err := NewDB("postgresql://postgres:postgres@localhost:5432/memoreel_test?sslmode=disable")
	_db = db

	require.NoError(t, err)

	return _db, func() {
		err = _db.truncateTables()
		require.NoError(t, err)
	}
}

func (p *PostgresDB) truncateTables() error {
	tables := `
		reels,
		videos,
		users
	`

	_, err := p.dbx.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE %s CASCADE;", tables))
	return err
}
