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

type reels []datastore.Reel

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

func TestGetReelsByEmail(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)

	user1 := seedUser(t, db)
	video1 := seedVideo(t, db)
	reel := &datastore.Reel{
		UID:                    ulid.Make().String(),
		UserID:                 user1.UID,
		VideoID:                video1.UID,
		Email:                  "testemail@memoreel.com",
		Title:                  "Test Reel",
		Description:            "",
		Private:                true,
		Recipients:             generateRecipients(2),
		EmailConfirmationToken: ulid.Make().String(),
		DeliveryStatus:         datastore.UnconfirmedReelStatus,
		DeliveryDate:           time.Now().Add(time.Hour * 24 * 4),
	}

	user2 := seedUser(t, db)
	video2 := seedVideo(t, db)
	reelWithMatchingEmail := &datastore.Reel{
		UID:                    ulid.Make().String(),
		UserID:                 user2.UID,
		VideoID:                video2.UID,
		Email:                  "testemail@memoreel.com",
		Title:                  "Test Reel",
		Description:            "",
		Private:                true,
		Recipients:             generateRecipients(2),
		EmailConfirmationToken: ulid.Make().String(),
		DeliveryStatus:         datastore.UnconfirmedReelStatus,
		DeliveryDate:           time.Now().Add(time.Hour * 24 * 4),
	}

	user3 := seedUser(t, db)
	video3 := seedVideo(t, db)
	reelWithDifferentEmail := generateReel(video3.UID, user3.UID)

	res, err := reelRepo.GetReelsByEmail(context.Background(), reel.Email)
	require.NoError(t, err)
	require.True(t, len(res) == 0)

	require.NoError(t, reelRepo.CreateReel(context.Background(), reel))
	require.NoError(t, reelRepo.CreateReel(context.Background(), reelWithMatchingEmail))
	require.NoError(t, reelRepo.CreateReel(context.Background(), reelWithDifferentEmail))

	foundReels, err := reelRepo.GetReelsByEmail(context.Background(), reel.Email)
	require.NoError(t, err)
	require.Equal(t, len(foundReels), 2)

	fr := reels(foundReels)
	require.NotNil(t, fr.find(reel.UID))
	require.NotNil(t, fr.find(reelWithMatchingEmail.UID))

}

func TestGetReelsByUserID(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	reelRepo := NewReelRepo(db)

	user1 := seedUser(t, db)
	video1 := seedVideo(t, db)
	reel := generateReel(video1.UID, user1.UID)

	video2 := seedVideo(t, db)
	reelWithMatchingUserID := generateReel(video2.UID, user1.UID)

	user2 := seedUser(t, db)
	video3 := seedVideo(t, db)
	reelWithDifferentUserID := generateReel(video3.UID, user2.UID)

	res, err := reelRepo.GetReelsByUserID(context.Background(), reel.UserID)
	require.NoError(t, err)
	require.True(t, len(res) == 0)

	require.NoError(t, reelRepo.CreateReel(context.Background(), reel))
	require.NoError(t, reelRepo.CreateReel(context.Background(), reelWithMatchingUserID))
	require.NoError(t, reelRepo.CreateReel(context.Background(), reelWithDifferentUserID))

	foundReels, err := reelRepo.GetReelsByUserID(context.Background(), reel.UserID)
	require.NoError(t, err)
	require.Equal(t, len(foundReels), 2)

	fr := reels(foundReels)
	require.NotNil(t, fr.find(reel.UID))
	require.NotNil(t, fr.find(reelWithMatchingUserID.UID))
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
		UserID:                 user.UID,
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
		UserID:                 userID,
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

func (r reels) find(id string) *datastore.Reel {
	for i := range r {
		reel := &r[i]
		if reel.UID == id {
			return reel
		}
	}

	return nil
}
