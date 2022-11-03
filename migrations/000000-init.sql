-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS users (
  uuid TEXT primary key,
  email TEXT,
  username TEXT,
  picture TEXT,
  authid TEXT,
  salt TEXT,
  UNIQUE(email)
);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE users;
