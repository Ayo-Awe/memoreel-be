package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/ayo-awe/memoreel-be/datastore"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestCreateVideo(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	videoRepo := NewVideoRepo(db)
	video := generateVideo()

	err := videoRepo.CreateVideo(context.Background(), video)
	require.NoError(t, err)

	newVideo, err := videoRepo.GetVideoByID(context.Background(), video.UID)
	require.NoError(t, err)

	require.Equal(t, newVideo, video)
}

func TestGetVideoByID(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	videoRepo := NewVideoRepo(db)
	video := generateVideo()

	_, err := videoRepo.GetVideoByID(context.Background(), video.UID)
	require.ErrorIs(t, err, datastore.ErrVideoNotFound)

	require.NoError(t, videoRepo.CreateVideo(context.Background(), video))

	foundVideo, err := videoRepo.GetVideoByID(context.Background(), video.UID)
	require.NoError(t, err)

	require.Equal(t, foundVideo, video)
}

func TestUpdateVideo(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	videoRepo := NewVideoRepo(db)
	video := generateVideo()

	require.NoError(t, videoRepo.CreateVideo(context.Background(), video))

	updatedVideo := &datastore.Video{
		UID:        video.UID,
		Key:        ulid.Make().String(),
		FileFormat: "mkv",
		SizeMB:     45,
	}

	require.NoError(t, videoRepo.UpdateVideo(context.Background(), updatedVideo))

	dbVideo, err := videoRepo.GetVideoByID(context.Background(), video.UID)
	require.NoError(t, err)

	dbVideo.CreatedAt = time.Time{}
	dbVideo.UpdatedAt = time.Time{}

	require.Equal(t, updatedVideo, dbVideo)
}

func TestDeleteVideo(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	videoRepo := NewVideoRepo(db)
	video := generateVideo()

	require.NoError(t, videoRepo.CreateVideo(context.Background(), video))

	require.NoError(t, videoRepo.DeleteVideo(context.Background(), video.UID))

	_, err := videoRepo.GetVideoByID(context.Background(), video.UID)
	require.ErrorIs(t, err, datastore.ErrVideoNotFound)
}

func generateVideo() *datastore.Video {
	return &datastore.Video{
		UID:        ulid.Make().String(),
		Key:        ulid.Make().String(),
		FileFormat: "mp4",
		SizeMB:     20,
	}
}
