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
- auth
    - ajout logout + redirect vers login si pas co
    - ajout gestion si déjà login pour pages login + createUser 
    - ajout check des dates de timeout du cookie à chaque requête + réinit du cookie quand 5 min sont passés
- ajout préférences utilisateur:
    - gestion des préférences de format de date EN + FR avec / ou -
    - gestion des préférences de format csv séparateur colonne + separateur décimal
    - laisser l'overide possible dans les parties import/export csv, mais préselectionner la préférence
- Ajout sauvegarde DB SQLite sur Drive
    - avec table SQLite qui garde les ID + nom + date de fichiers sauvegardés + le statut de l'upload
- PWA
    - Ajout SQLite en WebAssembly ?

- Tester HTMX sur différents type de réponse : 200, 400, 500 ... : https://htmx.org/extensions/response-targets/ 
- Amélioration download fichier csv : voir si possible de faire mieux directement via le serveur à la place du js


## Changelog
- 2023-11-12 : reorganize main, split funcs in another file
- 2023-11-12 : add session management in DB and transform gofiID to INT + cookie to random STR
- 2023-11-05 : add different date formats to allow YYYY-MM-DD, DD/MM/YYYY, YYYY/MM/DD, DD-MM-YYYY
- 2023-11-05 : improve date handle mostly for csv import
- 2023-11-05 : add import csv + small update on export + add leading 0 on some dates
- 2023-11-04 : improve export csv with ID in filename, can handle different IDs in // and delete the file
- 2023-10-29 : add export csv 
- 2023-10-27 : add last 5 rows registered in the list on insertrows GET page, also add account info 
- 2023-10-27 : optimize database connections
- 2023-10-26 : add default parameters for new gofiID and new route to edit them
- 2023-10-23 : rework + add list of parameters in DB + handle accounts
- 2023-10-15 : remove gsheet to swap for SQLite
- 2023-09-24 : add read all gsheet, start to use params in a new gsheet.
- 2023-09-13 : initialize project