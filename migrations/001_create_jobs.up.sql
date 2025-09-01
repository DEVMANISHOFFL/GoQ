CREATE TABLE jobs (
    id                      UUID PRIMARY KEY,
    type                    TEXT NOT NULL,
    payload                 JSONB NOT NULL,
    status                  TEXT NOT NULL,
    created_at              TIMESTAMPS WITH TIME ZONE DEFAULT NOW(),
    updated_at              TIMESTAMPS WITH TIME ZONE DEFAULT NOW()
);


ALTER TABLE jobs ADD COLUMN retries INT DEFAULT 0;
ALTER TABLE jobs ADD COLUMN max_retries INT DEFAULT 3;
ALTER TABLE jobs ADD COLUMN error_message TEXT;
