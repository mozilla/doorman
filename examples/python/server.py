#
# https://auth0.com/docs/quickstart/backend/python#add-api-authorization
# https://github.com/auth0-samples/auth0-python-api-samples/tree/master/00-Starter-Seed

import functools
import json
import os

from flask import Flask, request, jsonify, _app_ctx_stack
from flask_cors import cross_origin

import doorman

HERE = os.path.abspath(os.path.dirname(__file__))

DOORMAN_SERVER = os.getenv("DOORMAN_SERVER", "http://localhost:8080")
API_AUDIENCE = os.getenv("API_AUDIENCE")

app = Flask(__name__)

allowed = functools.partial(doorman.allowed, DOORMAN_SERVER, API_AUDIENCE)


@app.errorhandler(doorman.AuthZError)
def handle_auth_error(ex):
    response = jsonify(ex.error)
    response.status_code = ex.status_code
    return response


def authorized(**allowed_kw):
    def wrapped(f):
        @functools.wraps(f)
        def wrapper(*args, **kwargs):
            jwt = request.headers.get("Authorization", None)
            payload = allowed(jwt=jwt, **allowed_kw)
            _app_ctx_stack.top.current_user = payload
            return f(*args, **kwargs)
        return wrapper
    return wrapped


@app.route("/")
@cross_origin(headers=["Content-Type", "Authorization"])
@cross_origin(headers=["Access-Control-Allow-Origin", "*"])
@authorized(resource="hello", action="read")
def hello():
    """A valid access token is required to access this route
    """
    top = _app_ctx_stack.top
    return jsonify(top.current_user)


@app.route("/record/<record_id>", methods=('GET', 'PUT'))
@cross_origin(headers=["Content-Type", "Authorization"])
@cross_origin(headers=["Access-Control-Allow-Origin", "*"])
def record(record_id):
    jwt = request.headers.get("Authorization", None)
    filename = os.path.join(HERE, "records", "{record_id}.json".format(
        record_id=os.path.basename(record_id)))
    author = None
    record = None
    if os.path.exists(filename):
        with open(filename, 'r') as f:
            record = json.load(f)
            author = record["author"]
    if request.method == "GET":
        allowed(resource="record", action="read", jwt=jwt, context={"author": author})
        return jsonify(record['body'])

    elif request.method == "PUT":
        if record is not None:
            payload = allowed(resource="record", action="create", jwt=jwt)
        else:
            payload = allowed(resource="record", action="update", jwt=jwt,
                              context={"author": author})

        with open(filename, 'w') as f:
            email_principal = payload["principals"][1]
            body = {'body': request.get_json(), 'author': email_principal}
            json.dump(body, f)
        return jsonify(body)


if __name__ == "__main__":
    print("DOORMAN_SERVER", DOORMAN_SERVER)
    print("API_AUDIENCE", API_AUDIENCE)
    app.run(host="0.0.0.0", port=os.getenv("PORT", 8000))
