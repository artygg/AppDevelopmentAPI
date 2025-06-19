--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5 (Debian 17.5-1.pgdg120+1)
-- Dumped by pg_dump version 17.5 (Debian 17.5-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: capture_attempts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.capture_attempts (
    id integer NOT NULL,
    place_id integer NOT NULL,
    user_name text NOT NULL,
    correct boolean DEFAULT false NOT NULL,
    time_ms integer NOT NULL,
    finished_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.capture_attempts OWNER TO postgres;

--
-- Name: capture_attempts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.capture_attempts ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.capture_attempts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: category_icons; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.category_icons (
    category_id integer NOT NULL,
    icon_name character varying(128) NOT NULL
);


ALTER TABLE public.category_icons OWNER TO postgres;

--
-- Name: image_place; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.image_place (
    id integer NOT NULL,
    image_location text,
    place_id integer,
    last_time_updated timestamp without time zone
);


ALTER TABLE public.image_place OWNER TO postgres;

--
-- Name: image_place_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.image_place ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.image_place_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: mines; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.mines (
    place_id integer NOT NULL,
    qid integer NOT NULL,
    expires_at timestamp without time zone NOT NULL
);


ALTER TABLE public.mines OWNER TO postgres;

--
-- Name: place_scores; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.place_scores (
    place_id integer NOT NULL,
    best_correct smallint DEFAULT 0 NOT NULL,
    best_time_ms integer DEFAULT 2147483647 NOT NULL,
    holder text,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.place_scores OWNER TO postgres;

--
-- Name: places; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.places (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    category_id integer NOT NULL,
    captured boolean DEFAULT false NOT NULL,
    user_captured text
);


ALTER TABLE public.places OWNER TO postgres;

--
-- Name: places_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.places_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.places_id_seq OWNER TO postgres;

--
-- Name: places_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.places_id_seq OWNED BY public.places.id;


--
-- Name: quizzes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.quizzes (
    id integer NOT NULL,
    place_id integer,
    quiz_json jsonb NOT NULL
);


ALTER TABLE public.quizzes OWNER TO postgres;

--
-- Name: quizzes_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.quizzes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.quizzes_id_seq OWNER TO postgres;

--
-- Name: quizzes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.quizzes_id_seq OWNED BY public.quizzes.id;


--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.refresh_tokens (
    id character varying(36) NOT NULL,
    user_id integer,
    token_hash text NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    is_revoked boolean DEFAULT false
);


ALTER TABLE public.refresh_tokens OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username character varying(30) NOT NULL,
    email character varying(254) NOT NULL,
    password_hash text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    is_active boolean DEFAULT true,
    last_login_at timestamp without time zone,
    failed_logins integer DEFAULT 0,
    locked_until timestamp without time zone,
    map_image_url text DEFAULT ''::text
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: places id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.places ALTER COLUMN id SET DEFAULT nextval('public.places_id_seq'::regclass);


--
-- Name: quizzes id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quizzes ALTER COLUMN id SET DEFAULT nextval('public.quizzes_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: capture_attempts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.capture_attempts (id, place_id, user_name, correct, time_ms, finished_at) FROM stdin;
\.


--
-- Data for Name: category_icons; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.category_icons (category_id, icon_name) FROM stdin;
10000	mappin.circle.fill
10001	ferris.wheel
10002	drop.fill
10003	gamecontroller.fill
10004	photo.artframe
10005	circle.grid.cross
10006	bowlingball.fill
10007	tent.2
10008	die.face.5.fill
10009	tortoise.fill
10010	face.smiling.fill
10011	figure.golf
10012	figure.dance
10013	figure.dance
10014	flag.2.crossed.fill
10015	lock.open.fill
10016	sparkles
10017	tent.2
10018	desktopcomputer
10019	steeringwheel
10020	desktopcomputer
10021	music.mic
10022	dot.squareshape.split.2x2
10023	flag.2.crossed.fill
10024	film.fill
10025	film.fill
10026	film.fill
10027	building.columns.fill
10028	building.columns.fill
10029	building.columns.fill
10030	building.columns.fill
10031	building.columns.fill
10032	music.note.house.fill
10033	mappin.circle.fill
10034	mappin.circle.fill
10035	mappin.circle.fill
10036	theatermasks.fill
10037	music.note.list
10038	theatermasks.fill
10039	music.note
10040	music.note
10041	music.note
10042	music.quarternote.3
10043	theatermasks.fill
10044	sparkles
10045	circle.hexagongrid.fill
10046	mappin.circle.fill
10047	paintbrush.pointed.fill
10048	figure.2.arms.open
10049	mappin.circle.fill
10050	graduationcap.fill
10051	sportscourt
10052	figure.strengthtraining.traditional
10053	ticket.fill
10054	gamecontroller.fill
10055	water.waves
10056	pawprint.circle.fill
10057	flag.2.crossed.fill
10058	ferris.wheel
10059	paintbrush.pointed.fill
10060	sportscourt
10061	sportscourt
10062	sportscourt
10063	sportscourt
10064	sportscourt
10065	sportscourt
10066	sportscourt
10067	sportscourt
10068	sparkles
10069	paintbrush.pointed.fill
11041	bus.fill
12000	building.columns.fill
12009	building.columns.fill
12010	building.columns.fill
12011	graduationcap.fill
12012	tortoise.fill
12013	graduationcap
12014	graduationcap
12015	graduationcap
12016	graduationcap
12017	graduationcap
12018	graduationcap
12019	graduationcap
12020	graduationcap
12021	graduationcap
12022	graduationcap
12023	graduationcap
12024	graduationcap
12025	graduationcap
12026	graduationcap
12027	graduationcap
12028	graduationcap
12029	graduationcap
12030	graduationcap
12031	graduationcap
12032	graduationcap
12033	graduationcap
12034	graduationcap
12035	graduationcap
12036	graduationcap
12037	graduationcap
12038	sportscourt
12039	graduationcap
12040	graduationcap
12041	theatermasks.fill
12042	graduationcap
12043	graduationcap
12044	graduationcap.fill
12045	graduationcap.fill
12046	graduationcap
12047	graduationcap
12048	graduationcap
12049	graduationcap.fill
12050	graduationcap.fill
12051	graduationcap.fill
12052	graduationcap.fill
12053	graduationcap.fill
12054	graduationcap.fill
12055	graduationcap.fill
12056	graduationcap.fill
12057	graduationcap.fill
12058	graduationcap.fill
12059	graduationcap.fill
12060	graduationcap.fill
12061	graduationcap.fill
12062	graduationcap.fill
12063	graduationcap.fill
12064	building.columns.fill
12065	building.columns
12066	building.columns
12067	building.columns
12068	building.2.fill
12069	building.columns.fill
12070	mug.fill
12071	mug.fill
12072	mug.fill
12089	person.2.wave.2
13000	mappin.circle.fill
13001	mappin.circle.fill
13002	birthday.cake.fill
13003	wineglass.fill
13004	wineglass.fill
13005	wineglass.fill
13006	wineglass.fill
13007	wineglass.fill
13008	wineglass.fill
13009	wineglass.fill
13010	wineglass.fill
13011	wineglass.fill
13012	wineglass.fill
13013	wineglass.fill
13014	wineglass.fill
13015	music.mic
13016	wineglass.fill
13017	wineglass.fill
13018	wineglass.fill
13019	wineglass.fill
13020	wineglass.fill
13021	wineglass.fill
13022	wineglass.fill
13023	wineglass.fill
13024	wineglass.fill
13025	wineglass.fill
13026	fork.knife.circle.fill
13027	fork.knife.circle.fill
13028	mappin.circle.fill
13029	mappin.circle.fill
13030	fork.knife.circle.fill
13031	fork.knife.circle.fill
13032	mappin.circle.fill
13037	mappin.circle.fill
13038	mappin.circle.fill
13039	fork.knife.circle.fill
13040	birthday.cake.fill
13041	mappin.circle.fill
13051	fork.knife.circle.fill
13068	fork.knife.circle.fill
13076	fork.knife.circle.fill
13077	fork.knife.circle.fill
13078	fork.knife.circle.fill
13079	fork.knife.circle.fill
13080	fork.knife.circle.fill
13081	fork.knife.circle.fill
13082	fork.knife.circle.fill
13083	fork.knife.circle.fill
13084	fork.knife.circle.fill
13085	fork.knife.circle.fill
13086	fork.knife.circle.fill
13087	fork.knife.circle.fill
13088	fork.knife.circle.fill
13089	fork.knife.circle.fill
13090	fork.knife.circle.fill
13091	fork.knife.circle.fill
13092	fork.knife.circle.fill
13093	fork.knife.circle.fill
13094	fork.knife.circle.fill
13095	fork.knife.circle.fill
13096	fork.knife.circle.fill
13097	fork.knife.circle.fill
13098	fork.knife.circle.fill
14000	calendar
14001	person.3.sequence.fill
14002	calendar
14003	calendar
14004	party.popper.fill
14005	party.popper.fill
14006	calendar
14007	calendar
14008	calendar
14009	cart.fill
14010	cart.fill
14011	cart.fill
14012	tent.2
14013	cart.fill
14014	tent.2
14015	calendar
14016	party.popper.fill
15000	mappin.circle.fill
15001	cross.case.fill
15002	mappin.circle.fill
15003	cross.case.fill
15004	person.2.wave.2
15005	banknote.fill
15006	figure.strengthtraining.traditional
15007	cross.case.fill
15008	mappin.circle.fill
15009	cross.case.fill
15010	cross.case.fill
15011	cross.case.fill
15012	house.fill
15013	cross.case.fill
15014	cross.case.fill
15015	cross.case.fill
15016	cross.case.fill
15017	cross.case.fill
15018	brain.head.profile
15019	brain.head.profile
15020	brain.head.profile
15021	cross.case.fill
16000	leaf.circle
16001	drop.fill
16002	water.waves
16003	beach.umbrella.fill
16004	leaf.circle
16005	leaf.arrow.circlepath
16006	leaf.circle
16007	cube.fill
16008	leaf.circle
16009	drop.triangle
16010	drop.triangle
16011	castle.fill
16012	leaf.circle
16013	leaf.circle
16014	leaf.circle
16015	leaf.fill
16016	drop.fill
16017	leaf.arrow.circlepath
16018	sailboat.fill
16019	figure.hiking
16020	leaf.circle
16021	drop.fill
16022	leaf.fill
16023	water.waves
16024	lightbulb.max.fill
16025	mappin.and.ellipse
16026	mappin.and.ellipse
16027	mountain.2.fill
16028	leaf.circle
16029	beach.umbrella.fill
16030	leaf.circle
16031	crown.fill
16032	leaf.circle.fill
16033	leaf.circle.fill
16034	leaf.circle.fill
16035	leaf.circle.fill
16036	leaf.circle.fill
16037	leaf.circle.fill
16038	leaf.circle.fill
16039	leaf.circle.fill
16040	building.2
16041	building.2
16042	water.waves
16043	water.waves
16044	figure.climbing
16045	sun.max.fill
16046	binoculars.fill
16047	paintbrush.fill
16048	pawprint
16049	surfboard
16050	leaf.circle
16051	flame.fill
16052	water.waves
16053	leaf.circle
16054	wind
16055	leaf.circle
16056	drop.fill
16057	leaf
16058	mountain.2
16059	mountain.2.fill
16060	basket.fill
16061	map
16062	building.2.fill
16063	globe.europe.africa.fill
16064	map
16065	building.2
16066	map
16067	map
16068	building.2.crop.circle
16069	tree.fill
16070	drop.fill
17002	shippingbox
17003	paintbrush
17004	hammer.circle
17022	book.closed.fill
18000	sportscourt
18001	sportscourt
18002	sportscourt
18003	sportscourt
18004	sportscourt
18005	sportscourt
18006	sportscourt
18007	sportscourt
18008	sportscourt
18009	bowlingball.fill
18010	sportscourt
18011	sportscourt
18012	sportscourt
18013	sportscourt
18014	sportscourt
18015	sportscourt
18016	sportscourt
18017	sportscourt
18018	sportscourt
18019	sportscourt
18020	sportscourt
18021	sportscourt
18022	sportscourt
18023	sportscourt
18024	sportscourt
18025	figure.dance
18026	sportscourt
18027	sportscourt
18028	sportscourt
18029	sportscourt
18030	sportscourt
18031	sportscourt
18032	sportscourt
18033	sportscourt
18034	sportscourt
18035	sportscourt
18036	sportscourt
18037	sportscourt
18038	tram.fill
18039	sportscourt
18040	sportscourt
18041	sportscourt
18042	sportscourt
18043	sportscourt
18044	sportscourt
18045	sportscourt
18046	sportscourt
18047	sportscourt
18048	sportscourt
18049	sportscourt
18050	sportscourt
18051	sportscourt
18052	sportscourt
18053	sportscourt
18054	sportscourt
18055	leaf.circle.fill
18056	sportscourt
18057	sportscourt
18058	sportscourt
18059	sportscourt
18060	sportscourt
18061	sportscourt
18062	sportscourt
18063	sportscourt
18064	sportscourt
18065	sportscourt
18066	sportscourt
18067	sportscourt
18068	sportscourt
18069	sportscourt
18070	sportscourt
18071	sportscourt
18072	sportscourt
18073	sportscourt
18074	sportscourt
18075	sportscourt
18076	graduationcap.fill
18077	sportscourt
18078	sportscourt
18079	sportscourt
18080	sportscourt
18081	sportscourt
18082	sportscourt
18083	sportscourt
18084	sportscourt
18085	sportscourt
18086	sportscourt
19000	mappin.circle.fill
19031	airplane
19032	airplane
19033	airplane
19034	airplane
19035	airplane
19036	airplane
19037	airplane
19038	airplane
19039	airplane
19040	airplane
19041	airplane
19050	tram.fill
\.


--
-- Data for Name: image_place; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.image_place (id, image_location, place_id, last_time_updated) FROM stdin;
4	images/20250616_154848_photo_5327829827891361653_x.jpg	2	2025-06-16 15:48:48.866107
5	images/20250616_154901_photo_5327829827891361653_x.jpg	4	2025-06-16 15:49:01.719316
1	images/20250619_173551_Simulator Screenshot - iPhone 16 Pro - 2025-06-06 at 01.05.59.png	1	2025-06-19 17:35:51.804243
\.


--
-- Data for Name: mines; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.mines (place_id, qid, expires_at) FROM stdin;
\.


--
-- Data for Name: place_scores; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.place_scores (place_id, best_correct, best_time_ms, holder, updated_at) FROM stdin;
\.


--
-- Data for Name: places; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.places (id, name, latitude, longitude, category_id, captured, user_captured) FROM stdin;
2	Flamingo's Plaza Wokrestaurant	52.782527	6.8971	10008	f	\N
3	Travellers Taste "Wildlands Adventure Zoo Emmen"	52.782307	6.890582	10056	f	\N
4	Tweestryd	52.778517	6.887992	10001	f	\N
6	Utopolis Emmen	52.788117	6.888345	10024	f	\N
7	Emmen Raadhuisplein	52.782889	6.892564	16041	f	\N
8	Jungola	52.782509	6.887649	10056	f	\N
9	Aqua Mundo	52.67391	6.77455	18075	f	\N
10	Mommeriete Bierbrouwerij	52.611335	6.67826	10037	f	\N
11	Nationaal Monument Westerbork	52.916879	6.611729	16020	f	\N
12	Herinneringscentrum Kamp Westerbork	52.921008	6.569726	10030	f	\N
13	Aqua Mexicana	52.623961	6.561419	18075	f	\N
14	Gold Rush	52.623899	6.561734	10001	f	\N
15	Drents Museum	52.993341	6.56413	10030	f	\N
16	Nationaal Park Dwingelderveld	52.783369	6.373425	16034	f	\N
17	Beerze Bulten	52.511944	6.546296	16008	f	\N
18	Zwembad Tropiqua	53.105562	6.867629	18075	f	\N
19	Pier 99	52.435403	7.081435	13009	f	\N
20	Tierpark Nordhorn	52.427683	7.092108	10056	f	\N
21	Cafe Extrablatt Nordhorn BetriebsGmbH	52.434095	7.069768	13027	f	\N
22	Nationaal Park Drents-Friese Wold	52.927235	6.302528	16034	f	\N
23	Maallust Bierbrouwerij	53.033238	6.387525	13029	f	\N
24	Recreatieplas De Zwarte Dennen	52.624082	6.273048	16003	f	\N
25	Kino Papenburg	53.076176	7.404349	10024	f	\N
26	Schouwburg Ogterop	52.692608	6.190523	10043	f	\N
5	Rimbula River	52.781763	6.885543	10056	t	player1
27	Test	52.78084378872667	6.910397750597617	1	f	\N
28	Place 2	52.78	6.9	1	f	\N
29	The kingdom of kingdoms	52.78	6.9	2	f	\N
30	Test	52.78	6.9	9	f	\N
31	123	52.78	6.9	1	f	\N
32	Roundabout	52.77443236693949	6.9266169173208025	7	f	\N
33	KIGDOM	52.78164658722868	6.938662737451782	2	f	\N
34	Сральник	52.779696089101655	6.922779386251341	6	f	\N
35	Test 	52.77269752118785	6.900077880431531	1	f	\N
36	Test Place	52	5	1	f	\N
1	Groothuis	52.789762	6.897665	13027	t	testuser1
\.


--
-- Data for Name: quizzes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.quizzes (id, place_id, quiz_json) FROM stdin;
1	5	{"place_id": 5, "questions": [{"text": "Where is Rimbula River located?", "answer": 2, "options": ["Australia", "Brazil", "Netherlands", "Canada"]}, {"text": "What type of animals are commonly found around Rimbula River?", "answer": 1, "options": ["Lions", "Crocodiles", "Penguins", "Elephants"]}, {"text": "What is Rimbula River known for?", "answer": 1, "options": ["Historic landmarks", "Rich biodiversity", "Modern architecture", "Culinary delights"]}, {"text": "Which of the following is a common Dutch dish?", "answer": 2, "options": ["Sushi", "Tacos", "Stroopwafel", "Pasta"]}, {"text": "What is the capital city of the Netherlands?", "answer": 0, "options": ["Amsterdam", "Berlin", "Paris", "London"]}, {"text": "Which famous Dutch painter is known for his sunflower paintings?", "answer": 0, "options": ["Vincent van Gogh", "Pablo Picasso", "Leonardo da Vinci", "Claude Monet"]}, {"text": "What is the currency of the Netherlands?", "answer": 1, "options": ["Pound", "Euro", "Dollar", "Yen"]}]}
\.


--
-- Data for Name: refresh_tokens; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.refresh_tokens (id, user_id, token_hash, expires_at, created_at, is_revoked) FROM stdin;
f5525b2a-7591-4b2d-bdb4-829220ee24c6	1	$2a$12$bQa5.UuN3lEcPj7nS5Nh/OBj7xv8.jcHhWFckMEjxVYX.0vwtrz/6	2025-06-26 17:27:03.878624	2025-06-19 15:27:03.879789	f
35394302-cbdb-4954-afa6-862d7b81c448	4	$2a$12$l9ZEU/5LwgpvAOaEckYIrOAIO8dxPLpuINWf8JnwpImZ3lrH7.wWm	2025-06-26 17:29:07.330679	2025-06-19 15:29:07.331464	f
cf139790-5115-4861-8352-ff6dedf41a65	5	$2a$12$KQUb1wPqrJkLGrkJv3.RX.DO.g3fBiji3ybmAWnsY8UIbve2RS5cO	2025-06-26 17:54:25.19723	2025-06-19 15:54:25.197592	f
2b78c0e4-e613-4351-b2ef-273dcc35d31a	6	$2a$12$32Y869plh.hpqdPohwMckuT1NR97RpCR/z4quvoQZPN6pW0yiL2uu	2025-06-26 17:55:45.729626	2025-06-19 15:55:45.730771	f
1299ce9f-62d8-4973-8c65-ec8239743a59	7	$2a$12$NXPtWXyGrtWwwjCmmvR2z.0/e7rhijK7bXV6qvq7bgt3ewLGNcTma	2025-06-26 18:00:43.123312	2025-06-19 16:00:43.124209	f
c01b5aa3-21b5-471f-bbb9-0d0e26541234	7	$2a$12$0SOvja4z9xWgYaPFroBTs.BcR3l9zMf0IdQ6bc9GPFxqEwx7LjvtS	2025-06-26 18:06:56.648955	2025-06-19 16:06:56.650675	f
a0b2f927-e494-4359-b757-810259cc531d	8	$2a$12$WWIxy9OURzBfdlXz722nGuyYB3EGbr431JisxLRYYapzGUN6sfMHC	2025-06-26 18:10:29.429734	2025-06-19 16:10:29.430049	f
9209850a-a36d-495e-ada4-b1e63fb5f038	8	$2a$12$72bZdAP7S5QeQ0uJhfW3ruxMISseh/./aHN.jMm5xFRmyG0DvVvJq	2025-06-26 18:10:34.066845	2025-06-19 16:10:34.067196	f
b77e5fbb-448d-4146-a962-0251600c6ccc	8	$2a$12$vqAZo20XurLu5VszKlv6oeCiWf43S2hLUJ5dZKulzgIxkJkM6xQDa	2025-06-26 18:13:08.186863	2025-06-19 16:13:08.189159	f
ecaf6a4d-cc88-4f01-8a63-625a5027a66f	8	$2a$12$inDaDeV6AMsSxR9oGPd6c.WtK4QXkrKjnHAyz2qB3BLoepXjOL.Ie	2025-06-26 18:14:05.34263	2025-06-19 16:14:05.343323	f
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, username, email, password_hash, created_at, updated_at, is_active, last_login_at, failed_logins, locked_until, map_image_url) FROM stdin;
1	testuser	test@example.com	$2a$12$qRWr3nKnvWwTCC/2IddAvuWoXpatW3zeC56/QkySAqcE1iGwqR4KK	2025-06-19 15:27:03.686302	2025-06-19 15:27:03.686302	t	\N	0	\N	
4	testuser1	test@example1.com	$2a$12$.3S2L28eqo/VuwnUwuXo6Oiq1fIvBvWI4y/ugbIA0q4jhju6ntUse	2025-06-19 15:29:07.141418	2025-06-19 15:29:07.141418	t	\N	0	\N	
5	testuser11	test@example11.com	$2a$12$pZLaWcYxfomQmNaY1k.ET.hiCKflhITkRcTagT0CeQDty5aOU/WxW	2025-06-19 15:54:25.002532	2025-06-19 15:54:25.002532	t	\N	0	\N	
6	testuser111	test@example111.com	$2a$12$c.3VRhgC8VdA0/O.88Ye2e1oOVbdTijIl0KqKk3NS5l3zSDSKymXi	2025-06-19 15:55:45.536723	2025-06-19 15:55:45.536723	t	\N	0	\N	
7	test	test@gmail.com	$2a$12$cUFybJ8R19wpBWpXLF5w1.1pYZk7ysy277LhsKxMgLI9.TTBPYoQ2	2025-06-19 16:00:42.937581	2025-06-19 16:06:56.464212	t	2025-06-19 16:06:56.464212	0	\N	
8	test2	test@gmail2.com	$2a$12$29gYRkZ7uIimYGe.lNzbgufLjK7OzUK4KsOLNofKwNLeaMFhccCnK	2025-06-19 16:10:29.242142	2025-06-19 16:14:05.155109	t	2025-06-19 16:14:05.155109	0	\N	
\.


--
-- Name: capture_attempts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.capture_attempts_id_seq', 1, false);


--
-- Name: image_place_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.image_place_id_seq', 18, true);


--
-- Name: places_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.places_id_seq', 36, true);


--
-- Name: quizzes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.quizzes_id_seq', 1, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 8, true);


--
-- Name: category_icons category_icons_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.category_icons
    ADD CONSTRAINT category_icons_pkey PRIMARY KEY (category_id);


--
-- Name: image_place image_place_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_place
    ADD CONSTRAINT image_place_pkey PRIMARY KEY (id);


--
-- Name: image_place image_place_place_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_place
    ADD CONSTRAINT image_place_place_id_key UNIQUE (place_id);


--
-- Name: places places_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.places
    ADD CONSTRAINT places_pkey PRIMARY KEY (id);


--
-- Name: quizzes quizzes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quizzes
    ADD CONSTRAINT quizzes_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: quizzes unique_place_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quizzes
    ADD CONSTRAINT unique_place_id UNIQUE (place_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_refresh_tokens_expires_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refresh_tokens_expires_at ON public.refresh_tokens USING btree (expires_at);


--
-- Name: idx_refresh_tokens_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens USING btree (user_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_username ON public.users USING btree (username);


--
-- Name: image_place image_place_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_place
    ADD CONSTRAINT image_place_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id);


--
-- Name: quizzes quizzes_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quizzes
    ADD CONSTRAINT quizzes_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id);


--
-- Name: refresh_tokens refresh_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

