CREATE TABLE IF NOT EXISTS dim.conversations
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 10 ),
    user_id bigint NOT NULL DEFAULT 0,
    conversation_id text COLLATE pg_catalog."default" NOT NULL DEFAULT ''::text,
    last_msg_id bigint NOT NULL DEFAULT 0,
    start_msg_id bigint NOT NULL,
    preview text COLLATE pg_catalog."default" NOT NULL DEFAULT ''::text,
    unread integer NOT NULL DEFAULT 0,
    pinnd_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    target_id bigint NOT NULL DEFAULT 0
);