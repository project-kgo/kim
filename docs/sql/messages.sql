CREATE TABLE IF NOT EXISTS dim.messages
(
    id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    conversation_id text COLLATE pg_catalog."default" NOT NULL,
    sender_id bigint NOT NULL DEFAULT 0,
    receiver_id bigint NOT NULL DEFAULT 0,
    content jsonb NOT NULL,
    status integer NOT NULL DEFAULT 1,
	updated_at timestamp with time zone NOT NULL DEFAULT now()
) WITH (
 	tsdb.hypertable,
   	tsdb.partition_column='created_at',
   	tsdb.chunk_interval = '1 month'
)