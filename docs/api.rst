.. _api:

API
===

Summary
-------

Basically, authorization requests are checked using **POST /allowed**.

* The ``Origin`` request header specifies the service to match policies from.
* The ``Authorization`` request header provides the OpenID :term:`Access Token` to authenticate the request.

**Request**:

.. code-block:: HTTP

    POST /allowed HTTP/1.1
    Origin: https://api.service.org
    Authorization: Bearer f2457yu86yikhmbh

    {
      "action" : "delete",
      "resource": "articles/doorman-introduce",
    }

**Response**:

.. code-block:: HTTP

    HTTP/1.1 200 OK
    Content-Type: application/json

    {
      "allowed": true,
      "principals": [
        "userid:ada",
        "email:ada.lovelace@eff.org",
        "group:scientists",
        "group:history"
      ]
    }


Principals
----------

The authorization request :term:principals will be built from the user profile information like this:

* ``"sub"``: ``userid:{}``
* ``"email"``: ``email:{}`` (*optional*)
* ``"groups"``: ``group:{}, group:{}, ...`` (*optional*)

They will be matched against those specified in the policies rules to determine if the authorization request is denied or allowed.


Authentication
--------------

*Doorman* relies on OpenID to authenticate requests.

It will use the ``service`` and ``identityProvider`` fields from the service policies file to fetch the user profile information.

The ``Origin`` request header should match one of the services defined in the policies files.

The ``Authorization`` request header should contain a valid :term:`Access Token`, prefixed with ``Bearer ``.
This access token must have been requested with the ``openid profile`` scope for *Doorman* to be able to fetch the profile information (See `Auth0 docs <https://auth0.com/docs/tokens/access-token#access-token-format>`_).

The userinfo URI endpoint is then obtained from the metadata available at ``{identityProvider}/.well-known/openid-configuration``.

If the obtention of user infos is denied by the :term:`Identity Provider`, the authorization request is obviously denied.


Using ID tokens
'''''''''''''''

*Doorman* can verify and read user information from JWT :term:`ID tokens`. Since the ID token payload contains the user information, it saves a roundtrip to the Identity Provider when handling authorization requests.

For this to work, the ``service`` value in the policies file must match the ``audience`` value configured on the Identity Provider — the unique identifier of the target API. For example, in `Auth0 <https://auth0.com>`_ it can look like this: ``SLocf7Sa1ibd5GNJMMqO539g7cKvWBOI``.

.. important::

    When using JWT :term:`ID tokens`, only the validity of the token will be checked. In other words, users that are revoked from the Identity Provider after their ID token was issued will still considered authenticated until the token expires.


Without authentication
''''''''''''''''''''''

If the identity provider is not configured for a service (explicit empty value), no authentication is required and the principals are posted in the authorization body.

.. code-block:: HTTP

    POST /allowed HTTP/1.1
    Origin: https://api.service.org
    Authorization: Bearer f2457yu86yikhmbh

    {
      "action" : "delete",
      "resource": "articles/doorman-introduce",
      "principals": [
        "userid:mickaeljfox",
        "email:mj@fox.com",
        "group:actors"
      ]
    }

It is not especially recommended, but it can give a certain amount of flexibility when authentication is fully managed on the service.

A typical workflow in this case would be:

1. Users call the service API endpoint
1. The service authenticates the user and builds the list of principals
1. The service posts an authorization request on *Doorman* containing the list of principals to check if the user is allowed


.. _api-context:

Context
-------

Authorization requests can carry additional information contain any extra information to be matched in :ref:`policies conditions <policies-conditions>`.

The values provided in the ``roles`` context field will expand the principals with extra ``role:{}`` values.

.. code-block:: HTTP

    POST /allowed HTTP/1.1
    Origin: https://api.service.org
    Authorization: Bearer f2457yu86yikhmbh

    {
      "action" : "delete",
      "resource": "articles/doorman-introduce",
      "context": {
        "env", "stage",
        "roles": ["editor"]
      }
    }


API Endpoints
-------------

(Automatically generated from `the OpenAPI specs <https://github.com/mozilla/doorman/blob/master/api/openapi.yaml>`_)

.. openapi:: ../api/openapi.yaml
