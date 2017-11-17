Doorman
=======

[![Build Status](https://travis-ci.org/leplatrem/iam.svg?branch=master)](https://travis-ci.org/leplatrem/iam)
[![Coverage Status](https://coveralls.io/repos/github/leplatrem/iam/badge.svg?branch=master)](https://coveralls.io/github/leplatrem/iam?branch=master)
[![Go Report](https://goreportcard.com/badge/github.com/leplatrem/iam)](https://goreportcard.com/report/github.com/leplatrem/iam)

*Doorman* is an **authorization micro-service**.

* [Documentation](https://leplatrem.github.io/iam/)

## Run

    docker run mozilla/doorman

## Policies

Policies are defined in YAML files for each service, locally or in remote Github repos, as follow:

```yaml
audience: https://service.stage.net
tags:
  superusers:
    - userid:maria
    - group:admins
policies:
  -
    description: Authors and superusers can delete articles
    principals:
      - role:author
      - tag:superusers
    actions:
      - delete
    resources:
      - article
    effect: allow
```

* ``audience``: the unique identifier of the service
* ``tags``: Local «groups» of principals in addition to the ones provided by the Identity Provider
* ``effect``: Use `effect: deny` to deny explicitly. Otherwise, requests that don't match any rule are denied.

### Subjects

Supported prefixes:

* ``userid:``: provided by Identity Provider (IdP)
* ``tag:``: local tags
* ``role:``: provided in context of authorization request (see below)
* ``email:``: provided by IdP
* ``group:``: provided by IdP

## Settings

Via environment variables:

* ``POLICIES``: space separated locations of YAML files with policies. They can be single files, folders or Github URLs (default: ``./policies.yaml``)
* ``JWT_ISSUER``:  issuer of the JWT tokens to match. For JWTs issued by Auth0, use the domain with a `https://` prefix and a trailing `/` (eg. `https://auth.mozilla.auth0.com/`)
* ``GITHUB_TOKEN``: Github API token to be used when fetching policies files from private repositories

Advanced:

* ``PORT``: listen (default: ``8080``)
* ``GIN_MODE``: server mode (``release`` or default ``debug``)
* ``LOG_LEVEL``: logging level (``fatal|error|warn|info|debug``, default: ``info`` with ``GIN_MODE=release`` else ``debug``)
* ``VERSION_FILE``: location of JSON file with version information (default: ``./version.json``)

> Note: the ``Dockerfile`` contains different default values, suited for production.

## Advanced policies rules

### Regular expressions

Regular expressions begin with ``<`` and end with ``>``.

```yaml
principals:
  - userid:<[peter|ken]>
resources:
  - /page/<.*>
```

> Note: regular expressions are not supported in tags members definitions.

### Conditions

The conditions are **optional** on policies and are used to match field values from the authorization request context.

The context values ``remoteIP`` and ``audience`` are forced by the server.

For example:

```yaml
policies:
  -
    description: Allow everything from dev environment
    conditions:
      env:
        type: StringEqualCondition
        options:
          equals: dev
```

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

For example, match ``request.context["remoteIP"]`` with [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation):

```yaml
conditions:
  remoteIP:
    type: CIDRCondition
    options:
      # mask 255.255.0.0
      cidr: 192.168.0.1/16
```

## Run from source

    make serve

Or with JWT verification enabled:

    make serve -e JWT_ISSUER=https://minimal-demo-iam.auth0.com/

## Run tests

    make test

## Generate API docs

    make api-docs

## Build docker container

    make docker-build

## License

* MPLv2.0
