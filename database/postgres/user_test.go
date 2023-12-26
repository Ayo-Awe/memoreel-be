package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ayo-awe/memoreel-be/datastore"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestCreateUser(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	userRepo := NewUserRepo(db)
	newUser := &datastore.User{
		UID:                        ulid.Make().String(),
		Firstname:                  "test",
		Lastname:                   "user",
		Password:                   "fakepassword",
		Email:                      "fakeemail@gmail.com",
		EmailVerificationToken:     "dskjfkli",
		ResetPasswordToken:         "dskjpijkilii",
		EmailVerificationExpiresAt: null.NewTime(time.Now(), true),
		ResetPasswordExpiresAt:     null.NewTime(time.Time{}, false),
		EmailVerified:              false,
	}

	// Create user no error
	err := userRepo.CreateUser(context.Background(), newUser)
	require.NoError(t, err)

	// Duplicate user email
	userWithExistingEmail := &datastore.User{
		UID:                        ulid.Make().String(),
		Firstname:                  "test",
		Lastname:                   "user",
		Password:                   "demopassword",
		Email:                      "fakeemail@gmail.com",
		EmailVerificationToken:     "eijowjee",
		ResetPasswordToken:         "fjijoweie",
		EmailVerificationExpiresAt: null.NewTime(time.Now(), true),
		ResetPasswordExpiresAt:     null.NewTime(time.Time{}, false),
		EmailVerified:              false,
	}

	err = userRepo.CreateUser(context.Background(), userWithExistingEmail)
	require.Error(t, err)
	require.ErrorIs(t, err, datastore.ErrDuplicateUserEmail)
}

func TestGetUserByID(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	userRepo := NewUserRepo(db)
	user := generateUser()

	_, err := userRepo.GetUserByID(context.Background(), user.UID)
	require.ErrorIs(t, err, datastore.ErrUserNotFound)

	err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	foundUser, err := userRepo.GetUserByID(context.Background(), user.UID)
	require.NoError(t, err)

	require.NotEmpty(t, foundUser)
	require.NotNil(t, foundUser)

	require.Equal(t, user, foundUser)
}

func TestGetUserByEmail(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	userRepo := NewUserRepo(db)
	user := generateUser()

	_, err := userRepo.GetUserByEmail(context.Background(), user.Email)
	require.ErrorIs(t, err, datastore.ErrUserNotFound)

	err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	foundUser, err := userRepo.GetUserByEmail(context.Background(), user.Email)
	require.NoError(t, err)

	require.NotEmpty(t, foundUser)
	require.NotNil(t, foundUser)

	require.Equal(t, user, foundUser)
}

func TestGetUserByEmailVerificationToken(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	userRepo := NewUserRepo(db)
	user := generateUser()

	_, err := userRepo.GetUserByEmailVerificationToken(context.Background(), user.EmailVerificationToken)
	require.ErrorIs(t, err, datastore.ErrUserNotFound)

	err = userRepo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	foundUser, err := userRepo.GetUserByEmailVerificationToken(context.Background(), user.EmailVerificationToken)
	require.NoError(t, err)

	require.NotEmpty(t, foundUser)
	require.NotNil(t, foundUser)

	require.Equal(t, user, foundUser)
}

func TestGetUserByResetPasswordToken(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	userRepo := NewUserRepo(db)
	user := generateUser()

	_, err := userRepo.GetUserByResetPasswordToken(context.Background(), user.ResetPasswordToken)
	require.ErrorIs(t, err, datastore.ErrUserNotFound)

	require.NoError(t, userRepo.CreateUser(context.Background(), user))

	foundUser, err := userRepo.GetUserByResetPasswordToken(context.Background(), user.ResetPasswordToken)
	require.NoError(t, err)

	require.NotEmpty(t, foundUser)
	require.NotNil(t, foundUser)

	require.Equal(t, user, foundUser)
}

func TestUpdateUser(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	userRepo := NewUserRepo(db)
	user := generateUser()

	require.NoError(t, userRepo.CreateUser(context.Background(), user))

	updatedUser := &datastore.User{
		UID:                        user.UID,
		Firstname:                  fmt.Sprintf("first_name_%s", ulid.Make().String()),
		Lastname:                   fmt.Sprintf("last_name_%s", ulid.Make().String()),
		Password:                   fmt.Sprintf("password_%s", ulid.Make().String()),
		Email:                      fmt.Sprintf("%s@gmail.com", ulid.Make().String()),
		EmailVerificationToken:     ulid.Make().String(),
		ResetPasswordToken:         ulid.Make().String(),
		EmailVerificationExpiresAt: null.NewTime(time.Now().Add(time.Hour).UTC(), true),
		ResetPasswordExpiresAt:     null.NewTime(time.Now().Add(time.Hour).UTC(), true),
		EmailVerified:              true,
	}

	err := userRepo.UpdateUser(context.Background(), updatedUser)
	require.NoError(t, err)

	require.NotEqual(t, user.UpdatedAt, updatedUser.UpdatedAt)

	dbUser, err := userRepo.GetUserByID(context.Background(), user.UID)
	require.NoError(t, err)

	dbUser.UpdatedAt = time.Time{}
	dbUser.CreatedAt = time.Time{}

	require.InDelta(t, dbUser.EmailVerificationExpiresAt.Time.Unix(), updatedUser.EmailVerificationExpiresAt.Time.Unix(), float64(time.Second))
	require.InDelta(t, dbUser.ResetPasswordExpiresAt.Time.Unix(), updatedUser.ResetPasswordExpiresAt.Time.Unix(), float64(time.Second))

	dbUser.EmailVerificationExpiresAt = null.NewTime(time.Time{}, false)
	dbUser.ResetPasswordExpiresAt = null.NewTime(time.Time{}, false)
	updatedUser.EmailVerificationExpiresAt = null.NewTime(time.Time{}, false)
	updatedUser.ResetPasswordExpiresAt = null.NewTime(time.Time{}, false)

	require.Equal(t, dbUser, updatedUser)
}

func TestDeleteUser(t *testing.T) {
	db, closeFn := getDB(t)
	defer closeFn()

	userRepo := NewUserRepo(db)
	user := generateUser()

	require.NoError(t, userRepo.CreateUser(context.Background(), user))

	err := userRepo.DeleteUser(context.Background(), user.UID)
	require.NoError(t, err)

	_, err = userRepo.GetUserByID(context.Background(), user.UID)
	require.ErrorIs(t, err, datastore.ErrUserNotFound)
}

func generateUser() *datastore.User {
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

	return user
}
