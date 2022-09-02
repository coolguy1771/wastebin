# Wastebin

**Wastebin** is a self hosted web service that allows you to share pastes anonymously. Wastebin was designed to be stateless and uses the following tech stack

| Component |  Framework |
|-----------|------------|
| Backend   | Fiber      |
| Database  | PostgreSQL |
| Frontend  | Svelte     |

## Configuration

| Environment Variable         | Description                                                    | Default     | Required |
|:----------------------------:|----------------------------------------------------------------|-------------|:--------:|
| `WASTEBIN_WEBAPP_PORT`       |  The port wastebin will listen on                              | `3000`      | ❌       |
| `WASTEBIN_DB_USER`           |  The user to use when connecting to a database                 | `wastebin`  | ✅       |
| `WASTEBIN_DB_HOST`           |  The hostname or ip address of the datase to connect to        | `localhost` | ✅       |
| `WASTEBIN_DB_PORT`           |  The port to connect to the database on                        | `5432`      | ❌       |
| `WASTEBIN_DB_PASSWORD`       |  The password to connect to the database with                  |             | ✅       |
| `WASTEBIN_DB_NAME`           |  The name of the database to use                               | `wastebin`  | ❌       |
| `WASTEBIN_DB_MAX_IDLE_CONNS` |  The maximum number of idle connections to use                 | `10`        | ❌       |
| `WASTEBIN_DB_MAX_OPEN_CONNS` |  The maximum number of connections the database can have       | `50`        | ❌       |
| `WASTEBIN_DEV`               |  Disables postgres database support and uses a sqlite database | `false`     | ❌       |

## Running Wastebin

To run wastebin either use a docker-compose file like the one listen below or a `docker run` command

```yaml
version: '3'

services:
  wastebin:
    image: ghcr.io/coolguy1771/wastebin:0.0.1
    restart: always
    environment:
      - WASTEBIN_WEBAPP_PORT: 3000  # Optional, Defaults to 3000
      - WASTEBIN_DB_USER: "wastebin" # Optional, Defaults to wastebin
      - WASTEBIN_DB_HOST: "localhost" # Optional, Defaults to localhost
      - WASTEBIN_DB_PORT: 5432 # Optional, Defaults to 5432
      - WASTEBIN_DB_PASSWORD: "mysecretpassword"
      - WASTEBIN_DB_NAME: wastebin # Optional, defaults to wastebin
      - WASTEBIN_DB_MAX_IDLE_CONNS: 10 # Optional, defaults to 10
      - WASTEBIN_DB_MAX_OPEN_CONNS: 50 # Optional, defaults to 50
      - WASTEBIN_DEV: false # Use this to use a local sqlite database, if you want persistance you will need to specify a volume 
    ports:
      - "3000:3000"
  postgres:
    image: postgres:14.5
    restart: always
    enviorment:
      - POSTGRES_PASSWORD: "mysecretpassword"
      - POSTGRES_USER: "wastebin"
      - POSTGRES_DB: "wastebin"
    ports:
      - "5432:5432"
```

## Known Issues

- Currently pastes aren't deleted after being viewed
  - The burn toggle on the creation page doesn't do anything yet
- The CSS doesn't fill the whole page on the creation and view pages
- The raw view is just an HTML webpage and not a raw file causing issue with cURLing
- The LOG_LEVEL variable doesn't do anything

## Bugs and Suggestion

If you find a bug or have a suggestion, please open an issue or pull request
