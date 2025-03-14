CREATE TABLE IF NOT EXISTS creatures (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_anonymous BOOLEAN NOT NULL DEFAULT TRUE,
    session_token UUID,
    email TEXT,
    password TEXT,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verification_token UUID,
    email_verification_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- XXX: add index for email lookup?

CREATE TABLE IF NOT EXISTS lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    share_code UUID UNIQUE DEFAULT gen_random_uuid(),
    world TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    world TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS lists_users (
    list_id UUID NOT NULL REFERENCES lists(id),
    user_id UUID NOT NULL REFERENCES users(id),
    character_id UUID NOT NULL REFERENCES characters(id),
    PRIMARY KEY (list_id, user_id, character_id)
);

CREATE TYPE soulcore_status AS ENUM ('obtained', 'unlocked');

CREATE TABLE IF NOT EXISTS lists_soulcores (
    list_id UUID NOT NULL REFERENCES lists(id),
    creature_id UUID NOT NULL REFERENCES creatures(id),
    status soulcore_status NOT NULL,
    PRIMARY KEY (list_id, soulcore_id)
);

CREATE TABLE IF NOT EXISTS characters_soulcores (
    character_id UUID NOT NULL REFERENCES characters(id),
    creature_id UUID NOT NULL REFERENCES creatures(id),
    PRIMARY KEY (character_id, soulcore_id)
);