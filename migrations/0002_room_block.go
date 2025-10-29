package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			// Ensure block column exists on rooms
			if _, err := db.ExecContext(ctx, `ALTER TABLE rooms ADD COLUMN IF NOT EXISTS block TEXT`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				UPDATE rooms
				SET block = COALESCE(block, '')
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				DO $$
				BEGIN
					IF EXISTS (
						SELECT 1 FROM information_schema.columns
						WHERE table_name = 'rooms'
						  AND column_name = 'block'
					) THEN
						ALTER TABLE rooms ALTER COLUMN block SET NOT NULL;
					END IF;
				END$$
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				DO $$
				BEGIN
					IF EXISTS (
						SELECT 1 FROM pg_constraint
						WHERE conname = 'rooms_room_number_key'
					) THEN
						ALTER TABLE rooms DROP CONSTRAINT rooms_room_number_key;
					END IF;
				END$$
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'unique_room_per_block'
					) THEN
						ALTER TABLE rooms ADD CONSTRAINT unique_room_per_block UNIQUE (block, room_number);
					END IF;
				END$$
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'unique_room_block_id'
					) THEN
						ALTER TABLE rooms ADD CONSTRAINT unique_room_block_id UNIQUE (id, block);
					END IF;
				END$$
			`); err != nil {
				return err
			}

			// Add block column to room_members
			if _, err := db.ExecContext(ctx, `ALTER TABLE room_members ADD COLUMN IF NOT EXISTS block TEXT`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				UPDATE room_members rm
				SET block = r.block
				FROM rooms r
				WHERE rm.room_id = r.id
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				DO $$
				BEGIN
					IF EXISTS (
						SELECT 1 FROM information_schema.columns
						WHERE table_name = 'room_members' AND column_name = 'block'
					) THEN
						ALTER TABLE room_members ALTER COLUMN block SET NOT NULL;
					END IF;
				END$$
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				DO $$
				BEGIN
					IF EXISTS (
						SELECT 1 FROM pg_constraint
						WHERE conname = 'room_members_pkey'
					) THEN
						ALTER TABLE room_members DROP CONSTRAINT room_members_pkey;
					END IF;
				END$$
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				ALTER TABLE room_members
				ADD CONSTRAINT room_members_pkey PRIMARY KEY (room_id, block, user_id)
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				DO $$
				BEGIN
					IF EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'room_members_room_id_fkey'
					) THEN
						ALTER TABLE room_members DROP CONSTRAINT room_members_room_id_fkey;
					END IF;
				END$$
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_room_members_room'
					) THEN
						ALTER TABLE room_members
							ADD CONSTRAINT fk_room_members_room FOREIGN KEY (room_id, block)
								REFERENCES rooms(id, block) ON DELETE CASCADE;
					END IF;
				END$$
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_room_members_user'
					) THEN
						ALTER TABLE room_members
							ADD CONSTRAINT fk_room_members_user FOREIGN KEY (user_id)
								REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END$$
			`); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `
				ALTER TABLE room_members DROP CONSTRAINT IF EXISTS fk_room_members_room;
				ALTER TABLE room_members DROP CONSTRAINT IF EXISTS room_members_pkey;
				ALTER TABLE room_members ADD CONSTRAINT room_members_pkey PRIMARY KEY (room_id, user_id);
				ALTER TABLE room_members DROP COLUMN IF EXISTS block;
			`); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
				ALTER TABLE rooms DROP CONSTRAINT IF EXISTS unique_room_block_id;
				ALTER TABLE rooms DROP CONSTRAINT IF EXISTS unique_room_per_block;
				ALTER TABLE rooms ADD CONSTRAINT rooms_room_number_key UNIQUE (room_number);
				ALTER TABLE rooms DROP COLUMN IF EXISTS block;
			`); err != nil {
				return err
			}

			return nil
		},
	)
}
