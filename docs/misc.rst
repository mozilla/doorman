Misc
====

.. _misc-run-source:

Run from source
---------------

.. code-block:: bash

    make serve -e "POLICIES=sample.yaml /etc/doorman"


Run tests
---------

.. code-block:: bash

    make test


Generate API docs
-----------------

We use `Sphinx <http://www.sphinx-doc.org>`_, therefore the Python ``virtualenv`` command is required.

.. code-block:: bash

    make docs


Build docker container
----------------------

.. code-block:: bash

    make docker-build


Advanced settings
-----------------

* ``PORT``: listen (default: ``8080``)
* ``GIN_MODE``: server mode (``release`` or default ``debug``)
* ``LOG_LEVEL``: logging level (``fatal|error|warn|info|debug``, default: ``info`` with ``GIN_MODE=release`` else ``debug``)
* ``VERSION_FILE``: location of JSON file with version information (default: ``./version.json``)


Frequently Asked Questions
--------------------------

Why did you do this like that?
''''''''''''''''''''''''''''''

If something puzzles you, looks bad, or is not crystal clear, we would really appreciate your feedback! Please `file an issue <https://github.com/mozilla/doorman/issues>`_! — yes, even if you feel uncertain :)


Why should I use Doorman?
'''''''''''''''''''''''''

*Doorman* saves you the burden of implementing a fined-grained permission system into your service. Plus, it can centralize and track authorizations accross multiple services, which makes permissions management a lot easier.


How is it different than OpenID servers (like Hydra, etc.)?
'''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

*Doorman* is not responsible of managing users. It relies on an Identity Provider to authenticate requests and focuses on authorization.


What is the difference with my Identity Provider authorizations?
''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

Identity Providers may have some authorization/permissions system that allow to restrict access using user groups, audience and scopes.

This kind of access control is global for the whole service. *Doorman* provides advanced policies rules that can be matched per action, resource, or any domain specific context.


Why YAML?
'''''''''

Policies files are meant to be edited or at least reviewed by humans. And YAML is relatively human-friendly.
Plus, YAML allows to add comments.


Glossary
--------

.. glossary::

    Identity Provider
        An identity provider (abbreviated IdP) is a service in charge of managing identity information, and providing authentication endpoints (login forms, tokens manipulation etc.)

    Access Token
    Access Tokens
        An access token is an opaque string that is issued by the Identity Provider.

    ID Token
    ID Tokens
        The ID token is a JSON Web Token (JWT) that contains user profile information (like the user's name, email, and so forth), represented in the form of claims.

    Principal
    Principals
        In *Doorman*, the *principals* is the list of strings that characterize a user. It is built from the user information, tags from the policies file and roles from the authorization request. (see `Wikipedia <https://en.wikipedia.org/wiki/Principal_(computer_security)>`_)

