CREATE TABLE measurements (
    time TIMESTAMP WITHOUT TIME ZONE DEFAULT now() NOT NULL PRIMARY KEY,
    temperature REAL NOT NULL,
    humidity REAL NOT NULL
);
