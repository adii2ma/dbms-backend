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
    room_number TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

-- ==============================
-- ROOM MEMBERS (WEAK ENTITY)
-- ==============================
CREATE TABLE room_members (
    room_id INT REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT now(),
    PRIMARY KEY (room_id, user_id)
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
