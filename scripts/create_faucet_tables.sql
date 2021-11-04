CREATE TABLE claim (
  addr VARCHAR(64),
  ip VARCHAR(15),
  last_claim bigint,
  claims smallint
);

CREATE TABLE donation (
  addr VARCHAR(64),
  amount decimal
);