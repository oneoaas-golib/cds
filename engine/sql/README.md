# CDS Database Management

## Organization of this folder

This folder contains Migration scripts. Each migration scripts contains Upgrade and Downgrade statements.

## How to use

It is possible to **upgrade** and **downgrade** your database schema. It can also show to the migration **status**.

Commands below ask you to run :

```bash
    $ <PATH_TO_CDS>/engine/api/api database -h
```

`api` binary can be generated by running:

```bash
    $ cd <PATH_TO_CDS>/engine/api/
    $ go build
```

SubCommand database:

```
    $ <PATH_TO_CDS>/engine/api/api database -h
    Manage CDS database

    Usage:
    api database [command]

    Available Commands:
    upgrade     Upgrade schema
    downgrade   Downgrade schema
    status      Show current migration status

    Global Flags:
        --db-host string       DB Host (default "localhost")
        --db-maxconn int       DB Max connection (default 20)
        --db-name string       DB Name (default "cds")
        --db-password string   DB Password
        --db-port string       DB Port (default "5432")
        --db-sslmode string    DB SSL Mode: require (default), verify-full, or disable (default "require")
        --db-timeout int       Statement timeout value (default 3000)
        --db-user string       DB User (default "cds")

    Use "api database [command] --help" for more information about a command.
```

### Upgrade database

This will never-applied migration scripts (ie. run the `Up` parts) and mark them as applied. You can user `dry-run` option to see which scripts would be executed.

```shell
    $ <PATH_TO_CDS>/engine/api/api database upgrade -h
    Migrates the database to the most recent version available.

    Usage:
    api database upgrade [flags]

    Flags:
        --dry-run              Dry run upgrade
        --limit int            Max number of migrations to apply (0 = unlimited)
        --migrate-dir string   CDS SQL Migration directory (default "./engine/sql")

    Global Flags:
        --db-host string       DB Host (default "localhost")
        --db-maxconn int       DB Max connection (default 20)
        --db-name string       DB Name (default "cds")
        --db-password string   DB Password
        --db-port string       DB Port (default "5432")
        --db-sslmode string    DB SSL Mode: require (default), verify-full, or disable (default "require")
        --db-timeout int       Statement timeout value (default 3000)
        --db-user string       DB User (default "cds")
```

### Downgrade database

This will undo migration scripts (ie. run the `Down` parts) and mark them never applied. You can user `dry-run` option to see which scripts would be executed.

```shell
    $ <PATH_TO_CDS>/engine/api/api database downgrade -h
    Migrates the database to the most recent version available.

    Usage:
    api database upgrade [flags]

    Flags:
        --dry-run              Dry run upgrade
        --limit int            Max number of migrations to apply (0 = unlimited)
        --migrate-dir string   CDS SQL Migration directory (default "./engine/sql")

    Global Flags:
        --db-host string       DB Host (default "localhost")
        --db-maxconn int       DB Max connection (default 20)
        --db-name string       DB Name (default "cds")
        --db-password string   DB Password
        --db-port string       DB Port (default "5432")
        --db-sslmode string    DB SSL Mode: require (default), verify-full, or disable (default "require")
        --db-timeout int       Statement timeout value (default 3000)
        --db-user string       DB User (default "cds")
```

### Database migration status

Show migration status.

```shell
    $ <PATH_TO_CDS>/engine/api/api database status --db-host <host> --db-password <password> --db-name <database> --migrate-dir ./engine/sql/migrations
    |          MIGRATION           |                APPLIED                |
    |------------------------------|---------------------------------------|
    | 000_create_all.sql           | 2016-10-26 16:01:08.575758 +0200 CEST |

```

## How to write scripts

Rules:

- Never delete any scripts
- Always increment migration scripts prefix number
- Create scripts must be updated whenever Migration scripts are created or updated
- Never forget `Down` parts in migration scripts

Migrations are defined in SQL files, which contain a set of SQL statements. Special comments are used to distinguish up and down migrations.


```sql
    -- +migrate Up
    -- SQL in section 'Up' is executed when this migration is applied
    CREATE TABLE people (id int);


    -- +migrate Down
    -- SQL section 'Down' is executed when this migration is rolled back
    DROP TABLE people;
```

You can put multiple statements in each block, as long as you end them with a semicolon (;).

If you have complex statements which contain semicolons, use StatementBegin and StatementEnd to indicate boundaries:


```sql
    -- +migrate Up
    CREATE TABLE people (id int);

    -- +migrate StatementBegin
    CREATE OR REPLACE FUNCTION do_something()
    returns void AS $$
    DECLARE
    create_query text;
    BEGIN
    -- Do something here
    END;
    $$
    language plpgsql;
    -- +migrate StatementEnd

    -- +migrate Down
    DROP FUNCTION do_something();
    DROP TABLE people;
```
