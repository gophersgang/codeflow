# Codeflow

Extendable deployment pipeline

# Local Development
Install [Docker Compose](https://docs.docker.com/compose/install/)

Copy `server/codeflow.yml` to `server/codeflow.dev.yml`

To start the server and all dependencies run:
```
$ make up
```
Check `Makefile` to see what's happening under the hood.

Local dashboard will be started on [http://localhost:3000](http://localhost:3000)

`dashboard` and `server` will reload on any file change :boom: Happy coding!!!


## Overriding configurations

### Dashboard port

For instance, if you want to run the dashboard on the `3010` port instead of `3000`, you need to have these files changed as:

- `docker-compose.override.yml`

```
version: "2"
services:
  dashboard:
    ports:
      - "3010:3000"
```

- `dashboard/.env.development`, change/add the environment variable as:

```
REACT_APP_ROOT_PORT=3010
```

- `server/configs/codeflow.dev.yml`, change the following key to:

```
plugins:
  codeflow:
    allowed_origins:
      - "http://localhost:3010"
```
