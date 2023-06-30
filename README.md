<a name="readme-top"></a>
### Requirements
Go installed for running client, "taskfile" and test's.

Docker installed (to run docker-compose)


### Installation

1. Install <a href="https://taskfile.dev/">Taskfile</a>, alternative for makefile
   ```sh
   go install github.com/go-task/task/v3/cmd/task@latest
   ```
2. Start server
   ```sh
   task server
   ```
3. Start client
   ```sh
   task client
   ```





## Project structure

- protocol - protocol - protocol describing tcp communication and PoW implementation
- client - client side console app
- server - implementation of the "protocol" on the example of the simplest server docker application


## Environment variables

### Client

| name           | type    | default        | description
|----------------|---------|----------------|--------------------------------------
| SERVER_HOST    | string  | localhost | Server host
| SERVER_PORT  | string     | 12345              | Server tcp port

### Server

| name             | type    | default        | description
|------------------|---------|----------------|----------------------------------------
| SERVICE_NAME      | string  | Word of Wisdom   | Service name
| SERVICE_HOST       | string    | 0.0.0.0             | Service host
| SERVICE_PORT    | string     | 12345             | Service tcp port
| ENVIRONMENT | string(PROD/DEV)     | PROD             | Service environment stage. May be DEV or PROD. Affects the level of logging 
| TARGET_BITS | uint8     | 0             | The complexity of the PoW algorithm. The first N bits of the hash must be 0. The default value of 0 means that PoW is disabled.
| READ_TIMEOUT | int64     | 60000             | The maximum time required for a client to resolve and send a Challenge Response protocol response. Calculated in milliseconds
| DB_DIR | string     | db             | SQLite database directory inside the container. DO NOT CHANGE
| DB_NAME | string     | database             | database filename (if any)
| SQLITE_MODE | string(memory, ro, rw, rwc)     | rwc             | SqliteMode - Access Mode of the database. rwc - The database is opened for reading and writing



