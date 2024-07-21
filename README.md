# GOFI 
![Gopher](/assets/img/android-chrome-192x192.png)


## General informations
The purpose of this web app is to record and manage your money.  
The HTML files are currently only in french.

Features supported (all the data are registered in a local SQLite DB):
- main features
    - create users
    - auth with 1 active session per user
    - save general parameters preferences per user
    - record expenses per user, account and category
    - manage recurrent expenses or gains
    - handle multiple bank accounts per user and allow transfer between them
    - validate or cancel each record
    - import CSV files to insert/update data in bulk
    - export CSV files to keep/use all the data with other apps
    - stats year by year with current balance available per account
    - smartphone first front-end, tested on the viewport:
        - Screen Width: 360 pixels
        - Screen Height: 640 pixels
        - [screen viewport on viewportsizer](https://viewportsizer.com/lite/)
- admin features
    - ~~(optional) generate and manage backup~~ (with another app)
    - shutdown the application (also checkpoint SQLite, then clean the `db-shm` and `db-wal` files)


## API and UI monolith
This application is a monolith with some endpoints for the API and others for the UI.  
Both uses the same functions, but return JSON for the API and HTML for the UI.


## Run the app
- prerequisites : 
    1. [golang](https://go.dev/doc/install)
    2. [templ](https://templ.guide/quick-start/installation)
- test the app :
    ```bash
        # generate environment variables :
        export SQLITE_DB_FILENAME="test.db"
        export COOKIE_LENGTH=64
        export EXE_PATH="/gofi"
        export ADMIN_EMAIL="test@test.test"
        export ADMIN_EMAIL_B="testb@test.test"
        # run the tests (create a new DB named test.db) :
        cd /gofi
        go clean -testcache
        go test ./data/dbscripts/initDB
        go test ./back/api/test/users
        go test ./back/api/test/params
        go test ./back/api/test/records
        go test ./back/api/test/csv
        go test ./back/api/test/shutdown
    ```
- run the app with the real database :
    ```bash
        # only the first time.
        # exec initDB first to create DB file with required tables
        export SQLITE_DB_FILENAME="gofi.db"
        export EXE_PATH="/gofi"
        cd /gofi
        go run ./data/dbscripts/initDB
        # the DB is created in the "dbscripts" folder, move it under: "data/dbFiles"
    ```
    ```bash
        export SQLITE_DB_FILENAME="gofi.db"
        export COOKIE_LENGTH=64
        export EXE_PATH="/gofi"
        export ADMIN_EMAIL="example@gmail.com"
        export ADMIN_EMAIL_B="exampleb@gmail.com"
        cd /gofi
        templ generate
        go run .
    ```
- on Windows Powershell:
    - replace `export ` with `$Env:`
    - replace `cd /gofi` with `cd c:\gofi\`


## Folder Structure
> - legend:
>   - this is a Go `package`
>   - this is a standard ***folder***
> - Circular dependency is forbidden.

- ***back***
    1. ***routes***
        - `routes`
        - The **routes** package contains the configuration of the app.
    2. ***appmiddleware***
        - `appmiddleware`
            - The **appiddleware** package contains code executed on each request.
    3. ***api***
        - ***test***
            - All the behaviour testing of the back-end is done here under test subpackages.
        - `api`
            - The **api** package contains all the handlers for the back-end part of the app.

- ***data***
    1. ***appdata*** 
        - `appdata`
        - The **appdata** package contains global variables, funcs and go data structs.
    2. ***sqlite***
        - `sqlite`
        - The **sqlite** package contains all the database interaction.
    3. ***dbFiles*** 
        - contains the database files.
    4. ***dbscripts*** 
        - contains database script to execute manualy.

- ***front***
    - ***htmlComponents***
        - `htmlComponents`
            - All the HTML code is structured here with maximum usage of components to minimize repetition.
    - `front`
        - The **front** package contains all the handlers for the front-end part of the app.

- ***assets***
    - Everything here is used as file server for the front-end part.
    - ***css***
    - ***fonts***
    - ***img***
    - ***js***


## API and SQL corresponding statements
Operation | HTTP Method     | SQL Statement
--------- | --------------- | -------------
Create    | POST (body)     | INSERT
Read      | GET (params)    | SELECT
FullEdit  | PUT (body)      | UPDATE
PartEdit  | PATCH (body)    | UPDATE
Delete    | DELETE (params) | DELETE

- POST + PUT + PATCH with JSON body
- GET + DELETE with URL params


## GO struct usage
- the HTML attribute `name` is used when a form is sent to the backend (application/x-www-form-urlencoded)
    - the used `name` in the HTML file must correspond to the struct name (case sensitive)
- the struct `JSON name` is used for API calls with JSON body (case insensitive)
