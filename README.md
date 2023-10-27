# Gofi
![Gopher](/img/favicon.png)

## General informations
The purpose of this web app is to record expenses.

The HTML files are currently only in french.


## Technical informations

#### Built with 
- [go](https://go.dev/) & [gin-gonic](https://gin-gonic.com/)
- [htmx](https://htmx.org/)
- [pico](https://picocss.com/)
- [gopherize](https://gopherize.me/) for the nice logo


#### To run the app
- generate environment variables :
    ```bash
    export SQLITE_DB_FILENAME="gofi.db"
    ```
- locally :
    ```bash
    # exec initDB first
    cd /gofi/initDB
    go run .
    ```
    ```bash
    cd /gofi
    go run .
    ```

## TODO
- Ajout 5 dernières lignes enregistrées dans le listing des données à l'ouverture de insertrows
- Ajout sauvegarde sur Drive
    - avec table SQLite qui garde les ID de fichiers + le statut de l'upload
- Tester HTMX sur différents type de réponse : 200, 400, 500 ... : https://htmx.org/extensions/response-targets/ 
- Ajout import csv
- Ajout SQLite en WebAssembly


## Changelog
- 2023-10-27 : optimize database connections
- 2023-10-26 : add default parameters for new gofiID and new route to edit them
- 2023-10-23 : rework + add list of parameters in DB + handle accounts
- 2023-10-15 : remove gsheet to swap for SQLite
- 2023-09-24 : add read all gsheet, start to use params in a new gsheet.
- 2023-09-13 : initialize project