CREATE TABLE user_sync_mail
(
    syn_seq bigint NOT NULL,
    user_id bigint NOT NULL DEFAULT 0,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    send_id bigint NOT NULL DEFAULT 0,
    conversation_id text NOT NULL DEFAULT '',
    sync_type integer NOT NULL DEFAULT 0,
    msg_id bigint NOT NULL DEFAULT 0,
    content text NOT NULL DEFAULT ''
) WITH (
 	tsdb.hypertable,
   	tsdb.partition_column='created_at',
   	tsdb.chunk_interval = '1 month'
);

CREATE UNIQUE INDEX IF NOT EXISTS user_sync_mail_id_created_at_idx
    ON user_sync_mail USING btree
    (user_id, syn_seq DESC NULLS LAST, created_at DESC NULLS LAST)
    WITH (fillfactor=100, deduplicate_items=True)
    TABLESPACE pg_default;
