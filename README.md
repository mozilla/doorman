Doorman
=======

![](logo.svg)

*Doorman* is an **authorization micro-service**.

- [API Documentation](https://mozilla.github.io/doorman/)
- [Examples](examples/)

[![Build Status](https://travis-ci.org/mozilla/doorman.svg?branch=master)](https://travis-ci.org/mozilla/doorman)
[![Coverage Status](https://coveralls.io/repos/github/mozilla/doorman/badge.svg?branch=master)](https://coveralls.io/github/mozilla/doorman?branch=master)
[![Go Report](https://goreportcard.com/badge/github.com/mozilla/doorman)](https://goreportcard.com/report/github.com/mozilla/doorman)

## Run

```
    docker run \
      -e POLICIES=/config/policies.yaml \
      -v ./config/:/config \
      -p 8000:8080 \
      --name doorman \
      mozilla/doorman
```

*Doorman* is now ready to respond authorization requests on `http://localhost:8080`. See [API docs](https://mozilla.github.io/doorman/).

## Policies

Policies are defined in YAML files for each service, locally or in remote Github repos, as follow:

```yaml
service: https://service.stage.net
jwtIssuer: https://auth.mozilla.auth0.com/
tags:
  superusers:
    - userid:maria
    - group:admins
policies:
  -
    id: authors-superusers-delete
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

* **service**: the unique identifier of the service
* **jwtIssuer** (*optional*): when the issuer is set, *Doorman* will verify the JSON Web Token provided in the authorization request and extract the Identity Provider information from its payload
* **tags**: Local «groups» of principals in addition to the ones provided by the Identity Provider
* **actions**: a domain-specific string representing an action that will be defined as allowed by a principal (eg. `publish`, `signoff`, …)
* **resources**: a domain-specific string representing a resource. Preferably not a full URL to decouple from service API design (eg. `print:blackwhite:A4`, `category:homepage`, …).
* **effect**: Use `effect: deny` to deny explicitly. Requests that don't match any rule are denied.

### Principals

The principals is a list of prefixed strings to refer to the «user» as the combination of ids, emails, groups, roles…

Supported prefixes:

* ``userid:``: provided by Identity Provider (IdP)
* ``tag:``: local tags
* ``role:``: provided in context of authorization request (see below)
* ``email:``: provided by IdP
* ``group:``: provided by IdP

Example: `["userid:ldap|user", "email:user@corp.com", "group:Employee", "group:Admins", "role:editor"]`

## Settings

Via environment variables:

* ``POLICIES``: space separated locations of YAML files with policies. They can be single files, folders or Github URLs (default: ``./policies.yaml``)
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

The context value ``remoteIP`` is forced by the server.

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

**Match principals**

* type: ``MatchPrincipalsCondition``

For example, allow requests where ``request.context["owner"]`` is in principals:

```yaml
conditions:
  owner:
    type: MatchPrincipalsCondition
```

> Note: This also works when a the context field is list (e.g. list of collaborators).

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

    make serve -e POLICIES=sample.yaml

## Run tests

    make test

## Generate API docs

    make api-docs

## Build docker container

    make docker-build

## License

* MPLv2.0
* The logo was made by Mathieu Leplatre with [Inkscape](https://inkscape.org/)
  and released under [CC0](https://creativecommons.org/share-your-work/public-domain/cc0/)
