ldap-proxy
==========

A proxy delegating ldap requests to backends.

The goal of this project is to make other forms of authentication available to
the ldap protocol and combine multiple datasources into one.

Backends
--------

Backends must implement the `pkg.BackendFactory` and must register themselves
with `config.Loader.AddFactory(pkg.BackendFactory)`.

### in-memory

The *in-memory* backend allows to define users in the configuration file. This
backend should be used to setup application users to authenticate for searching
*real* users

### postgres

The *postgres* backend connects to a database using the postgres protocoll
(postgres or cockroachdb) and uses sql to authenticate and query users.
