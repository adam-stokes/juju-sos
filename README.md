juju sos
========

a juju plugin for capturing sosreports from deployed juju machines.


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

# copyright

(c) 2014 Adam Stokes <adam.stokes@ubuntu.com>
(c) 2014 Canonical Ltd.

# license

AGPLv3
