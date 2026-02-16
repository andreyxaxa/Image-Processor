DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'outbox_status') THEN
        CREATE TYPE outbox_status AS ENUM ('pending', 'processing', 'processed', 'failed');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS images_outbox
(
    id              UUID PRIMARY KEY,
    aggregate_id    UUID NOT NULL,
    payload         JSONB NOT NULL,
    status          outbox_status NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMP,
    retry_count     INTEGER NOT NULL DEFAULT 0,

    CONSTRAINT fk_aggregate_image FOREIGN KEY (aggregate_id) REFERENCES images(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_outbox_status_created 
    ON images_outbox(status, created_at) 
    WHERE status IN ('pending');

CREATE INDEX IF NOT EXISTS idx_outbox_aggregate 
    ON images_outbox(aggregate_id);

CREATE INDEX IF NOT EXISTS idx_outbox_processed_at 
    ON images_outbox(processed_at) 
    WHERE status = 'processed';