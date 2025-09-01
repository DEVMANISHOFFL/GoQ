CREATE TABLE jobs (
    id                      UUID PRIMARY KEY,
    type                    TEXT NOT NULL,
    payload                 JSONB NOT NULL,
    status                  TEXT NOT NULL,
    created_at              TIMESTAMPS WITH TIME ZONE DEFAULT NOW(),
    updated_at              TIMESTAMPS WITH TIME ZONE DEFAULT NOW()
);


