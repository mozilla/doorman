IAM
===

[![Build Status](https://travis-ci.org/leplatrem/iam.svg?branch=master)](https://travis-ci.org/leplatrem/iam)
[![Coverage Status](https://coveralls.io/repos/github/leplatrem/iam/badge.svg?branch=master)](https://coveralls.io/github/leplatrem/iam?branch=master)
[![Go Report](https://goreportcard.com/badge/github.com/leplatrem/iam)](https://goreportcard.com/report/github.com/leplatrem/iam)

## Policies

Policies are defined in YAML file (default ``./policies.yaml``) as follow:

```yaml
  -
    description: One policy to rule them all.
    subjects:
      - users:<[peter|ken]>
      - users:maria
      - groups:admins
    actions:
      - delete
      - <[create|update]>
    effect:
      - allow
    resources:
      - resources:articles:<.*>
      - resources:printer
    conditions:
      remoteIP:
        type: CIDRCondition
        options:
          - cidr: 192.168.0.1/16
```

## Run locally

    make serve

## API

**POST /allowed**

```
    {
      "subject": "users:peter",
      "action" : "delete",
      "resource": "resource:articles:ladon-introduction",
      "context": {
        "remoteIP": "192.168.0.5"
      }
    }

```


## Configuration

Via environment variables:

* ``PORT``: listen (default: ``8080``)
* ``GIN_MODE``: server mode (``release`` or default ``debug``)
* ``VERSION_FILE``: location of JSON file with version information (default: ``./version.json``)
* ``POLICIES_FILE``: location of YAML file with policies (default: ``./policies.yaml``)

> Note: the ``Dockerfile`` contains different default values, suited for production.


## Run tests

    make test

## License

* MPLv2.0
