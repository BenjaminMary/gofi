# Gofi
![Gopher](/front/img/android-chrome-192x192.png)

## General informations
The purpose of this web app is to record expenses.

Features supported (all the data are registered in a local SQLite DB):
- main features
    - create users
    - auth with 1 active session per user
    - save general parameters preferences per user
    - record expenses per user
    - import CSV files to insert/update data in bulk
    - export CSV files to keep/use all the data with other apps 
- admin features
    - (optional) generate and manage backup

The HTML files are currently only in french.


## Technical informations

#### Built with 
- [go](https://go.dev/) & [gin-gonic](https://gin-gonic.com/)
- [htmx](https://htmx.org/)
- [pico](https://picocss.com/)
- [gopherize](https://gopherize.me/) for the nice logo
- [sqlite](https://www.sqlite.org/)


#### To run the app
- generate environment variables :
    ```bash
    export GIN_MODE="release"
    export SQLITE_DB_FILENAME="gofi.db"
    export COOKIE_LENGTH=64
    export EXE_PATH="/gofi"
    export ADMIN_EMAIL="example@gmail.com"
    export DRIVE_SAVE_ENABLED=0
    ```
- locally :
    ```bash
    # exec initDB first to create required tables
    # only the first time.
    cd /gofi/initDB
    go run .
    ```
    ```bash
    cd /gofi
    go run .
    ```

#### OPTIONAL: Generate Backups on Google Drive
This optional feature adds some prerequisites:
- enable it with env var: `DRIVE_SAVE_ENABLED=1`
- only the `ADMIN_EMAIL` (also set with env var) will be able to use this feature
- you need to make a [Google Service Account](https://developers.google.com/workspace/guides/create-credentials#service-account) to get the following credentials.
- add these credentials as environment variables:
    ```bash
    export type="service_account"
    export project_id="project"
    export private_key_id="XY"
    export private_key="-----BEGIN PRIVATE KEY-----\nXYZ\n-----END PRIVATE KEY-----\n"
    export client_email="X@Y.iam.gserviceaccount.com"
    export client_id="1"
    export auth_uri="https://accounts.google.com/o/oauth2/auth"
    export token_uri="https://oauth2.googleapis.com/token"
    export auth_provider_x509_cert_url="https://www.googleapis.com/oauth2/v1/certs"
    export client_x509_cert_url="https://www.googleapis.com/robot/v1/metadata/x509/X%Y.iam.gserviceaccount.com"
    export universe_domain="googleapis.com" 
    ```

## TODO
☑☐☒
- dès que toutes les fonctionnalités essentielles sont en place, démarrer des tests fonctionnels
- ajout préférences utilisateur:
    - gestion des préférences de format de date EN + FR avec / ou -
    - gestion des préférences de format csv séparateur colonne + separateur décimal
    - laisser l'overide possible dans les parties import/export csv, mais préselectionner la préférence
    - ☐ partie import export: 
        - ☑ gérer les formats ANSI (à faire pour ensuite visualiser les é dans Excel par défaut) et UTF8 (déjà ok)
            - ☒ pas gestion du format ANSI
            - ☑ force l'utilisation du UTF8 avec ajout de caractères de contrôle
            - ☑ CRLF et LF gérés en ajoutant une colonne non utilisée en fin de fichier
        - ☑ ajouter la possibilité de supprimer des lignes en mettant des "-" devant les ID de lignes
            - ☑ UPDATE (et non pas réel DELETE) avec mise à 0 du prix + MAJ compte et catégo en "-" (permet de mieux gérer les exports de données et de les réutiliser lorsque partiel pour obtenir le bon résultat avec l'ensemble des modifs appliquées les unes à la suite des autres)
            - ☒ via DELETE, aurait posé problème lorsque la ligne est créé puis exportée, et ensuite DELETE avec un nouvel export où aucune info sur le DELETE n'aurait été présente car ligne inexistante en DB (l'export à pour objectif d'exporter uniquements les dernières modifications, voir pour mettre une option pour tout rééxporter)
        - ☐ MAJ le champ `exported` lors des exports et modifications de données
        - ☐ mettre un champ `lastCSVexport` de type date par user à ramener dans le menu (ok si - d'1 mois, ko sinon)
            - ☐ compter le nombre de lignes à exporter et afficher/bloquer un import si différent de 0 ?
        - ☐ ajouter un template de fichier csv
        - ☐ objectif: chaque export génère un fichier avec l'ensemble des dernières modifs
            - ☐ en jouant toutes les sauvegardes historisées dans l'ordre chronologique, on retrouve l'état des données souhaité
            - ☐ mettre une option d'export de toutes les lignes même celles non modifiées
- ☑ PRIO ajout validation des dépenses
    - ☑ système qui ramène l'ensemble des lignes encore non validées
    - ☑ voir pour permettre de la validation de groupe en saisissant une date unique et en sélectionnant X lignes
- ❗ PRIO Ajout de statistiques 
    - ☑ visualisation des données avec filtre et tri via table simple
        - ☐ voir pour mettre un tableur ? + rendre editable ou suppr de ligne
            - https://github.com/wenzhixin/bootstrap-table
            - https://github.com/jspreadsheet/ce
                - https://bossanova.uk/jspreadsheet/v4/docs/quick-reference
    - sur les dépenses
    - sur le nombre de requêtes
    - ❗ globales sur les montants dispo par compte
- Ajout sauvegarde DB SQLite sur Drive
    - ☑ avec table SQLite qui garde les ID + nom + date de fichiers sauvegardés + le statut de l'upload (pas besoin l'API Google redonne toutes les infos)
    - ☑ voir pour fermer le server et faire la sauvegarde au restart après quelques commandes de nettoyage de DB (semble ok)
    - ☑ voir si la gestion d'une seule ouverture/fermeture DB ferait fonctionner le PRAGMA wal_checkpoint(TRUNCATE) sans retourner BUSY
        - ☑ obj nettoyer les fichiers wal + shm avant sauvegarde
    - ☐ cron based backup : https://litestream.io/alternatives/cron/ + monitoring : https://deadmanssnitch.com/account/sign_up?plan=the_lone_snitch
- PWA
    - Ajout SQLite en WebAssembly ?
- voir pour réduire le nombre d'ouverture/fermeture de DB
    - en go, open démarre un pool de connexion, mettre en place une route DB.Stats pour avoir des infos en temps réel
    - https://go.dev/doc/database/open-handle
    - https://go.dev/doc/database/manage-connections
    - tuto DB in Go : https://dev.to/techschoolguru/how-to-handle-db-errors-in-golang-correctly-11ek
    - DB test case : https://stackoverflow.com/questions/48196746/using-ping-to-find-out-if-db-connection-is-alive-in-golang
- autres améliorations non prioritaires
    - voir pour split le SQL dans des fichiers .sql (exemple: https://github.com/qustavo/dotsql)
    - voir pour split le HTML dans des fichiers séparés (via templating par block?) OU mieux gérer le HTML directement dans go: https://github.com/a-h/templ
    - voir pour créer des packages mieux définis et pouvoir les sortir complètement de cet app (exemple partie auth/session)
    - Tester HTMX sur différents type de réponse : 200, 400, 500 ... : https://htmx.org/extensions/response-targets/ 
    - Amélioration download fichier csv : voir si possible de faire mieux directement via le serveur à la place du js
    - auth
        - check des changements d'IP / user agent pour forcer un relogin


## Changelog
- 2024-01-09 : new global statistics page.
- 2024-01-09 : fix advanced mode to validate or cancel already checked records.
- 2024-01-08 : fix some front UI following the front folder. Add an advanced mode to validate or cancel specific records if needed.
- 2024-01-06 : new page to validate or cancel records.
- 2024-01-03 : some more HTML improvements.
- 2024-01-02 : add buttons and svg icons, app relooking. Store svg icons in a dedicated file.
- 2024-01-01 : add DELETE option from import csv by addind a "-" before the ID. Add new controls before import csv.
- 2024-01-01 : fix backup download.
- 2023-12-31 : add a new empty column at the end of the csv to handle CRLF end of line. Rename a used DB field, BREAKING CHANGE, needs to run the `migrateDB.go` file.
- 2023-12-30 : add UTF-8 control characters on the csv file export, and control their presence before import.
- 2023-12-29 : add groups on front pages where width is > 1000px with Pico class.
- 2023-12-27 : add a front folder. Use Pico with classes, ex: <code>class="grid"</code>. Handle positive and negative values from the html input page. Add theme switcher on the index page (dark or light).
- 2023-12-23 : add a subtotal row in the data table page.
- 2023-12-22 : improve struct and readability for user params.
- 2023-12-17 : add a screen to visualize data with filter, sort and limit.
- 2023-12-15 : fix import CSV, add some doc on how it works. Also improve admin backup part.
- 2023-12-10 : add optional backup saves with Google Drive API. Also add context with timeout everywhere and simplify DB open.
- 2023-11-20 : add var env for the executable file path + change port used
- 2023-11-13 : auto update cookie when idle timeout reached, force new login when absolute timeout reached (all dates are generated with SQLite)
- 2023-11-12 : logo update + logout feature + cookie length param + rework some html
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