# Playwright e2e tests

Frontend tests using [Playwright](https://playwright.dev/) with Python, running in Docker.


## Prerequisites

- Docker installed
- A running local instance of the gofi server (default: `http://localhost:8083`)


## First run

Build the Docker image (only needed once, or after changing `requirements.txt`):

```bash
cd ~/gofi/e2e
sudo docker build --tag gofi-playwright .
```


## Run the tests

From the `e2e/` folder `cd ~/gofi/e2e`:

### main without launching browser
```bash
# headless (default)
sudo docker compose run --rm playwright
```

### alternate with a browser
```bash
# headed (GUI) — run xhost +local:docker first
xhost +local:docker
HEADED=true DISPLAY=$DISPLAY sudo -E docker compose run --rm playwright
```

### alternate with a browser and inspector on a specific test
```bash
# target a specific test with -k param
HEADED=true PWDEBUG=1 DISPLAY=$DISPLAY sudo -E docker compose run --rm playwright pytest tests/ -v -s -k "test_user_create_empty_fields_blocked"
```


## Test structure

```
e2e/
  conftest.py       ← shared fixtures (browser, page, base_url)
  requirements.txt  ← Python dependencies
  tests/
    test_user.py    ← user creation tests
```

## Adding tests

Each test file goes in `tests/` and follows the `test_*.py` naming convention.
Fixtures `page` and `base_url` are available in all test files via `conftest.py`.
