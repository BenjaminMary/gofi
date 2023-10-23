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
    export SQLITE_DB_FILENAME="sqlite.db"
    ```
- locally :
    ```bash
    go run .
    ```
- with Docker :
    ```bash
    docker build --tag name/gofi:tag .

    docker run --detach -e type -e project_id -e private_key_id -e private_key -e client_email -e client_id -e auth_uri -e token_uri -e auth_provider_x509_cert_url -e client_x509_cert_url -e universe_domain --publish 127.0.0.1:8082:8082 imageIdJustBuilt
    ```

## TODO
- Améliorer page Insert Rows
    - variable sur la liste des catégories
- Ajout sauvegarde sur Drive
    - avec table SQLite qui garde les ID de fichiers + le statut de l'upload
- Tester HTMX sur différents type de réponse : 200, 400, 500 ... : https://htmx.org/extensions/response-targets/ 
- Ajout import csv
- Ajout SQLite en WebAssembly


## Changelog
- 2023-10-15 : remove gsheet to swap for SQLite
- 2023-09-24 : add read all gsheet, start to use params in a new gsheet.
- 2023-09-13 : initialize project