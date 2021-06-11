# url-shortener - An url shorten server written in Go

## APIs

- `POST /api/v1/urls`
    - Create a short URL with given expire date and original URL.

- `DELETE /api/v1/urls/<url_id>`
    - Deletes an existed URL.

- `GET /<url_id>`
    - Redirect URL with `<url_id>` to its original URL created by `POST /api/v1/urls`.

## Features and supported functionality:

- Default generated `url_id` is a string converted from a unique integer id starting from one.
    - can be extended to any string that is convertible back to the id.

- Recycle expired and deleted URLs
    - recycle for expired URLs is not realtime

- Basic end-to-end tests

## Development Environments:

- Ubuntu (18.04)

- Golang (1.16.3)
    - required for building and running server

- Docker Engine (20.10.6), docker-compose (1.29.1):
    - required for running standalone url-shortener server with local cache and database

- Python (3.8):
    - required for end-to-end tests

- MySQL (8.0.25)

- Redis (6.2.4)

## How to run

### 1. Standalone with docker-compose

- Go to the project root directory, and build the docker image for `url_shortener`

```shell
export URL_SHORTENER_IMAGE=url_shortener:demo

docker build -t ${URL_SHORTENER_IMAGE} .
```

- Start the docker-compose, and it will use environment variable `URL_SHORTENER_IMAGE` as the image name
  for `url_shortener`

```shell
# this line will make docker-compose to expose url_shortener server at port 16000
export EXTERNAL_SERVER_PORT=16000
# NOTE: you need to set external ip or hostname to make redirect API available to serve remotely
export REDIRECT_SERVE_ENDPOINT=http://localhost:16000

docker-compose up
```

- Wait for state of `mysql` and `redis` server to be `Up (healthy)`

```shell
docker-compose ps
```

- example output (`mysql` and `redis` are both in healthy state)

```shell
          Name                         Command                  State                    Ports
------------------------------------------------------------------------------------------------------------
url_shortener_mysql_db      docker-entrypoint.sh mysqld      Up (healthy)   3306/tcp, 33060/tcp
url_shortener_redis_cache   docker-entrypoint.sh redis ...   Up (healthy)   6379/tcp
url_shortener_server        ./start.sh                       Up             0.0.0.0:80->80/tcp,:::80->80/tcp

```

- The server will listen at `http://localhost:80` by default. See default and different
  configurations [here](#How-to-configure).

### 2. Build from source:

* Note: Need to set up MySQL and Redis server manually

- Run `init.sql` SQL script on the `mysql` server

```shell
mysql -u root -p < init.sql
```

- Go to the `script` directory

```shell
cd script
```

- Add executable permission to `build.sh` and `start.sh`

```shell
chmod +x build.sh start.sh
```

- Build and start the server with default configurations. See default and different
  configurations [here](#How-to-configure).

```shell
# this line will make url_shortener server listen at :16000
export SERVER_PORT=16000
# NOTE: you need to set external ip or hostname to make redirect API available to serve remotely
export REDIRECT_SERVE_ENDPOINT=http://localhost:16000

./build.sh && ./start.sh
```

## How to configure

### The server can be configured by passing environment variables to it.

- `SERVER_PORT` : server listen port for `url_shortener` (default: `80`)
- `REDIRECT_SERVE_ENDPOINT` : endpoint to serve redirect api (default: `http://localhost`)
- `MYSQL_SERVER_ADDR` : mysql server addr (default: `localhost:3306`)
- `MYSQL_SERVER_ROOT_PASSWORD` : root password for connecting mysql server (default: `''`)
- `REDIS_SERVER_ADDR` : redis server addr (default: `localhost:6379`)
- `REDIS_SERVER_ADMIN_PASSWORD` : redis server admin password (default: `''`)
- `CHECK_EXPIRATION_INTERVAL` : time interval in seconds to check expired records (default: 60)

### For standalone docker-compose environment, there are additional environment variables:

- `URL_SHORTENER_IMAGE` : image name for `url_shortener` docker (default: `url_shortener`)
- `EXTERNAL_SERVER_PORT` : external server port for publishing internal `SERVER_PORT` (default: `80`)

## How to run end-to-end tests

* NOTE: You need to start an `url_shortener` server first. [How to start a server.](#How-to-run)

Install required python packages.

```shell
pip3 install -r requirements.txt
```

Run the python script `test_e2e.py` in the project root directory. The default endpoint for testing
is `http://localhost`.

```shell
export CHECK_EXPIRATION_INTERVAL=${CHECK_EXPIRATION_INTERVAL}
# SHORTENER_E2E_ENDPOINT is the endpoint of the url-shortener server that we want to tested
# In most cases, it should be the same as the endpoint for serving redirect API.
export SHORTENER_E2E_ENDPOINT=${REDIRECT_SERVE_ENDPOINT}

python3 test_e2e.py
```
