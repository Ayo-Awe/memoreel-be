package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/ayo-awe/memoreel-be/config"
	"github.com/ayo-awe/memoreel-be/database"
	"github.com/stretchr/testify/require"
)

var (
	_db *PostgresDB
)

func getDB(t *testing.T) (database.Database, func()) {

	require.NoError(t, config.LoadConfig())
	cfg := config.Get(config.Test)

	db, err := NewDB(cfg)
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
