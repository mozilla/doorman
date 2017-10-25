IAM
===

[![Build Status](https://travis-ci.org/leplatrem/iam.svg?branch=master)](https://travis-ci.org/leplatrem/iam)
[![Coverage Status](https://coveralls.io/repos/github/leplatrem/iam/badge.svg?branch=master)](https://coveralls.io/github/leplatrem/iam?branch=master)
[![Go Report](https://goreportcard.com/badge/github.com/leplatrem/iam)](https://goreportcard.com/report/github.com/leplatrem/iam)

IAM is an **authorization micro-service** that allows to checks if an arbitrary subject is allowed to perform an action on a resource, based on a set of rules (policies). It is inspired by [AWS IAM Policies](http://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html).

## Policies

Policies are defined in YAML file (default ``./policies.yaml``) as follow:

```yaml
  -
    description: One policy to rule them all.
    subjects:
      - maria
      - <[peter|ken]>
      - group:admin
    actions:
      - delete
      - <[create|update]>
    resources:
      - resources:articles:<.*>
      - resources:printer
    conditions:
      remoteIP:
        type: CIDRCondition
        options:
          cidr: 192.168.0.1/16
    effect: allow
```

Use `effect: deny` to deny explicitly.

Otherwise, requests that don't match any rule are denied.

Regular expressions begin with ``<`` and end with ``>``.

### Conditions

The conditions are **optional** and are used to match field values from the requested context.
There are several ``type``s of conditions:

**Field comparison**

* type: ``StringEqualCondition``

For example, match ``request.context["country"] == "catalunya"``:

```yaml
conditions:
  country:
    type: StringEqualCondition
    options:
      equals: catalunya
```

**Field pattern**

* type: ``StringMatchCondition``

For example, match ``request.context["bucket"] ~= "blocklists-.*"``:

```yaml
conditions:
  bucket:
    type: StringMatchCondition
    options:
      matches: blocklists-.*
```

**Subject comparison**

* type: ``EqualsSubjectCondition``

For example, allow requests where ``request.context["owner"] == request.subject``:

```yaml
conditions:
  owner:
    type: EqualsSubjectCondition
```

**IP/Range**

* type: ``CIDRCondition``

For example, match ``request.context["clientIP"]`` with [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation):

```yaml
conditions:
  clientIP:
    type: CIDRCondition
    options:
      # mask 255.255.0.0
      cidr: 192.168.0.1/16
```

## API

### POST /allowed

Is this ``subject`` allowed to perform this ``action`` on this ``resource`` in this ``context``?

**Requires authentication**

A valid JSON Web Token (JWT) must be provided in the ``Authorization`` header.
The JWT subject is used to match the policies.

**Request**:

```HTTP
POST /allowed HTTP/1.1
Content-Type: application/json
Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbG...9USXpOalEzUXpV

{
  "action" : "delete",
  "resource": "resource:articles:ladon-introduction",
  "context": {
    "remoteIP": "192.168.0.5"
  }
}

```

**Response**:

```HTTP
HTTP/1.1 200 OK
Content-Type: application/json

{
  "allowed": true,
  "policy": {
    "id": "1",
    "description": "One policy to rule them all."
  },
  "user": {
    "id": "google-auth|2664978978"
  }
}
```


### More

* [See full API spec](./utilities/openapi.yaml)


## Configuration

Via environment variables:

* ``PORT``: listen (default: ``8080``)
* ``GIN_MODE``: server mode (``release`` or default ``debug``)
* ``LOG_LEVEL``: logging level (``fatal|error|warn|info|debug``, default: ``info`` with ``GIN_MODE=release`` else ``debug``)
* ``VERSION_FILE``: location of JSON file with version information (default: ``./version.json``)
* ``POLICIES_FILE``: location of YAML file with policies (default: ``./policies.yaml``)
* ``JWT_ISSUER``:  issuer of the JWT tokens to match. For JWTs issued by Auth0, use the domain with a `https://` prefix and a trailing `/` (eg. `https://auth.mozilla.auth0.com/`)

> Note: the ``Dockerfile`` contains different default values, suited for production.


## Run locally

    make serve

Or with JWT verification enabled:

    make serve -e JWT_ISSUER=https://minimal-demo-iam.auth0.com/


## Run tests

    make test


## License

* MPLv2.0
