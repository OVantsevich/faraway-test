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
- client - client side app
- server - implementation of the "protocol" on the example of the simplest server application



