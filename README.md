# Station Manager: database module

## Regenerating the database models

The build uses `sqlboiler` to generate the database models. To regenerate the models, each database
must be rebuilt separately, but both follow roughly the same steps.

### SQLite

1. Delete the existing development database file at `build/db/data.db`.
2. Run the `database/sqlite/example/main.go` file.
3. `cd database/sqlite`
4. `sqlboiler sqlite3`

### Postgres

1. Delete the existing development database file at `build/db/postgres_data` - this will require the use of `sudo`.
2. Start the Postgres Docker container.

   ```
    sudo systemctl start docker.service`
    docker-compose up
   ```

3. Run the `database/postgres/example/main.go` file.
4. `cd database/postgres`
5. `export PSQL_PASS=[some password]` - obviously replace `[some password]` with the correct password.
6. `sqlboiler psql`
