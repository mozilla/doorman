Quickstart
==========

Policies
--------

Policies are defined in YAML files for each consuming service, locally or in remote (private) Github repos, as follow:

.. code-block:: YAML

    service: https://api.service.org
    identityProvider: https://api.auth0.com/
    policies:
      - id: alice-bob-create-keys
        description: Alice and Bob can create keys
        principals:
          - userid:alice
          - userid:bob
        actions:
          - create
        resources:
          - key
        effect: allow
      -
        id: crud-articles
        description: Editors can CRUD articles
        principals:
          - role:editor
        actions:
          - create
          - read
          - delete
          - update
        resources:
          - article
        effect: allow

Save it to ``config/api-policies.yaml`` for example.

Run
---

*Doorman* is available as a Docker image (but can also be :ref:`ran from source <misc-run-source>`).

In order to read the local files from the container, we will mount the local ``config`` folder to ``/config``.
We'll then use ``/config`` as the ``POLICIES`` location.

.. code-block:: bash

    docker run \
      -e POLICIES=/config \
      -v ./config:/config \
      -p 8000:8080 \
      --name doorman \
      mozilla/doorman

*Doorman* is now ready to respond authorization requests on `http://localhost:8080`. See :ref:`API docs <api>`!


Examples
--------

See the `examples folder <https://github.com/mozilla/doorman/tree/master/examples>`_ on Github.
