# Doorman, Python / Flask example with a Web UI

## How to run doorman?

    make serve -e POLICIES=examples/python/policies.yaml


## How to run the server

    cd examples/python/
    pipenv install
    export DOORMAN_SERVER=http://localhost:8080
    export API_AUDIENCE="SLocf7Sa1ibd5GNJMMqO539g7cKvWBOI"
    pipenv run python server.py

## How to run the web UI

- Serve the UI static files:

    cd examples/python/ui/
    python3 -m http.server 3000

- Update your `/etc/hosts` so that you can resolve `iam.local`:

	127.0.0.1 iam.local

- Access http://iam.local:3000/
- Click **Login**
