--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4 (Debian 17.4-1.pgdg120+2)
-- Dumped by pg_dump version 17.4 (Debian 17.4-1.pgdg120+2)

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
-- Name: category_icons; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.category_icons (
    category_id integer NOT NULL,
    icon_name character varying(128) NOT NULL
);


ALTER TABLE public.category_icons OWNER TO postgres;

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
-- Name: places id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.places ALTER COLUMN id SET DEFAULT nextval('public.places_id_seq'::regclass);


--
-- Name: quizzes id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quizzes ALTER COLUMN id SET DEFAULT nextval('public.quizzes_id_seq'::regclass);


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
-- Data for Name: places; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.places (id, name, latitude, longitude, category_id, captured, user_captured) FROM stdin;
1	Groothuis	52.789762	6.897665	13027	f	\N
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
\.


--
-- Data for Name: quizzes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.quizzes (id, place_id, quiz_json) FROM stdin;
1	5	{"place_id": 5, "questions": [{"text": "Where is Rimbula River located?", "answer": 2, "options": ["Australia", "Brazil", "Netherlands", "Canada"]}, {"text": "What type of animals are commonly found around Rimbula River?", "answer": 1, "options": ["Lions", "Crocodiles", "Penguins", "Elephants"]}, {"text": "What is Rimbula River known for?", "answer": 1, "options": ["Historic landmarks", "Rich biodiversity", "Modern architecture", "Culinary delights"]}, {"text": "Which of the following is a common Dutch dish?", "answer": 2, "options": ["Sushi", "Tacos", "Stroopwafel", "Pasta"]}, {"text": "What is the capital city of the Netherlands?", "answer": 0, "options": ["Amsterdam", "Berlin", "Paris", "London"]}, {"text": "Which famous Dutch painter is known for his sunflower paintings?", "answer": 0, "options": ["Vincent van Gogh", "Pablo Picasso", "Leonardo da Vinci", "Claude Monet"]}, {"text": "What is the currency of the Netherlands?", "answer": 1, "options": ["Pound", "Euro", "Dollar", "Yen"]}]}
\.


--
-- Name: places_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.places_id_seq', 34, true);


--
-- Name: quizzes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.quizzes_id_seq', 1, true);


--
-- Name: category_icons category_icons_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.category_icons
    ADD CONSTRAINT category_icons_pkey PRIMARY KEY (category_id);


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
-- Name: quizzes unique_place_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quizzes
    ADD CONSTRAINT unique_place_id UNIQUE (place_id);


--
-- Name: quizzes quizzes_place_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quizzes
    ADD CONSTRAINT quizzes_place_id_fkey FOREIGN KEY (place_id) REFERENCES public.places(id);


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

