# url-shortener - An url shorten server written in Go

## APIs

- `POST /api/v1/urls`
    - Create a short URL with given expire date and original URL.

- `DELETE /api/v1/urls/<url_id>`
    - Deletes an existed URL.

- `GET /<url_id>`
    - Redirect URL with `<url_id> `to its original URL created by `POST /api/v1/urls`.

## Development Environments:

- Ubuntu (18.04)

- Golang (1.16.3)
    - required for building and running server

- Docker (20.10.6):
    - required for running standalone url-shortener server with local cache and database
    - required for end-to-end tests

- Python (3.8):
    - required for end-to-end tests

- MySQL (8.0.25)

## How to run

### 1. Standalone with docker-compose

- Go to the project root directory, and build the docker image for `url_shortener`

```shell
docker build -t url_shortener .
```

- Start the docker-compose

```shell
docker-compose up 
```

- Wait for state of servers to be `(healthy)`

```shell
docker-compose ps
```

```shell
          Name                         Command                  State                    Ports
------------------------------------------------------------------------------------------------------------
url_shortener_mysql_db      docker-entrypoint.sh mysqld      Up (healthy)   3306/tcp, 33060/tcp
url_shortener_redis_cache   docker-entrypoint.sh redis ...   Up (healthy)   6379/tcp
url_shortener_server        ./start.sh                       Up             0.0.0.0:80->80/tcp,:::80->80/tcp

```

- The server will listen at `http://localhost:80` by default. See different configurations [here](#How-to-configure).

### 2. Build from source:

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

- Build and start the server with default configurations. See different configurations [here](#How-to-configure).

```shell
./build.sh && ./start.sh
```

## How to configure

### The server can be configured by passing environment variables to it.

- `SERVER_PORT`: server listen port for `url_shortener` (default: `80`)
- `REDIRECT_SERVE_URL`: url to serve redirect api (default: `http://localhost`)
- `MYSQL_SERVER_ADDR`: mysql server addr (default: `localhost:3306`)
- `MYSQL_SERVER_ROOT_PASSWORD`: root password for connecting mysql server (default: `''`)
- `REDIS_SERVER_ADDR`: redis server addr (default: `localhost:6379`)
- `REDIS_SERVER_ADMIN_PASSWORD`: redis server admin password (default: `''`)

## How to run end-to-end tests

* NOTE: You need to start an `url_shortener` server first.

Install required python packages.

```shell
pip3 install -r requirements.txt
```

Run the python script `e2e.py` in the project root directory. The default endpoint for testing is `http://localhost`.

```shell
python3 e2e.py
```
