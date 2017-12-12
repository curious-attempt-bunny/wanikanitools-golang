CREATE TABLE leech_train (
 train_id        SERIAL PRIMARY KEY,
 api_key         VARCHAR (50) NOT NULL,
 key             VARCHAR (20) NOT NULL,
 worst_incorrect INTEGER NOT NULL
);
