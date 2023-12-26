package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ayo-awe/memoreel-be/database"
	"github.com/ayo-awe/memoreel-be/datastore"
	"github.com/jmoiron/sqlx"
)

const (
	createVideo = `
	INSERT INTO videos (id, key, file_format,size_mb)
	VALUES ($1,$2,$3,$4)
	RETURNING *;
	`

	fetchVideoById = `
	SELECT
		id,
		key,
		file_format,
		size_mb,
		created_at,
		updated_at,
		deleted_at
	FROM videos
	WHERE id = $1 AND deleted_at IS NULL;
	`

	updatedVideo = `
	UPDATE videos SET
		key = $2,
		file_format = $3,
		size_mb = $4,
		updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL;
	`

	deleteVideo = `
	UPDATE videos SET
		deleted_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL;
	`
)

var (
	ErrVideoNotUpdated = errors.New("video could not be updated")
	ErrVideoNotDeleted = errors.New("video could not be deleted")
)

type videoRepo struct {
	db *sqlx.DB
}

func NewVideoRepo(db database.Database) datastore.VideoRepository {
	return &videoRepo{db: db.GetDB()}
}

func (v videoRepo) GetVideoByID(ctx context.Context, id string) (*datastore.Video, error) {
	video := &datastore.Video{}
	err := v.db.QueryRowxContext(ctx, fetchVideoById, id).StructScan(video)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datastore.ErrVideoNotFound
		}

		return nil, err
	}

	return video, nil
}

func (v videoRepo) CreateVideo(ctx context.Context, video *datastore.Video) error {
	row := v.db.QueryRowxContext(ctx, createVideo,
		video.UID,
		video.Key,
		video.FileFormat,
		video.SizeMB,
	)

	err := row.StructScan(video)
	if err != nil {
		return err
	}

	return nil
}

func (v videoRepo) UpdateVideo(ctx context.Context, video *datastore.Video) error {
	res, err := v.db.ExecContext(ctx, updatedVideo,
		video.UID,
		video.Key,
		video.FileFormat,
		video.SizeMB,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		return ErrVideoNotUpdated
	}

	return nil
}

func (v videoRepo) DeleteVideo(ctx context.Context, videoID string) error {
	res, err := v.db.ExecContext(ctx, deleteVideo, videoID)
	if err != nil {
		return nil
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		return ErrVideoNotDeleted
	}

	return nil
}
