<!-- https://github.com/othneildrew/Best-README-Template -->
<a id="readme-top"></a>

<!--[![Contributors][contributors-shield]][contributors-url]-->
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]


<!-- PROJECT LOGO -->
<br />
<div align="center">
    <img src="./assets/img/android-chrome-192x192.png" alt="Logo" width="80" height="80" />
    <h3 align="center">GOFI</h3>
    <p align="center">
        Full Stack App to manage your expenses
        <br />
        <a href="https://gofi.benjamin-mary.com/"><strong>Explore the app »</strong></a>
        <!-- TODO doc 
        <br />
        <a href="https://github.com/BenjaminMary/gofi"><strong>Explore the docs »</strong></a> -->
        <br />
        <br />
        <!-- TODO demo url 
        <a href="https://github.com/BenjaminMary/gofi">View Demo</a>
        · -->
        <a href="https://github.com/BenjaminMary/gofi/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
        ·
        <a href="https://github.com/BenjaminMary/gofi/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
    </p>
</div>


<!-- TABLE OF CONTENTS -->
<details>
    <summary>Table of Contents</summary>
    <ol>
        <li>
            <a href="#about-the-project">About The Project</a>
            <ul>
                <li><a href="#built-with">Built With</a></li>
                <li><a href="#general-informations">General informations</a></li>
                <li><a href="#features">Features</a></li>
            </ul>
        </li>
        <li>
            <a href="#getting-started">Getting Started</a>
            <ul>
                <li><a href="#prerequisites">Prerequisites</a></li>
                <li><a href="#installation">Installation</a></li>
            </ul>
        </li>
        <li><a href="#usage">Usage</a></li>
        <!-- <li><a href="#roadmap">Roadmap</a></li>
        <li><a href="#contributing">Contributing</a></li> -->
        <li><a href="#license">License</a></li>
        <li><a href="#contact">Contact</a></li>
        <li><a href="#acknowledgments">Acknowledgments</a></li>
    </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

### Built With
[![SQLite][SQLite-shield]][SQLite-url] 
[![Go][Golang]][Golang-url] 
[![HTMX][HTMX-shield]][HTMX-url] 


### General informations
- The purpose of this web app is to record and manage your money.  
- The HTML files (templ files here) are currently only in french.
- The deployment consist of 3 things : 
    1. the binary built which contains the front-end and back-end code
    2. an asset folder with all the images, icons and fonts
    3. a SQLite database file
- This application is a monolith with some endpoints for the API and others for the UI. Both uses the same functions, but return JSON for the API and HTML for the UI.


### Features
- basic features for users
    - onboarding checklist to discover basics features
    - record expenses per user, account and category
    - edit / update your records
    - stats: 
        - year by year with current balance available per account
        - year by year or month by month per category
- advanced features for users
    - ways to handle recurrent records
        - use URL shortcuts with default values in forms to speed up data entries (like groceries)
		    - `/record/insert/{account}/{category}/{product}/{priceDirection}/{price}`
		    - `/record/insert/LA/Epargne/designation/+/56.78`
            - space = `%20`
		    - `/record/insert/LA/Epargne/designation%20avec%20espace/+/56.78`
        - manage recurrent expenses or gains in a specific tab with a schedule
    - handle multiple categories per user and allow budgeting for each, 2 budgeting options:
        - reset the budget each period
        - keep the rest of the last budget period and add it to the next
    - handle multiple bank accounts per user and allow transfer between them
    - validate or cancel each record
    - lend / borrow with registered tiers
- bulk data operation for users
    - import CSV files to insert / update / delete records
    - export CSV files to keep / use all the data with other apps or update your data in bulk
- generic features in app
    - create users
    - auth with 1 active session per user
    - save general parameters preferences per user
    - smartphone first front-end
        - tested on the viewport:
            - Screen Width: 360 pixels
            - Screen Height: 640 pixels
            - [screen viewport on viewportsizer](https://viewportsizer.com/lite/)
        - tested on Chrome for Android and Chrome for Windows Desktop
            - if you have any visual trouble on a different Browser/OS combination, submit an issue
- admin features
    - shutdown the application with SQLite checkpoint, which clean the `db-shm` and `db-wal` files



<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started
To get a local copy up and running follow these simple example steps.


### Prerequisites
1. [golang](https://go.dev/doc/install)
2. [templ](https://templ.guide/quick-start/installation)


### Installation
- test the app :
    ```bash
        # generate environment variables :
        export SQLITE_DB_FILENAME="test.db"
        export COOKIE_LENGTH=64
        export EXE_PATH="/gofi"
        export ADMIN_EMAIL="test@test.test"
        export ADMIN_EMAIL_B="testb@test.test"
        export NOTIFICATION_FLAG=0
        export NOTIFICATION_URL="https://notification.server/example"
        export HEADER_IP="header-IP-test"
        # run the tests (create a new DB named test.db) :
        cd /gofi
        go clean -testcache
        go test ./data/dbscripts/initDB
        go test ./back/api/test/users
        go test ./back/api/test/params
        go test ./back/api/test/records
        go test ./back/api/test/csv
        go test ./back/api/test/save
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
        export NOTIFICATION_FLAG=1
        export NOTIFICATION_URL="https://notification.server/example"
        export HEADER_IP="header-IP"
        cd /gofi
        templ generate
        go run .
    ```
- on Windows Powershell:
    - replace `export ` with `$Env:`
    - replace `cd /gofi` with `cd c:\gofi\`


<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage
- more informations on the structure of this app can be found in [infos.md](infos.md)
<!-- TODO usage
Use this space to show useful examples of how a project can be used. Additional screenshots, code examples and demos work well in this space. You may also link to more resources.

_For more examples, please refer to the [Documentation](https://example.com)_ -->



<!-- ROADMAP -->
<!-- ## Roadmap

- [ ] Feature 1
- [ ] Feature 2
- [ ] Feature 3
    - [ ] Nested Feature

See the [open issues](https://github.com/BenjaminMary/gofi/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p> -->



<!-- CONTRIBUTING -->
<!--## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please open an issue with the tag "enhancement" or create a pull request.
Don't forget to give the project a star! Thanks again!

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Top contributors:

<a href="https://github.com/BenjaminMary/gofi/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=BenjaminMary/gofi" alt="contrib.rocks image" />
</a>
-->


<!-- LICENSE -->
## License
- Distributed under the MPL-2.0 License. See `LICENSE` file for more information.
- Link to the original template of the license: https://www.mozilla.org/en-US/MPL/2.0/

©2024 Benjamin MARY



<!-- CONTACT -->
## Contact
- Benjamin MARY - benjamin-mary@outlook.com
- Project Link: [https://github.com/BenjaminMary/gofi](https://github.com/BenjaminMary/gofi)

Join the community on (limited to 100 invitations, contact me if used):  
[![Discord][Discord-shield]][Discord-url]



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments
* [Pico CSS](https://picocss.com/)
* [Apexcharts graph library](https://apexcharts.com/)
* [Gopherize mascot](https://gopherize.me/)
* [Feather icons](https://feathericons.com/)
* [Lucide icons](https://lucide.dev/)
* [Icomoon fonts](https://icomoon.io/)
* [UnDraw illustrations](https://undraw.co/illustrations)


<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
-->

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/BenjaminMary/gofi.svg?style=for-the-badge
[contributors-url]: https://github.com/BenjaminMary/gofi/graphs/contributors
[stars-shield]: https://img.shields.io/github/stars/BenjaminMary/gofi.svg?style=for-the-badge
[stars-url]: https://github.com/BenjaminMary/gofi/stargazers
[issues-shield]: https://img.shields.io/github/issues/BenjaminMary/gofi.svg?style=for-the-badge
[issues-url]: https://github.com/BenjaminMary/gofi/issues
[license-shield]: https://img.shields.io/github/license/BenjaminMary/gofi.svg?style=for-the-badge
[license-url]: https://github.com/BenjaminMary/gofi/blob/master/LICENSE

[Golang]: https://img.shields.io/badge/Go-00ADD8?logo=Go&logoColor=white&style=for-the-badge
[Golang-url]: https://go.dev/
[SQLite-shield]: https://img.shields.io/badge/SQLite-003B57?style=for-the-badge&logo=sqlite&logoColor=white
[SQLite-url]: https://www.sqlite.org/
[HTMX-shield]: https://img.shields.io/badge/HTMX-36C?style=for-the-badge&logo=htmx&logoColor=white
[HTMX-url]: https://htmx.org/

[Discord-shield]: https://img.shields.io/badge/Discord-7289DA?style=for-the-badge&logo=discord&logoColor=white
[Discord-url]: https://discord.gg/R9ysnyjayw
<!-- 100 uses of the discord invit without time expiration -->
