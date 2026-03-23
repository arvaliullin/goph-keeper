package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateSecrets, downCreateSecrets)
}

func upCreateSecrets(ctx context.Context, tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS secrets (
			id         UUID PRIMARY KEY,
			user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type       VARCHAR(50) NOT NULL,
			data       BYTEA NOT NULL,
			metadata   BYTEA,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);
		CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id);
		CREATE INDEX IF NOT EXISTS idx_secrets_user_id_updated_at ON secrets(user_id, updated_at DESC)
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}

func downCreateSecrets(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS secrets`)
	return err
}
