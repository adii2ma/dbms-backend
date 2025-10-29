package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			// Ensure UUID generation extension exists
			if _, err := db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "pgcrypto"`); err != nil {
				return err
			}

			// users table
			if _, err := db.ExecContext(ctx, `
                CREATE TABLE IF NOT EXISTS users (
                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                    name TEXT NOT NULL,
                    email TEXT UNIQUE NOT NULL,
                    password TEXT NOT NULL,
                    block TEXT,
                    room_name TEXT,
                    phone TEXT,
                    created_at TIMESTAMP DEFAULT now()
                )
            `); err != nil {
				return err
			}

			// rooms table
			if _, err := db.ExecContext(ctx, `
                CREATE TABLE IF NOT EXISTS rooms (
                    id SERIAL PRIMARY KEY,
                    block TEXT NOT NULL,
                    room_number TEXT NOT NULL,
                    created_at TIMESTAMP DEFAULT now()
                )
            `); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
                DO $$
                BEGIN
                    IF EXISTS (
                        SELECT 1 FROM information_schema.columns
                        WHERE table_name = 'rooms' AND column_name = 'block'
                    ) THEN
                        IF NOT EXISTS (
                            SELECT 1 FROM pg_constraint
                            WHERE conname = 'unique_room_per_block'
                        ) THEN
                            ALTER TABLE rooms
                                ADD CONSTRAINT unique_room_per_block UNIQUE (block, room_number);
                        END IF;
                    END IF;
                END$$
            `); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
                DO $$
                BEGIN
                    IF EXISTS (
                        SELECT 1 FROM information_schema.columns
                        WHERE table_name = 'rooms' AND column_name = 'block'
                    ) THEN
                        IF NOT EXISTS (
                            SELECT 1 FROM pg_constraint
                            WHERE conname = 'unique_room_block_id'
                        ) THEN
                            ALTER TABLE rooms
                                ADD CONSTRAINT unique_room_block_id UNIQUE (id, block);
                        END IF;
                    END IF;
                END$$
            `); err != nil {
				return err
			}

			// room_members table
			if _, err := db.ExecContext(ctx, `
                CREATE TABLE IF NOT EXISTS room_members (
                    room_id INT NOT NULL,
                    block TEXT NOT NULL,
                    user_id UUID NOT NULL,
                    joined_at TIMESTAMP DEFAULT now()
                )
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
                        IF NOT EXISTS (
                            SELECT 1 FROM pg_constraint
                            WHERE conname = 'room_members_pkey'
                        ) THEN
                            ALTER TABLE room_members
                                ADD CONSTRAINT room_members_pkey PRIMARY KEY (room_id, block, user_id);
                        END IF;
                    END IF;
                END$$
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
                        IF NOT EXISTS (
                            SELECT 1 FROM pg_constraint
                            WHERE conname = 'fk_room_members_room'
                        ) THEN
                            ALTER TABLE room_members
                                ADD CONSTRAINT fk_room_members_room FOREIGN KEY (room_id, block)
                                    REFERENCES rooms(id, block) ON DELETE CASCADE;
                        END IF;
                    END IF;
                END$$
            `); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
                DO $$
                BEGIN
                    IF NOT EXISTS (
                        SELECT 1 FROM pg_constraint
                        WHERE conname = 'fk_room_members_user'
                    ) THEN
                        ALTER TABLE room_members
                            ADD CONSTRAINT fk_room_members_user FOREIGN KEY (user_id)
                                REFERENCES users(id) ON DELETE CASCADE;
                    END IF;
                END$$
            `); err != nil {
				return err
			}

			// requests table
			if _, err := db.ExecContext(ctx, `
                CREATE TABLE IF NOT EXISTS requests (
                    id SERIAL PRIMARY KEY,
                    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
                    room_id INT REFERENCES rooms(id) ON DELETE CASCADE,
                    type TEXT CHECK (type IN ('cleaning', 'maintenance')) NOT NULL,
                    status TEXT CHECK (status IN ('active', 'completed', 'cancelled')) DEFAULT 'active',
                    description TEXT,
                    created_at TIMESTAMP DEFAULT now(),
                    updated_at TIMESTAMP DEFAULT now()
                )
            `); err != nil {
				return err
			}

			if _, err := db.ExecContext(ctx, `
                CREATE UNIQUE INDEX IF NOT EXISTS unique_active_request_per_room_type
                ON requests (room_id, type)
                WHERE status = 'active'
            `); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS room_members`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS requests`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS rooms`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS users`); err != nil {
				return err
			}
			return nil
		},
	)
}
