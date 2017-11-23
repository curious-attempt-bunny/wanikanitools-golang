CREATE TABLE scripts (
 script_id      SERIAL PRIMARY KEY,
 api_key        INTEGER NOT NULL,
 browser_uuid   VARCHAR (50) NOT NULL,
 script_url     VARCHAR (256) NOT NULL,
 script_version VARCHAR (50) NOT NULL,
 last_seen      INTEGER NOT NULL
);
