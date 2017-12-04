# Integration Examples

- [Python / Flask](python/): A Web UI interacts with Auth0 and a Flask API

# Configuration Examples

## Superusers with Doorman tags

We define a tag ``superuser`` and use it in principals in policies rules.

```yaml

    service: https://api.service.org
    tags:
      superuser:
        - userid:maria
        - group:admins
    policies:
      -
        id: super-users
        description: Superusers can do everything
        principals:
          - tag:superuser
        actions:
          - <.*>
        resources:
          - <.*>
```

## Superusers from service

The superusers are managed on the service itself (eg. database). We will rely on a `superuser` role. The service can send user roles in authorization requests context.

- *TODO: add a small Django demo*

```yaml

    service: https://api.service.org
    policies:
      -
        id: super-users
        description: Superusers can do everything
        principals:
          - role:superuser
        actions:
          - <.*>
        resources:
          - <.*>
```

An authorization request sent from the service looks like this:

```HTTP

    POST /allowed HTTP/1.1
    Origin: https://api.service.org

    {
      "principals": ["userid:myself"],
      "action": "delete",
      "resource": "articles/doorman-introduce",
      "context": {
        "roles": ["superuser"]
      }
    }
```

> Note: this implies the service will also verify and read the JWT payload, so it can lookup whether the user is marked as superuser in the database.
