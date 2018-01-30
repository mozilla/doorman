Policies
========

Policies are defined in YAML files for each consuming service as follow:

.. code-block:: YAML

    service: https://service.stage.net
    identityProvider: https://auth.mozilla.auth0.com/
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

- **service**: the unique identifier of the service
- **identityProvider** (*optional*): when the identify provider is not empty, *Doorman* will verify the Access Token or the ID Token provided in the authorization header to authenticate the request and obtain the subject profile information (*principals*)
- **tags**: Local «groups» of principals in addition to the ones provided by the Identity Provider
- **actions**: a domain-specific string representing an action that will be defined as allowed by a principal (eg. ``publish``, ``signoff``, …)
- **resources**: a domain-specific string representing a resource. Preferably not a full URL to decouple from service API design (eg. `print:blackwhite:A4`, `category:homepage`, …).
- **effect**: Use ``effect: deny`` to deny explicitly. Requests that don't match any rule are denied.


Settings
--------

Policies can be read locally or in remote (private) Github repos.

Settings are set via environment variables:

* ``POLICIES``: space separated locations of YAML files with policies. They can be **single files**, **folders** or **Github URLs** (default: ``./policies.yaml``)
* ``GITHUB_TOKEN``: Github API token to be used when fetching policies files from private repositories

.. note::

  The ``Dockerfile`` contains different default values, suited for production.


Principals
----------

The principals is a list of prefixed strings to refer to the «user» as the combination of ids, emails, groups, roles…

Supported prefixes:

* ``userid:``: provided by Identity Provider (IdP)
* ``tag:``: local tags from policies file
* ``role:``: provided in :ref:`context of authorization requests <api-context>`
* ``email:``: provided by IdP
* ``group:``: provided by IdP

Example: ``["userid:ldap|user", "email:user@corp.com", "group:Employee", "group:Admins", "role:editor"]``


Advanced policies rules
-----------------------

Regular expressions
'''''''''''''''''''

Regular expressions begin with ``<`` and end with ``>``.

.. code-block:: YAML

    principals:
      - userid:<[peter|ken]>
    resources:
      - /page/<.*>

.. note::

    Regular expressions are not supported in tags members definitions.

.. _policies-conditions:

Conditions
''''''''''

The conditions are **optional** on policies and are used to match field values from the :ref:`authorization request context <api-context>`.

The context value ``remoteIP`` is forced by the server.

For example:

.. code-block:: YAML

    policies:
      -
        description: Allow everything from dev environment
        conditions:
          env:
            type: StringEqualCondition
            options:
              equals: dev

There are several types of conditions:

**Field comparison**

* type: ``StringEqualCondition``

For example, match ``request.context["country"] == "catalunya"``:

.. code-block:: YAML

    conditions:
      country:
        type: StringEqualCondition
        options:
          equals: catalunya

**Field pattern**

* type: ``StringMatchCondition``

For example, match ``request.context["bucket"] ~= "blocklists-.*"``:

.. code-block:: YAML

    conditions:
      bucket:
        type: StringMatchCondition
        options:
          matches: blocklists-.*

**Match principals**

* type: ``MatchPrincipalsCondition``

For example, allow requests where ``request.context["owner"]`` is in principals:

.. code-block:: YAML

    conditions:
      owner:
        type: MatchPrincipalsCondition

.. note::

    This also works when a the context field is list (e.g. list of collaborators).

**IP/Range**

* type: ``CIDRCondition``

For example, match ``request.context["remoteIP"]`` with [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation):

.. code-block:: YAML

    conditions:
      remoteIP:
        type: CIDRCondition
        options:
          # mask 255.255.0.0
          cidr: 192.168.0.1/16
