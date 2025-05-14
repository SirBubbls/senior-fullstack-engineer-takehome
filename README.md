# Requirements
Run `docker compose up -d` to start the postgresql database server.

# Go Implementation
## Starting
Use `go build main.go && ./main` to build and run the application. 
The database server should be running at this point.
Database migrations will be executed automatically.

## Testing
The database is required to run and migrations should be already applied once before running the tests.
Run `go test` to run integration test suite.

> Warning: Database gets wiped during testing


# Rust Implementation
[Install cargo (& rust toolchain)](https://www.rust-lang.org/tools/install). 
`libpq` as a driver for the postgres database client may also be required.
The Rust Implementation doesn't run any migrations, instead it requires the same schema as the Go implementation. 
So run the Go migrations first.
Run the application with `cargo run` and tests with `cargo test`.
Both APIs are compatible.

Configure the database connection via:
```
export DATABASE_URL=postgres://postgres:postgres@localhost/postgres
```

# Emmit Script
The emmit script uses `uv` as its project management tool.
Install `uv` and execute the script with `uv run emmit.py`.
A web server should be running, otherwise the script won't work.

# Endpoints
## POST `/submit`
Creates a new entry for the weather at a specific day.

> If there is an entry for the specific day already, an update will be performed.

Example Payload:
```json
{
    "date": "2022-01-01",
    "temperature": 32.2,
    "humidity": 2.1
}
```

## GET `/day?day=<day>`
Returns the data for a specific day, if there is no data at specified date, the request will fail.

`/day?day=2022-01-01`  

Returns:
```json
{
    "date": "2022-01-01T00:00:00Z",
    "temperature": 32.2,
    "humidity": 2.1
}
```


## GET `/range?start=<start>&end=<end>`
Returns a range of days between `start` and `end` including `start` and `end` dates. 
If there is no data for the specified range, an empty array will be returned. 
If there are gaps in the dataset, the endpoint will return every datapoint available and skip said gaps in the response.

`/range?start=2022-01-01&end=2022-01-04`  

Returns:
```json
[
    {
        "date": "2022-01-01T00:00:00Z",
        "temperature": 32.2,
        "humidity": 2.1
    },
    {
        "date": "2022-01-03T00:00:00Z",
        "temperature": 12.2,
        "humidity": 23.1
    }
]
```

## WS `/updates`
Connects to a websocket connection which will notify clients about new measurements submitted via the `/submit` endpoint.

Example Payload:
```json
{"date":"2022-01-09T00:00:00Z","temperature":12.2,"humidity":23.1}
```
