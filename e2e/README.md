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
GOFI_E2E_ADMIN_EMAIL="test@test.test" GOFI_E2E_ADMIN_PASSWORD="test" sudo -E docker compose run --rm playwright
```

```bash
# run a single file
GOFI_E2E_ADMIN_EMAIL="test@test.test" GOFI_E2E_ADMIN_PASSWORD="test" sudo -E docker compose run --rm playwright pytest tests/test_maintenance.py

# run 2 files
sudo docker compose run --rm playwright pytest tests/test_record_alter.py tests/test_record_edit.py


# run a single test
sudo docker compose run --rm playwright pytest tests/test_home.py::test_home_online_advanced_mode
```

### alternate with a browser
```bash
# headed (GUI) — run xhost +local:docker first
xhost +local:docker
HEADED=true DISPLAY=$DISPLAY sudo -E docker compose run --rm playwright

HEADED=true DISPLAY=$DISPLAY sudo -E docker compose run --rm playwright pytest tests/test_record_edit.py -v -s

HEADED=true DISPLAY=$DISPLAY sudo -E docker compose run --rm playwright pytest tests/ -v -s -k "test_record_edit_page_loads"
```

### alternate with a browser and inspector on a specific test
```bash
# target a specific test with -k param
HEADED=true PWDEBUG=1 DISPLAY=$DISPLAY sudo -E docker compose run --rm playwright pytest tests/ -v -s -k "test_user_create_empty_fields_blocked"

# target a specific file
HEADED=true PWDEBUG=1 DISPLAY=$DISPLAY sudo -E docker compose run --rm playwright pytest tests/test_record_edit.py -v -s
```


## Test structure

```
e2e/
  conftest.py            ← shared fixtures (browser, page, base_url, created_user, auth_state, logged_in_page)
  requirements.txt       ← Python dependencies
  tests/
    test_home.py         ← home page tests (/ route)
    test_user_create.py  ← user creation tests (/user/create)
    test_user_login.py   ← login tests (/user/login)
```

## Adding tests

Each test file goes in `tests/` and follows the `test_*.py` naming convention.
- Use `page` fixture for unauthenticated tests
- Use `logged_in_page` fixture for authenticated tests (depends on `created_user`)
- Fixtures are available in all test files via `conftest.py`
