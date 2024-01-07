package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ayo-awe/memoreel-be/database"
	"github.com/ayo-awe/memoreel-be/datastore"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestCreateReel(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)

	user := seedUser(t, db)
	video := seedVideo(t, db)
	reel := generateReel(video.UID, user.UID)

	require.NoError(t, reelRepo.CreateReel(context.Background(), reel))

	newReel, err := reelRepo.GetReelByID(context.Background(), reel.UID)
	require.NoError(t, err)

	require.Equal(t, reel, newReel)
}

func TestGetReelById(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)

	user := seedUser(t, db)
	video := seedVideo(t, db)
	reel := generateReel(video.UID, user.UID)

	_, err := reelRepo.GetReelByID(context.Background(), reel.UID)
	require.ErrorIs(t, err, datastore.ErrReelNotFound)

	require.NoError(t, reelRepo.CreateReel(context.Background(), reel))

	dbReel, err := reelRepo.GetReelByID(context.Background(), reel.UID)
	require.NoError(t, err)

	require.Equal(t, reel, dbReel)
}

func TestGetReelByConfirmationToken(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)

	user := seedUser(t, db)
	video := seedVideo(t, db)
	reel := generateReel(video.UID, user.UID)

	_, err := reelRepo.GetReelByEmailConfirmationToken(context.Background(), reel.EmailConfirmationToken)
	require.ErrorIs(t, err, datastore.ErrReelNotFound)

	require.NoError(t, reelRepo.CreateReel(context.Background(), reel))

	dbReel, err := reelRepo.GetReelByEmailConfirmationToken(context.Background(), reel.EmailConfirmationToken)
	require.NoError(t, err)

	require.Equal(t, reel, dbReel)
}

func TestGetReelsPaged(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)
	user := seedUser(t, db)

	var reels []datastore.Reel

	for i := 0; i < 8; i++ {
		video := seedVideo(t, db)
		reel := generateReel(video.UID, user.UID)

		if i == 0 || i == 2 || i == 4 {
			reel.DeliveryStatus = datastore.ScheduledReelStatus
		} else {
			reel.DeliveryStatus = datastore.UnconfirmedReelStatus
		}

		require.NoError(t, reelRepo.CreateReel(context.Background(), reel))
		reels = append(reels, *reel)
	}

	// Fetch with empty filter
	pageable := datastore.Pageable{PerPage: 20, Cursor: "7ZZZZZZZZZZZZZZZZZZZZZZZZZ"}
	filter := datastore.ReelFilter{}
	reels, PaginationData, err := reelRepo.GetReelsPaged(context.Background(), user.UID, filter, pageable)
	require.NoError(t, err)

	require.Len(t, reels, 8)
	require.False(t, PaginationData.HasMorePages)

	// Filter with delivery status scheduled
	pageable = datastore.Pageable{PerPage: 4, Cursor: "7ZZZZZZZZZZZZZZZZZZZZZZZZZ"}
	filter = datastore.ReelFilter{DeliveryStatus: datastore.ScheduledReelStatus}
	reels, PaginationData, err = reelRepo.GetReelsPaged(context.Background(), user.UID, filter, pageable)
	require.NoError(t, err)

	require.Len(t, reels, 3)
	require.False(t, PaginationData.HasMorePages)

	// Filter with delivery status unconfirmed
	pageable = datastore.Pageable{PerPage: 3, Cursor: "7ZZZZZZZZZZZZZZZZZZZZZZZZZ"}
	filter = datastore.ReelFilter{DeliveryStatus: datastore.UnconfirmedReelStatus}
	reels, PaginationData, err = reelRepo.GetReelsPaged(context.Background(), user.UID, filter, pageable)
	require.NoError(t, err)

	require.Len(t, reels, 3)
	require.True(t, PaginationData.HasMorePages)
	require.NotEmpty(t, PaginationData.Cursor)

}

func TestAssignReelsToUserByEmail(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)
	user := seedUser(t, db)

	var reels []datastore.Reel
	for i := 0; i < 8; i++ {
		video := seedVideo(t, db)
		reel := generateReel(video.UID, "")
		reel.UserID = null.NewString("", false)

		if i == 0 || i == 2 || i == 4 {
			reel.Email = user.Email
		}

		require.NoError(t, reelRepo.CreateReel(context.Background(), reel))
		reels = append(reels, *reel)
	}

	defaultPerpage := 10

	pageable := datastore.Pageable{Cursor: "7ZZZZZZZZZZZZZZZZZZZZZZZZZ", PerPage: defaultPerpage}
	filter := datastore.ReelFilter{}
	reels, _, err := reelRepo.GetReelsPaged(context.Background(), user.UID, filter, pageable)

	require.NoError(t, err)
	require.Len(t, reels, 0)

	err = reelRepo.AssignReelsToUserByEmail(context.Background(), user.Email, user.UID)
	require.NoError(t, err)

	pageable = datastore.Pageable{Cursor: "7ZZZZZZZZZZZZZZZZZZZZZZZZZ", PerPage: defaultPerpage}
	filter = datastore.ReelFilter{}
	reels, _, err = reelRepo.GetReelsPaged(context.Background(), user.UID, filter, pageable)

	require.NoError(t, err)
	require.Len(t, reels, 3)
}

func TestUpdateReel(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)

	user := seedUser(t, db)
	video := seedVideo(t, db)
	reel := generateReel(video.UID, user.UID)

	require.NoError(t, reelRepo.CreateReel(context.Background(), reel))

	updatedReel := &datastore.Reel{
		UID:                    reel.UID,
		UserID:                 null.NewString(user.UID, true),
		VideoID:                video.UID,
		Email:                  "fakemail@gmail.com",
		Title:                  "Test",
		Description:            "Test Update Reel",
		Private:                false,
		EmailConfirmationToken: "jbkl",
		DeliveryStatus:         datastore.DeliveredReelStatus,
		DeliveryDate:           time.Now().Add(time.Hour * 24 * 7).UTC(),
	}

	require.NoError(t, reelRepo.UpdateReel(context.Background(), updatedReel))

	dbReel, err := reelRepo.GetReelByID(context.Background(), reel.UID)
	require.NoError(t, err)

	dbReel.Recipients = nil
	dbReel.CreatedAt = time.Time{}
	dbReel.UpdatedAt = time.Time{}
	dbReel.DeliveryDate = time.Time{}

	updatedReel.DeliveryDate = time.Time{}

	require.Equal(t, updatedReel.Title, dbReel.Title)
}

func TestDeleteReel(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)

	user := seedUser(t, db)
	video := seedVideo(t, db)
	reel := generateReel(video.UID, user.UID)

	require.NoError(t, reelRepo.CreateReel(context.Background(), reel))

	require.NoError(t, reelRepo.DeleteReel(context.Background(), reel.UID))

	_, err := reelRepo.GetReelByID(context.Background(), reel.UID)
	require.ErrorIs(t, err, datastore.ErrReelNotFound)
}

func TestAddRecipients(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)

	user := seedUser(t, db)
	video := seedVideo(t, db)
	reel := generateReel(video.UID, user.UID)

	require.NoError(t, reelRepo.CreateReel(context.Background(), reel))

	recipients := generateRecipients(4)

	require.NoError(t, reelRepo.AddRecipients(context.Background(), reel, recipients))

	dbReel, err := reelRepo.GetReelByID(context.Background(), reel.UID)
	require.NoError(t, err)

	// Check if new recipients were added
	for _, recipient := range recipients {
		dbRecipient := dbReel.FindRecipient(recipient.UID)
		require.NotNil(t, dbRecipient)

		dbRecipient.CreatedAt = time.Time{}
		recipient.CreatedAt = time.Time{}

		require.Equal(t, recipient, *dbRecipient)
	}

	// Check if existing recipients are still in the db
	for _, recipient := range reel.Recipients {
		dbRecipient := dbReel.FindRecipient(recipient.UID)
		require.NotNil(t, dbRecipient)

		dbRecipient.CreatedAt = time.Time{}
		recipient.CreatedAt = time.Time{}

		require.Equal(t, recipient, *dbRecipient)
	}

}

func TestDeleteRecipient(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)

	user := seedUser(t, db)
	video := seedVideo(t, db)
	reel := generateReel(video.UID, user.UID)

	require.NoError(t, reelRepo.CreateReel(context.Background(), reel))

	recipient := reel.Recipients[0]
	require.NoError(t, reelRepo.DeleteRecipient(context.Background(), reel, recipient.UID))

	dbReel, err := reelRepo.GetReelByID(context.TODO(), reel.UID)
	require.NoError(t, err)

	dbRecipient := dbReel.FindRecipient(recipient.UID)
	require.Nil(t, dbRecipient)
}

func generateReel(videoID, userID string) *datastore.Reel {
	return &datastore.Reel{
		UID:                    ulid.Make().String(),
		UserID:                 null.NewString(userID, true),
		VideoID:                videoID,
		Email:                  fmt.Sprintf("%s@memoreel.com", ulid.Make().String()),
		Title:                  "Test Reel",
		Description:            "",
		Private:                true,
		Recipients:             generateRecipients(2),
		EmailConfirmationToken: ulid.Make().String(),
		DeliveryStatus:         datastore.UnconfirmedReelStatus,
		DeliveryDate:           time.Now().Add(time.Hour * 24 * 4),
	}
}

func generateRecipients(n int) datastore.Recipients {
	if n <= 0 {
		return nil
	}

	recipients := make(datastore.Recipients, n)
	for i := range recipients {
		recipients[i] = datastore.Recipient{
			UID:       ulid.Make().String(),
			Email:     fmt.Sprintf("recipient_%s@gmail.com", ulid.Make().String()),
			CreatedAt: time.Now(),
		}
	}

	return recipients
}

func seedUser(t *testing.T, db database.Database) *datastore.User {
	user := &datastore.User{
		UID:                        ulid.Make().String(),
		Firstname:                  "test",
		Lastname:                   "user",
		Password:                   "demopassword",
		Email:                      fmt.Sprintf("%s@gmail.com", ulid.Make().String()),
		EmailVerificationToken:     ulid.Make().String(),
		ResetPasswordToken:         ulid.Make().String(),
		EmailVerificationExpiresAt: null.NewTime(time.Now(), true),
		ResetPasswordExpiresAt:     null.NewTime(time.Time{}, false),
		EmailVerified:              false,
	}

	userRepo := NewUserRepo(db)

	err := userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	return user
}

func seedVideo(t *testing.T, db database.Database) *datastore.Video {
	video := &datastore.Video{
		UID:        ulid.Make().String(),
		Key:        ulid.Make().String(),
		FileFormat: "mp4",
		SizeMB:     20,
	}

	videoRepo := NewVideoRepo(db)

	err := videoRepo.CreateVideo(context.Background(), video)
	require.NoError(t, err)

	return video
}
