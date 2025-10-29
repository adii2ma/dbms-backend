-- ==============================
-- USERS TABLE
-- ==============================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    block TEXT,
    room_name TEXT,
    phone TEXT,
    created_at TIMESTAMP DEFAULT now()
);

-- ==============================
-- ROOMS TABLE
-- ==============================
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    block TEXT NOT NULL,
    room_number TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    CONSTRAINT unique_room_per_block UNIQUE (block, room_number),
    CONSTRAINT unique_room_block_id UNIQUE (id, block)
);

-- ==============================
-- ROOM MEMBERS (WEAK ENTITY)
-- ==============================
CREATE TABLE room_members (
    room_id INT NOT NULL,
    block TEXT NOT NULL,
    user_id UUID NOT NULL,
    joined_at TIMESTAMP DEFAULT now(),
    PRIMARY KEY (room_id, block, user_id),
    CONSTRAINT fk_room_members_room FOREIGN KEY (room_id, block) REFERENCES rooms(id, block) ON DELETE CASCADE,
    CONSTRAINT fk_room_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- ==============================
-- REQUESTS TABLE
-- ==============================
CREATE TABLE requests (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    room_id INT REFERENCES rooms(id) ON DELETE CASCADE,
    type TEXT CHECK (type IN ('cleaning', 'maintenance')) NOT NULL,
    status TEXT CHECK (status IN ('active', 'completed', 'cancelled')) DEFAULT 'active',
    description TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- ==============================
-- CONSTRAINTS
-- ==============================
-- One active request per room per type
CREATE UNIQUE INDEX unique_active_request_per_room_type
ON requests (room_id, type)
WHERE status = 'active';
