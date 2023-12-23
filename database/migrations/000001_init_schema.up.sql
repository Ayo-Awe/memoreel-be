CREATE TABLE IF NOT EXISTS "users" (
	"id" CHAR(26) PRIMARY KEY,
	"first_name" VARCHAR(255) NOT NULL,
	"last_name" VARCHAR(255) NOT NULL,
	"email" VARCHAR(255) NOT NULL,
	"password" TEXT NOT NULL,
	"email_verified" BOOLEAN NOT NULL DEFAULT(false),
	"reset_password_token" VARCHAR(255),
	"email_verification_token" VARCHAR(255),
	"reset_password_expires_at" TIMESTAMPTZ,
	"email_verification_expires_at" TIMESTAMPTZ,
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT(NOW()),
	"updated_at" TIMESTAMPTZ NOT NULL DEFAULT(NOW()),
	"deleted_at" TIMESTAMPTZ,

	CONSTRAINT users_email_key UNIQUE NULLS NOT DISTINCT (email, deleted_at)
);

CREATE TYPE "reel_delivery_status" AS ENUM ('unconfirmed', 'scheduled', 'deivered', 'failed');

CREATE TABLE IF NOT EXISTS "videos" (
	"id" CHAR(26) PRIMARY KEY,
	"key" VARCHAR(255) NOT NULL,
	"file_format" VARCHAR(255)  NOT NULL,
	"size_mb" FLOAT NOT NUll,
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT(NOW()),
	"updated_at" TIMESTAMPTZ NOT NULL DEFAULT(NOW()),
	"deleted_at" TIMESTAMPTZ
);


CREATE TABLE IF NOT EXISTS "reels"(
	"id" CHAR(26) PRIMARY KEY,
	"user_id" CHAR(26) REFERENCES users(id),
	"video_id" CHAR(26) NOT NULL REFERENCES videos(id),
	"email" CHAR(255) NOT NULL,
	"title" CHAR(255) NOT NULL,
	"description" TEXT,
	"private" BOOLEAN NOT NULL DEFAULT(true),
	"email_confirmation_token" CHAR(255),
	"delivery_status" reel_delivery_status NOT NULL DEFAULT ('unconfirmed'),
	"delivery_date" TIMESTAMPTZ NOT NULL,
	"recipients" JSONB NOT NULL,
	"created_at" TIMESTAMPTZ NOT NULL DEFAULT(NOW()),
	"updated_at" TIMESTAMPTZ NOT NULL DEFAULT(NOW()),
	"deleted_at" TIMESTAMPTZ,

    CONSTRAINT reel_video_keys UNIQUE NULLS NOT DISTINCT (video_id, deleted_at)
);

