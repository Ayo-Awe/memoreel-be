package datastore

import "context"

type UserRepository interface {
	GetUserByID(context.Context, string) (*User, error)
	GetUserByEmail(context.Context, string) (*User, error)
	GetUserByResetPasswordToken(context.Context, string) (*User, error)
	GetUserByEmailVerificationToken(context.Context, string) (*User, error)
	CreateUser(context.Context, *User) error
	UpdateUser(context.Context, *User) error
	DeleteUser(ctx context.Context, userID string) error
}

type ReelRepository interface {
	GetReelByID(context.Context, string) (*Reel, error)
	GetReelsByUserID(context.Context, string) ([]Reel, error)
	GetReelsByEmail(context.Context, string) ([]Reel, error)
	GetReelByEmailConfirmationToken(context.Context, string) (*Reel, error)
	CreateReel(context.Context, *Reel) error
	UpdateReel(context.Context, *Reel) error
	AddRecipients(ctx context.Context, reel *Reel, recipients Recipients) error
	DeleteRecipient(ctx context.Context, reel *Reel, recipientID string) error
	DeleteReel(ctx context.Context, reelID string) error
}

type VideoRepository interface {
	GetVideoByID(context.Context, string) (*Video, error)
	CreateVideo(context.Context, *Video) error
	UpdateVideo(context.Context, *Video) error
	DeleteVideo(ctx context.Context, videoID string) error
}
