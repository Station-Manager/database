# Station Manager: database module

## Postgres

This database is used by the server for the online services, whereas SQLite is used for
the local database.

### Development setup

`cd database/postgres`

`sudo systemctl start docker.service`

`docker-compose --env-file .env up -d`

`docker-compose down`

### SQLBoiler

Your location must be the `database/postgres` directory.

`export PSQL_PASS=[some password]`

`sqlboiler psql`

## SQLite

