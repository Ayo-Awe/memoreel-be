package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ayo-awe/memoreel-be/database"
	"github.com/ayo-awe/memoreel-be/datastore"
	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotCreated = errors.New("user could not be created")
	ErrUserNotUpdated = errors.New("user could not be updated")
	ErrUserNotDeleted = errors.New("user could not be deleted")
)

const (
	createUser = `
	INSERT INTO users (
		id, first_name, last_name,
		email, password, email_verified, reset_password_token,
		email_verification_token, reset_password_expires_at,
		email_verification_expires_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	RETURNING *;
	`

	updateUser = `
	UPDATE users SET
		first_name = $2,
		last_name = $3,
		email = $4,
		password =$5,
		email_verified = $6,
		reset_password_token = $7,
		email_verification_token = $8,
		reset_password_expires_at = $9,
		email_verification_expires_at = $10,
		updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	`

	deleteUser = `
	UPDATE users
	SET deleted_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL
	`

	fetchUser = `
	SELECT
		id,
		first_name,
		last_name,
		email,
		password,
		email_verified,
		reset_password_token,
		email_verification_token,
		reset_password_expires_at,
		email_verification_expires_at,
		created_at,
		updated_at,
		deleted_at
	FROM users
	WHERE %s = $1 AND deleted_at IS NULL
	`
)

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(d database.Database) datastore.UserRepository {
	return &userRepo{db: d.GetDB()}
}

func (u userRepo) GetUserByID(ctx context.Context, userID string) (*datastore.User, error) {
	user := &datastore.User{}

	err := u.db.QueryRowxContext(ctx, fmt.Sprintf(fetchUser, "id"), userID).StructScan(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datastore.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (u userRepo) GetUserByEmail(ctx context.Context, userID string) (*datastore.User, error) {
	user := &datastore.User{}

	err := u.db.QueryRowxContext(ctx, fmt.Sprintf(fetchUser, "email"), userID).StructScan(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datastore.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (u userRepo) GetUserByEmailVerificationToken(ctx context.Context, userID string) (*datastore.User, error) {
	user := &datastore.User{}

	err := u.db.QueryRowxContext(ctx, fmt.Sprintf(fetchUser, "email_verification_token"), userID).StructScan(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datastore.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (u userRepo) GetUserByResetPasswordToken(ctx context.Context, userID string) (*datastore.User, error) {
	user := &datastore.User{}

	err := u.db.QueryRowxContext(ctx, fmt.Sprintf(fetchUser, "reset_password_token"), userID).StructScan(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datastore.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (u userRepo) CreateUser(ctx context.Context, user *datastore.User) error {

	row := u.db.QueryRowxContext(ctx, createUser,
		user.UID,
		user.Firstname,
		user.Lastname,
		user.Email,
		user.Password,
		user.EmailVerified,
		user.ResetPasswordToken,
		user.EmailVerificationToken,
		user.ResetPasswordExpiresAt,
		user.EmailVerificationExpiresAt)

	if err := row.StructScan(user); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return datastore.ErrDuplicateUserEmail
		}
		return err
	}

	return nil
}

func (u userRepo) UpdateUser(ctx context.Context, user *datastore.User) error {

	res, err := u.db.ExecContext(ctx, updateUser,
		user.UID,
		user.Firstname,
		user.Lastname,
		user.Email,
		user.Password,
		user.EmailVerified,
		user.ResetPasswordToken,
		user.EmailVerificationToken,
		user.ResetPasswordExpiresAt,
		user.EmailVerificationExpiresAt)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		return ErrUserNotUpdated
	}

	return nil
}

func (u userRepo) DeleteUser(ctx context.Context, userID string) error {

	res, err := u.db.ExecContext(ctx, deleteUser, userID)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		return ErrUserNotDeleted
	}

	return nil
}
