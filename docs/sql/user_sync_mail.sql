CREATE TABLE dim.user_sync_mail
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
