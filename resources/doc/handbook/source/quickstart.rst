Getting Started With :term:`paymentd`
=====================================

This section assumes that :term:`paymentd` is :ref:`installed <install>`.

Database Schemas
----------------

The schemas can be found in::

	$GOPATH/src/github.com/fritzpay/paymentd/resources/mysql/paymentd.sql

Note that the database names are part of the SQL file. If you want to use
different database names, you need to update the references accordingly.

Configuration
-------------

:term:`paymentd` is configured using a JSON-file. For a detailed description
of the configuration variables please refer to :ref:`config`.

The ``paymentd`` command will look for the environment variable ``PAYMENTDCFG``, which
should contain the path to the config JSON file. Alternatively ``paymentd`` can be invoked
with the ``-c`` flag and the path to the config JSON file::

	$ $GOPATH/bin/paymentd -c /path/to/paymentd.config.json

:term:`paymentd` comes with a default configuration which can be displayed using 
the ``paymentdctl`` tool.

.. topic:: The default configuration

	Obtaining the default configuration with the ``paymentdctl`` tool::

		$ $GOPATH/bin/paymentdctl cfg w -o /dev/stdout 
		no config file flag provided. will use default config...
		{
			... default config output
		}config file /dev/stdout written.

Values not present in the configuration will use the default values.

You can use the ``paymentdctl`` tool to write a configuration template like so::

	$ $GOPATH/bin/paymentdctl cfg w -o /path/to/paymentd.config.json

*************************
The default configuration
*************************

.. include:: tables.rst
	:start-after: startPaymentdDefaultConfigJSON
	:end-before: endPaymentdDefaultConfigJSON

Running :term:`paymentd`
------------------------

The server will start serving when you run the binary::

	$ $GOPATH/bin/paymentd

Or with the config flag::

	$ $GOPATH/bin/paymentd -c /path/to/paymentd.config.json

Or with the environment variable set::

	$ export PAYMENTDCFG=/path/to/paymentd.config.json
	$ $GOPATH/bin/paymentd

Restarting :term:`paymentd`
---------------------------

:term:`paymentd` supports live reloading capabilities without dropping active connections.

Sending a ``USR2`` signal to the running process will spawn a new process and pass the
connection fd(s) to the new process. After the handover is completed, the parent process
will be stopped with a ``TERM`` signal.
