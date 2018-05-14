CREATE TABLE IF NOT EXISTS article (
	id           SERIAL PRIMARY KEY,
	title	     text,
	lang     	 text,
	pubtime      timestamp with time zone,
	updatetime   timestamp with time zone,
	fetchtime    timestamp with time zone,
	thumburl 	 text,
	html  		 text,
	author       text,
	url     	 text UNIQUE
);

CREATE TABLE IF NOT EXISTS tag (
	id   SERIAL PRIMARY KEY,
	name text UNIQUE
);

CREATE TABLE IF NOT EXISTS map_tag_article (
	tagid     int,
	articleid int
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_map_tag_article ON map_tag_article (tagid, articleid);

CREATE TABLE IF NOT EXISTS line_xml (
	id   SERIAL PRIMARY KEY,
	time timestamp with time zone
);
