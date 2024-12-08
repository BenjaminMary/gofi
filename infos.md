# GOFI infos
[README](README.md)

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
    - the form use (this package)[https://github.com/ajg/form]
    - possible to bind struct in struct by joining with `.` : `parent.child`
- the struct `JSON name` is used for API calls with JSON body (case insensitive)

---

[README](README.md)