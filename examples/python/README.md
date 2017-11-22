# Doorman, Python / Flask example with a Web UI

## How to run doorman?

- Run `make serve -e POLICIES=examples/python/policies.yaml`


## How to run the server

    pipenv install
    API_AUDIENCE="SLocf7Sa1ibd5GNJMMqO539g7cKvWBOI" pipenv run python server.py


## How to run the web UI

- Serve the UI static files:

    cd ui/
    python3 -m http.server 3000

- Update your `/etc/hosts` so that you can resolve `doorman.local`:

	127.0.0.1 iam.local

- Access http://iam.local:3000/
- Click **Login**
