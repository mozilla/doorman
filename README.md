IAM
===

[![Build Status](https://travis-ci.org/leplatrem/iam.svg?branch=master)](https://travis-ci.org/leplatrem/iam)
[![Coverage Status](https://coveralls.io/repos/github/leplatrem/iam/badge.svg?branch=master)](https://coveralls.io/github/leplatrem/iam?branch=master)
[![Go Report](https://goreportcard.com/badge/github.com/leplatrem/iam)](https://goreportcard.com/report/github.com/leplatrem/iam)

IAM is an **authorization micro-service** that allows to checks if an arbitrary subject is allowed to perform an action on a resource, based on a set of rules (policies). It is inspired by [AWS IAM Policies](http://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html).

## Policies

Policies are defined in YAML files (default ``./policies.yaml``) as follow:

```yaml
  audience: https://service.stage.net
  tags:
    admins:
      - userid:maria
  policies:
    -
      description: One policy to rule them all.
      principals:
        - userid:<[peter|ken]>
        - tag:admins
        - group:europe
      actions:
        - delete
        - <[create|update]>
      resources:
        - resources:articles:<.*>
        - resources:printer
      conditions:
        clientIP:
          type: CIDRCondition
          options:
            cidr: 192.168.0.1/16
      effect: allow
```

Use `effect: deny` to deny explicitly.

Otherwise, requests that don't match any rule are denied.

Regular expressions begin with ``<`` and end with ``>``.

### Subjects

Supported prefixes:

* ``userid:``: provided by IdP
* ``tag:``: local tags
* ``role:``: provided in context of authorization request (see below)
* ``email:``: provided by IdP
* ``group:``: provided by IdP/LDAP

## API

[See full API docs](https://leplatrem.github.io/iam/)

Basically, **POST /allowed** to check an authorization request:

**Request**:

```HTTP
POST /allowed HTTP/1.1
Origin: https://api.service.org
Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbG...9USXpOalEzUXpV

{
  "action" : "delete",
  "resource": "resource:articles:ladon-introduction",
  "context": {
    "clientIP": "192.168.0.5",
    "roles": ["editor"]
  }
}
```

**Response**:

```HTTP
HTTP/1.1 200 OK
Content-Type: application/json

{
  "allowed": true,
  "principals": [
    "userid:google-auth|2664978978",
    "email:alex@skynet.corp",
    "role:editor",
    "group:admins"
  ]
}
```

## Configuration

Via environment variables:

* ``POLICIES``: space separated locations of YAML files with policies. They can be single files, folders or Github URLs (default: ``./policies.yaml``)
* ``JWT_ISSUER``:  issuer of the JWT tokens to match. For JWTs issued by Auth0, use the domain with a `https://` prefix and a trailing `/` (eg. `https://auth.mozilla.auth0.com/`)
* ``GITHUB_TOKEN``: Github API token to be used when fetching policies from private repositories

Advanced:

* ``PORT``: listen (default: ``8080``)
* ``GIN_MODE``: server mode (``release`` or default ``debug``)
* ``LOG_LEVEL``: logging level (``fatal|error|warn|info|debug``, default: ``info`` with ``GIN_MODE=release`` else ``debug``)
* ``VERSION_FILE``: location of JSON file with version information (default: ``./version.json``)

> Note: the ``Dockerfile`` contains different default values, suited for production.

## Run locally

    make serve

Or with JWT verification enabled:

    make serve -e JWT_ISSUER=https://minimal-demo-iam.auth0.com/

## Advanced policies rules

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

**In principals**

* type: ``InPrincipalsCondition``

For example, allow requests where ``request.context["owner"]`` is in principals:

```yaml
conditions:
  owner:
    type: InPrincipalsCondition
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

## Run tests

    make test

## Generate API docs

    make api-docs

## Build docker container

    make docker-build

## License

* MPLv2.0
