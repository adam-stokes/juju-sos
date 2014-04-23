juju sos
========

a juju plugin for capturing sosreports from deployed juju machines.

my motivation and reference comes form the  work of
[Casey Marshall's juju-nat](https://github.com/cmars/juju-nat). so if
you see any resemblance thats why :)


## install

```console
$ go get github.com/battlemidget/juju-sos
```

## running

Specific machine (1 in this case)

```console
$ juju sos -d $HOME/sosreports -m 1
```

All machines

```console
$ juju sos -d $HOME/sosreports
```

## todo

* use juju api

* filter sos captures based on services

* unittests

* pass arguments to sosreport for specific capturing options, for example,

```console

$ juju sos -d ~/sosreport -- -b -o juju,maas,nova-compute

```

Would only execute sosreport in batch mode (-b) and only the plugins `juju, maas, nova-compute`.

# copyright

(c) 2014 Adam Stokes <adam.stokes@ubuntu.com>

(c) 2014 Canonical Ltd.

# license

AGPLv3
