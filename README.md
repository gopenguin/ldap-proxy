ldap-proxy
==========

[![Build Status](https://travis-ci.org/kolleroot/ldap-proxy.svg?branch=master)](https://travis-ci.org/kolleroot/ldap-proxy)

A proxy delegating ldap requests to backends.

The goal of this project is to make other forms of authentication available to
the ldap protocol and combine multiple datasources into one.

Getting Started
---------------

To get started, build the app with `go build` and start the proxy with `./ldap-proxy porxy`. The the proxy loads the configuration by default from a file named `config.json`. You can change the file used for configuration using the flag `--filename <config-file.json>`.

Some examples can be found in [examples](examples/).

Backends
--------

Backends must implement the `pkg.BackendFactory` and must register themselves
with `config.Loader.AddFactory(pkg.BackendFactory)`.

The base configuration has the following keys:
* `name`: The name of the backend. This is only used for display and logging.
* `baseDn`: the base dn for this backend.
* `peopleRdn`: the rdn for users
* `userRdnAttribute`: the rdn attribute of a single user

### in-memory

The *in-memory* backend allows to define users in the configuration file. This
backend should be used to setup application users to authenticate for searching
*real* users

Options:
* `listUsers`: weather to list the users during search
* `users`: a list of all the users
    * `name`: the name of the users
    * `password`: a bcrypt protected password like `$2a$12$ti1w7IG6I1hsyVcv/C2Z9OvX/DnG8ldHYQm1jqfN38q2GtSZW0NvG`

### postgres

The *postgres* backend connects to a database using the postgres protocoll
(postgres or cockroachdb) and uses sql to authenticate and query users.

Options:
* `url`: the url to the postgres server (or cockroach) e. g. `postgres://test:test@localhost:26257/auth?sslmode=disable`
* `columns`: the db columns and their ldap attribute names. The key represents the db column name, the value the ldap attribute name.
