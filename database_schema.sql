-- Auth and core tables schema only

-- users table (from main.go)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(30) UNIQUE NOT NULL,
    email VARCHAR(254) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP WITHOUT TIME ZONE,
    failed_logins INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITHOUT TIME ZONE,
    map_image_url TEXT
);
ALTER TABLE users
    ADD COLUMN map_image_url TEXT DEFAULT '';


-- refresh_tokens table (from main.go)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    is_revoked BOOLEAN DEFAULT FALSE
);

CREATE TABLE public.capture_attempts (
    id integer NOT NULL,
    place_id integer NOT NULL,
    user_name text NOT NULL,
    correct boolean DEFAULT false NOT NULL,
    time_ms integer NOT NULL,
    finished_at timestamp without time zone DEFAULT now() NOT NULL
);
ALTER TABLE public.capture_attempts ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.capture_attempts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

CREATE TABLE public.category_icons (
    category_id integer NOT NULL,
    icon_name character varying(128) NOT NULL
);

CREATE TABLE public.image_place (
    id integer NOT NULL,
    image_location text,
    place_id integer,
    last_time_updated timestamp without time zone
);
ALTER TABLE public.image_place ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.image_place_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

CREATE TABLE public.mines (
    place_id integer NOT NULL,
    qid integer NOT NULL,
    expires_at timestamp without time zone NOT NULL
);

CREATE TABLE public.place_scores (
    place_id integer NOT NULL,
    best_correct smallint DEFAULT 0 NOT NULL,
    best_time_ms integer DEFAULT 2147483647 NOT NULL,
    holder text,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.places (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    category_id integer NOT NULL,
    captured boolean DEFAULT false NOT NULL,
    user_captured text
);
CREATE SEQUENCE public.places_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.places_id_seq OWNED BY public.places.id;

CREATE TABLE public.quizzes (
    id integer NOT NULL,
    place_id integer,
    quiz_json jsonb NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE SEQUENCE public.quizzes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.quizzes_id_seq OWNED BY public.quizzes.id; 