package types

import (
	"log/slog"

	"github.com/ayo-awe/memoreel-be/database"
)

type APIOptions struct {
	DB     database.Database
	Logger slog.Logger
}
