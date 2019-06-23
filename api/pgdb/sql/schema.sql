CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS secrets
(
        id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
        hash CHARACTER VARYING NOT NULL,
        secret_text CHARACTER VARYING NOT NULL,
        created_at CHARACTER VARYING NOT NULL,
        expires_at CHARACTER VARYING NOT NULL,
        remaining_views INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS hash_ind ON secrets (hash);

ALTER TABLE secrets OWNER TO secrets;
