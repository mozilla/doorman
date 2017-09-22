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
          cidr: 192.168.0.1/16
```

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


## Run locally

    make serve


## API

### POST /allowed

**Authentication required** (currently hard-coded as Basic Auth ``foo:bar``)

**Request**:

```json
POST /allowed HTTP/1.1
Authorization: Basic Zm9vOmJhcg==
Content-Type: application/json

{
  "subject": "users:peter",
  "action" : "delete",
  "resource": "resource:articles:ladon-introduction",
  "context": {
    "remoteIP": "192.168.0.5"
  }
}

```

**Response**:

```json
HTTP/1.1 200 OK
Content-Length: 17
Content-Type: application/json; charset=utf-8
Date: Fri, 22 Sep 2017 09:29:49 GMT

{
  "allowed": true
}
```


### More

* [See full API spec](./utilities/openapi.yml)


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
