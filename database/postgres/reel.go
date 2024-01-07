package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ayo-awe/memoreel-be/database"
	"github.com/ayo-awe/memoreel-be/datastore"
	"github.com/jmoiron/sqlx"
)

var (
	ErrReelNotCreated          = errors.New("reel could not be created")
	ErrReelNotUpdated          = errors.New("reel could not be updated")
	ErrReelRecipientNotDeleted = errors.New("reel recipient could not be deleted")
	ErrReelRecipientsNotAdded  = errors.New("reel recipients could not added")
	ErrReelNotDeleted          = errors.New("reel could not be deleted")
)

const (
	createReel = `
	INSERT INTO reels (
		id, user_id, video_id, email,
		title, description, private, recipients,
		email_confirmation_token, delivery_status, delivery_date
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	RETURNING *;
	`

	fetchReel = `
	SELECT
		id,
		user_id,
		video_id,
		email,
		title,
		description,
		private,
		recipients,
		email_confirmation_token,
		delivery_status,
		delivery_date,
		updated_at,
		created_at,
		deleted_at
	FROM reels
	WHERE %s = $1 AND deleted_at IS NULL;
	`

	fetchReelsPaged = `
	SELECT
		id,
		user_id,
		video_id,
		email,
		title,
		description,
		private,
		recipients,
		email_confirmation_token,
		delivery_status,
		delivery_date,
		updated_at,
		created_at,
		deleted_at
	FROM reels
	WHERE deleted_at IS NULL
	%s
	AND id < :cursor
	ORDER BY id DESC
	LIMIT :limit;
	`

	baseReelsFilter = `
	AND user_id = :user_id
	`

	reelDeliveryStatusFilter = `
	%s
	AND delivery_status = :delivery_status
	`

	updateReel = `
	UPDATE reels SET
		user_id = $2,
		video_id = $3,
		email = $4,
		title = $5,
		description = $6,
		private = $7,
		delivery_status = $8,
		delivery_date = $9,
		email_confirmation_token = $10,
		updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL;
	`

	assignReelsToUserByEmail = `
	UPDATE reels SET
		user_id = $2
	WHERE user_id IS NULL
	AND email = $1
	AND deleted_at IS NULL;
	`

	// postgres jsonb array concatenation
	addRecipients = `
	UPDATE reels SET
		recipients = recipients || $2,
		updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL;
	`

	deleteRecipient = `
	UPDATE reels
		SET recipients = (
			SELECT jsonb_agg(
				CASE
					WHEN r->>'uid' = $2 AND r->>'deleted_at' IS NULL THEN jsonb_set(r, '{deleted_at}', to_jsonb(NOW()))
					ELSE r
				END
			)
			FROM jsonb_array_elements(recipients) r
		)
	WHERE id = $1 AND deleted_at IS NULL;
	`

	deleteReel = `
	UPDATE reels SET
		deleted_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL;`
)

type reelRepo struct {
	db *sqlx.DB
}

func NewReelRepo(db database.Database) datastore.ReelRepository {
	return &reelRepo{db: db.GetDB()}
}

func (r reelRepo) GetReelByID(ctx context.Context, id string) (*datastore.Reel, error) {
	reel := &datastore.Reel{}
	err := r.db.QueryRowxContext(ctx, fmt.Sprintf(fetchReel, "id"), id).StructScan(reel)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datastore.ErrReelNotFound
		}
		return nil, err
	}

	return reel, nil
}

func (r reelRepo) GetReelByEmailConfirmationToken(ctx context.Context, token string) (*datastore.Reel, error) {
	reel := &datastore.Reel{}
	err := r.db.QueryRowxContext(ctx, fmt.Sprintf(fetchReel, "email_confirmation_token"), token).StructScan(reel)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datastore.ErrReelNotFound
		}
		return nil, err
	}

	return reel, nil
}

func (r reelRepo) GetReelsPaged(ctx context.Context, userID string, filter datastore.ReelFilter, pageable datastore.Pageable) ([]datastore.Reel, datastore.PaginationData, error) {
	queryFilter := baseReelsFilter
	args := map[string]interface{}{
		"user_id": userID,
		"cursor":  pageable.Cursor,
		"limit":   pageable.Limit(),
	}

	if filter.DeliveryStatus.IsValid() {
		args["delivery_status"] = filter.DeliveryStatus
		queryFilter = fmt.Sprintf(reelDeliveryStatusFilter, queryFilter)
	}

	query := fetchReelsPaged
	query = fmt.Sprintf(query, queryFilter)

	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, datastore.PaginationData{}, err
	}

	var reels []datastore.Reel
	for rows.Next() {
		reel := datastore.Reel{}
		err = rows.StructScan(&reel)

		if err != nil {
			return nil, datastore.PaginationData{}, err
		}

		reels = append(reels, reel)
	}

	ids := make([]string, len(reels))
	for i := range reels {
		ids[i] = reels[i].UID
	}

	if len(reels) > pageable.PerPage {
		reels = reels[:len(reels)-1]
	}

	pagination := &datastore.PaginationData{}
	pagination.Build(pageable, ids)

	return reels, *pagination, nil
}

func (r reelRepo) CreateReel(ctx context.Context, reel *datastore.Reel) error {

	row := r.db.QueryRowxContext(ctx, createReel,
		reel.UID,
		reel.UserID,
		reel.VideoID,
		reel.Email,
		reel.Title,
		reel.Description,
		reel.Private,
		reel.Recipients,
		reel.EmailConfirmationToken,
		reel.DeliveryStatus,
		reel.DeliveryDate,
	)

	err := row.StructScan(reel)

	if err != nil {
		return err
	}

	return nil
}

func (r reelRepo) UpdateReel(ctx context.Context, reel *datastore.Reel) error {
	res, err := r.db.ExecContext(ctx, updateReel,
		reel.UID,
		reel.UserID,
		reel.VideoID,
		reel.Email,
		reel.Title,
		reel.Description,
		reel.Private,
		reel.DeliveryStatus,
		reel.DeliveryDate,
		reel.EmailConfirmationToken)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		return ErrReelNotUpdated
	}

	return nil
}

func (r reelRepo) AssignReelsToUserByEmail(ctx context.Context, email string, userID string) error {
	_, err := r.db.ExecContext(ctx, assignReelsToUserByEmail, email, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r reelRepo) AddRecipients(ctx context.Context, reel *datastore.Reel, newRecipients datastore.Recipients) error {
	reel.Recipients = append(reel.Recipients, newRecipients...)

	res, err := r.db.ExecContext(ctx, addRecipients, reel.UID, newRecipients)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		return ErrReelRecipientsNotAdded
	}

	return nil
}

func (r reelRepo) DeleteRecipient(ctx context.Context, reel *datastore.Reel, recipientID string) error {
	recipient := reel.FindRecipient(recipientID)
	if recipient == nil {
		return datastore.ErrRecipientNotFound
	}

	res, err := r.db.ExecContext(ctx, deleteRecipient, reel.UID, recipientID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		return ErrReelRecipientNotDeleted
	}

	return nil
}

func (r reelRepo) DeleteReel(ctx context.Context, reelID string) error {
	res, err := r.db.ExecContext(ctx, deleteReel, reelID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		return ErrReelNotDeleted
	}

	return nil
}
