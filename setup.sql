CREATE TABLE IF NOT EXISTS chats (
  id          SERIAL PRIMARY KEY,
  chatid      BIGINT  UNIQUE NOT NULL,
  mode        INTEGER DEFAULT 0,
  studymodes  INTEGER DEFAULT 3
);

CREATE TABLE IF NOT EXISTS phrases (
  id          SERIAL PRIMARY KEY,
  chatid      BIGINT,
  phrase      TEXT,
  explanation TEXT
);

CREATE TABLE IF NOT EXISTS studies (
  id        SERIAL PRIMARY KEY,
  phraseid  INTEGER REFERENCES phrases NOT NULL,
  score     INTEGER DEFAULT 0,
  studymode INTEGER NOT NULL,
  timestamp timestamp DEFAULT CURRENT_TIMESTAMP
);

