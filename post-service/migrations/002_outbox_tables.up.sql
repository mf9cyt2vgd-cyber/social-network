CREATE TABLE IF NOT EXISTS posts_outbox (
    "offset" BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payload JSONB,
    metadata JSONB,
    transaction_id xid8 NOT NULL
);

--table for watermill services
CREATE TABLE IF NOT EXISTS watermill_offsets_posts_outbox_forwarder (
    consumer_group TEXT NOT NULL,
    offset_acked BIGINT NOT NULL,
    last_processed_transaction_id xid8 NOT NULL,
    PRIMARY KEY (consumer_group)
);