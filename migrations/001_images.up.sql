DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'image_status') THEN
        CREATE TYPE image_status AS ENUM ('pending', 'processed');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS images
(
    id UUID PRIMARY KEY,
    original_key  VARCHAR(255) NOT NULL,
    processed_key VARCHAR(255),
    original_name VARCHAR(255) NOT NULL,
    content_type  VARCHAR(100) NOT NULL,
    size          BIGINT NOT NULL,
    status        image_status NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at  TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_image_status 
    ON images(status);
    
CREATE INDEX IF NOT EXISTS idx_created_at_status 
    ON images(created_at DESC);