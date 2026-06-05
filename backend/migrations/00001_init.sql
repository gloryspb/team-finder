-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE player_profiles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nickname TEXT NOT NULL,
    region TEXT NOT NULL DEFAULT '',
    languages TEXT[] NOT NULL DEFAULT '{}',
    voice_chat BOOLEAN NOT NULL DEFAULT FALSE,
    bio TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE games (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    modes TEXT[] NOT NULL DEFAULT '{}',
    roles TEXT[] NOT NULL DEFAULT '{}'
);

CREATE TABLE listings (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    mode TEXT NOT NULL DEFAULT '',
    required_roles TEXT[] NOT NULL DEFAULT '{}',
    rank_min TEXT NOT NULL DEFAULT '',
    rank_max TEXT NOT NULL DEFAULT '',
    region TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT listings_status_check CHECK (status IN ('open', 'closed'))
);

CREATE TABLE applications (
    id UUID PRIMARY KEY,
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    applicant_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (listing_id, applicant_id),
    CONSTRAINT applications_status_check CHECK (status IN ('pending', 'accepted', 'rejected'))
);

CREATE INDEX idx_listings_filters ON listings (game_id, status, region, mode);
CREATE INDEX idx_applications_applicant ON applications (applicant_id);

INSERT INTO games (id, name, modes, roles) VALUES
('11111111-1111-1111-1111-111111111111', 'Dota 2', ARRAY['All Pick', 'Ranked', 'Turbo', 'Captain Mode'], ARRAY['Carry', 'Mid', 'Offlane', 'Support', 'Hard Support']),
('22222222-2222-2222-2222-222222222222', 'Counter-Strike 2', ARRAY['Premier', 'Competitive', 'Wingman'], ARRAY['Entry Fragger', 'AWPer', 'Support', 'Lurker', 'IGL']),
('33333333-3333-3333-3333-333333333333', 'Valorant', ARRAY['Competitive', 'Unrated', 'Premier'], ARRAY['Duelist', 'Controller', 'Initiator', 'Sentinel', 'Flex']),
('44444444-4444-4444-4444-444444444444', 'League of Legends', ARRAY['Ranked Solo/Duo', 'Ranked Flex', 'Normal Draft'], ARRAY['Top', 'Jungle', 'Mid', 'ADC', 'Support']),
('55555555-5555-5555-5555-555555555555', 'Overwatch 2', ARRAY['Competitive', 'Quick Play', 'Arcade'], ARRAY['Tank', 'Damage', 'Support', 'Flex']),
('66666666-6666-6666-6666-666666666666', 'Apex Legends', ARRAY['Ranked', 'Trios', 'Duos'], ARRAY['Entry', 'Support', 'Recon', 'Controller', 'Flex'])
ON CONFLICT (name) DO UPDATE SET modes = EXCLUDED.modes, roles = EXCLUDED.roles;

-- +goose Down
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS listings;
DROP TABLE IF EXISTS games;
DROP TABLE IF EXISTS player_profiles;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";
