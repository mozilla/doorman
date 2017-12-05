# Integration Examples

- [Python / Flask](python/): A Web UI interacts with Auth0 and a Flask API

# Configuration Examples

## Doorman tags

For example, you can use Doorman to maintain a carefully curated list of people who should become "superusers" when they log in to a certain service. This means the service doesn't have to build the functionality to promote and demote superusers.

To do that, we define a tag ``superuser`` along with the principals it applies to in the service configuration. And then in the policies rules, we refer to this tag as the ``tag:superuser`` principal.

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
        effect: allow
```

In the example above, the userid ``maria`` and the members of the ``admins`` group (from Identity Provider) are allowed to perform any action on any resource on the ``https://api.service.org`` service.


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
