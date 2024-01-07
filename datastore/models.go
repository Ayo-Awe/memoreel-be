package datastore

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gopkg.in/guregu/null.v4"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrDuplicateUserEmail = errors.New("a user with this email already exists")
)

type User struct {
	UID                        string    `json:"id" db:"id"`
	Firstname                  string    `json:"first_name" db:"first_name"`
	Lastname                   string    `json:"last_name" db:"last_name"`
	Email                      string    `json:"email" db:"email"`
	Password                   string    `json:"-" db:"password"`
	EmailVerified              bool      `json:"email_verified" db:"email_verified"`
	ResetPasswordToken         string    `json:"-" db:"reset_password_token,omitempty"`
	EmailVerificationToken     string    `json:"-" db:"email_verification_token,omitempty"`
	ResetPasswordExpiresAt     null.Time `json:"-" db:"reset_password_expires_at,omitempty"`
	EmailVerificationExpiresAt null.Time `json:"-" db:"email_verification_expires_at,omitempty"`
	CreatedAt                  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt                  null.Time `json:"deleted_at,omitempty" db:"deleted_at,omitempty"`
}

var (
	ErrVideoNotFound = errors.New("video not found")
)

type Video struct {
	UID        string    `json:"uid" db:"id"`
	Key        string    `json:"-" db:"key"`
	FileFormat string    `json:"file_format" db:"file_format"`
	SizeMB     float32   `json:"size_md" db:"size_mb"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt  null.Time `json:"deleted_at" db:"deleted_at"`
}

type Recipient struct {
	UID       string    `json:"uid" db:"id"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	DeletedAt null.Time `json:"deleted_at" db:"deleted_at"`
}

type (
	ReelDeliveryStatus string
	Recipients         []Recipient
)

func (r Recipients) Value() (driver.Value, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	// database driver should treat nil array as empty array in db
	if string(b) == "null" {
		return []byte("[]"), nil
	}

	return b, nil
}

func (r *Recipients) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	if string(b) == "[]" {
		return nil
	}

	var recipients Recipients

	err := json.Unmarshal(b, &recipients)
	if err != nil {
		return err
	}

	// filter out deleted recipients
	var notDeleted Recipients
	for _, recipient := range recipients {
		if recipient.DeletedAt.IsZero() {
			notDeleted = append(notDeleted, recipient)
		}
	}

	*r = notDeleted

	return nil
}

const (
	UnconfirmedReelStatus ReelDeliveryStatus = "unconfirmed"
	ScheduledReelStatus   ReelDeliveryStatus = "scheduled"
	FailedReelStatus      ReelDeliveryStatus = "failed"
	DeliveredReelStatus   ReelDeliveryStatus = "delivered"
)

func (r ReelDeliveryStatus) IsValid() bool {
	switch r {

	case UnconfirmedReelStatus,
		ScheduledReelStatus,
		FailedReelStatus,
		DeliveredReelStatus:
		return true
	default:
		return false
	}
}

var (
	ErrReelNotFound      = errors.New("reel not found")
	ErrRecipientNotFound = errors.New("recipient not found")
)

type ReelFilter struct {
	DeliveryStatus ReelDeliveryStatus
}

type Reel struct {
	UID                    string             `json:"id" db:"id"`
	UserID                 null.String        `json:"user_id" db:"user_id"`
	VideoID                string             `json:"video_id" db:"video_id"`
	Email                  string             `json:"email" db:"email"`
	Title                  string             `json:"title,omitempty" db:"title,omitempty"`
	Description            string             `json:"description" db:"description"`
	Private                bool               `json:"private" db:"private"`
	Recipients             Recipients         `json:"recipients" db:"recipients"`
	EmailConfirmationToken string             `json:"-" db:"email_confirmation_token"`
	DeliveryStatus         ReelDeliveryStatus `json:"delivery_status" db:"delivery_status"`
	DeliveryDate           time.Time          `json:"delivery_date,omitempty" db:"delivery_date,omitempty"`
	CreatedAt              time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at" db:"updated_at"`
	DeletedAt              null.Time          `json:"deleted_at,omitempty" db:"deleted_at"`
}

func (r Reel) FindRecipient(recipientID string) *Recipient {
	for i := range r.Recipients {
		recipient := r.Recipients[i]

		if recipient.UID == recipientID {
			return &recipient
		}
	}

	return nil
}

type PageDirection string

type Pageable struct {
	PerPage int    `json:"per_page"`
	Cursor  string `json:"cursor"`
}

func (p Pageable) Limit() int {
	return p.PerPage + 1
}

type PreviousRowCount struct {
	Count int
}

type PaginationData struct {
	PerPage      int    `json:"per_page"`
	Cursor       string `json:"cursor"`
	HasMorePages bool   `json:"has_more_pages"`
}

func (p *PaginationData) Build(pageable Pageable, items []string) *PaginationData {
	p.PerPage = pageable.PerPage

	var last string

	if len(items) > 0 {
		last = items[len(items)-1]
	}

	p.Cursor = last

	// an extra exists. It's used to check if there are more pages to be loaded
	p.HasMorePages = len(items) > p.PerPage

	return p
}
