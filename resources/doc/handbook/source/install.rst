Installing
==========

This part will guide you through the installation process of :term:`paymentd`.

Dependencies
------------

:term:`paymentd` is very light on dependencies. You will need:

* A current version of `Go`_ (although Go version 1.2 is still supported).
* `git`_ so the ``go`` tool can obtain the sources from git repositories.
* `mercurial`_ so the ``go`` tool can obtain the sources from mercurial repositories.

Building from source
--------------------

:term:`paymentd` is written in `Go`_. This section will guide you through building
the binaries from the sources.

*********************
Obtaining the sources
*********************

You will need a working `Go`_ installation to build :term:`paymentd` from the source
code.

The GitHub project and repository can be found at https://github.com/fritzpay/paymentd

Obtaining the sources the "Go Way" is simple as::

	$ go get github.com/fritzpay/paymentd

The ``go`` tool will download the sources and create a tree like::

	$ tree -d $GOPATH -L 5
	/home/seong/tmp
	├── bin
	└── src
	    └── github.com
	        └── fritzpay
	            └── paymentd
	                ├── cmd
	                ├── Godeps
	                ├── htmlSrc
	                ├── pkg
	                └── resources

.. note::

	You should be familiar with the `$GOPATH`_ concept of `Go`_. The ``go`` tool will
	download the project into your `$GOPATH`_ and create the standard structure.

*********************
Building the binaries
*********************

Building the binaries can be invoked through::

	$ go run $GOPATH/src/github.com/fritzpay/paymentd/make.go

This will build the binaries and run the basic tests.

The ``make.go`` tool supports various flags, which are explained when calling::

	$ go run $GOPATH/src/github.com/fritzpay/paymentd/make.go -h

After a successful build, the compiled binaries can be found under::

	$ ls $GOPATH/bin 
	paymentd  paymentdctl

along with some utilities.

.. links

.. _Go: http:/golang.org
.. _$GOPATH: https://golang.org/doc/code.html#GOPATH
.. _git: http://git-scm.com/
.. _mercurial: http://mercurial.selenic.com/
