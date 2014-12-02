Getting Started With :term:`paymentd`
=====================================

This section assumes that :term:`paymentd` is :ref:`installed <install>`.

Configuration
-------------

:term:`paymentd` is configured using a JSON-file. 

The ``paymentd`` command will look for the environment variable ``PAYMENTDCFG``, which
should contain the path to the config JSON file. Alternatively ``paymentd`` can be invoked
with the ``-c`` flag and the path to the config JSON file::

	$ $GOPATH/bin/paymentd -c /path/to/paymentd.config.json

:term:`paymentd` comes with a default configuration which can be displayed using 
the ``paymentdctl`` tool.

Values not present in the configuration will use the default values.

.. topic:: The default configuration

	::

		$ $GOPATH/bin/paymentdctl cfg w -o /dev/stdout 
		no config file flag provided. will use default config...
		{
		  "Payment": {
		    "PaymentIDEncPrime": 982450871,
		    "PaymentIDEncXOR": 123456789
		  },
		  "Database": {
		    "TransactionMaxRetries": 5,
		    "MaxOpenConns": 10,
		    "MaxIdleConns": 5,
		    "Principal": {
		      "Write": {
		        "mysql": "paymentd@tcp(localhost:3306)/fritzpay_principal?charset=utf8mb4\u0026parseTime=true\u0026loc=UTC\u0026timeout=1m\u0026wait_timeout=30\u0026interactive_timeout=30\u0026time_zone=%22%2B00%3A00%22"
		      },
		      "ReadOnly": null
		    },
		    "Payment": {
		      "Write": {
		        "mysql": "paymentd@tcp(localhost:3306)/fritzpay_payment?charset=utf8mb4\u0026parseTime=true\u0026loc=UTC\u0026timeout=1m\u0026wait_timeout=30\u0026interactive_timeout=30\u0026time_zone=%22%2B00%3A00%22"
		      },
		      "ReadOnly": null
		    }
		  },
		  "API": {
		    "Active": true,
		    "Service": {
		      "Address": ":8080",
		      "ReadTimeout": "10s",
		      "WriteTimeout": "10s",
		      "MaxHeaderBytes": 0
		    },
		    "Timeout": "5s",
		    "ServeAdmin": false,
		    "Secure": false,
		    "Cookie": {
		      "AllowCookieAuth": false,
		      "HTTPOnly": true
		    },
		    "AdminGUIPubWWWDir": "",
		    "AuthKeys": []
		  },
		  "Web": {
		    "Active": false,
		    "URL": "http://localhost:8443",
		    "Service": {
		      "Address": ":8443",
		      "ReadTimeout": "10s",
		      "WriteTimeout": "10s",
		      "MaxHeaderBytes": 0
		    },
		    "Timeout": "5s",
		    "PubWWWDir": "",
		    "TemplateDir": "",
		    "Secure": false,
		    "Cookie": {
		      "HTTPOnly": true
		    },
		    "AuthKeys": []
		  },
		  "Provider": {
		    "URL": "http://localhost:8443",
		    "ProviderTemplateDir": ""
		  }
		}config file /dev/stdout written.

You can use the ``paymentdctl`` tool to write a configuration template like so::

	$ $GOPATH/bin/paymentdctl cfg w -o /path/to/paymentd.config.json

Configuration sections
----------------------

*******
Payment
*******

.. topic:: The Payment section

	::

		"Payment": {
			"PaymentIDEncPrime": 982450871,
			"PaymentIDEncXOR": 123456789
		}

This section contains values related to payments.

PaymentIDEncPrime
+++++++++++++++++

This is a (large) prime (``int64``), which is used to obfuscate the sequential payment IDs.
This value has to be consistent throughout the whole cluster.

Obfuscation is performed using `Modular multiplicative inverses <http://en.wikipedia.org/wiki/Modular_multiplicative_inverse>`_.

PaymentIDEncXOR
+++++++++++++++

This is an arbitrary ``int64``, which XORs the ModInv of the payment ID.

The pair ``PaymentIDEncPrime`` and ``PaymentIDEncXOR`` is the "secret" which allows
encoding and decoding of payment IDs throughout the cluster.

********
Database
********

.. topic:: The Database section

	::

		"Database": {
			"TransactionMaxRetries": 5,
			"MaxOpenConns": 10,
			"MaxIdleConns": 5,
			"Principal": {
				"Write": {
					"mysql": "paymentd@tcp(localhost:3306)/fritzpay_principal?charset=utf8mb4\u0026parseTime=true\u0026loc=UTC\u0026timeout=1m\u0026wait_timeout=30\u0026interactive_timeout=30\u0026time_zone=%22%2B00%3A00%22"
				},
				"ReadOnly": null
			},
			"Payment": {
				"Write": {
					"mysql": "paymentd@tcp(localhost:3306)/fritzpay_payment?charset=utf8mb4\u0026parseTime=true\u0026loc=UTC\u0026timeout=1m\u0026wait_timeout=30\u0026interactive_timeout=30\u0026time_zone=%22%2B00%3A00%22"
				},
				"ReadOnly": null
			}
		}

The Database section holds values for connecting with the RDBMS (Relational Database
Management System).

:term:`paymentd` operates on two separate databases:

* The principal database.
* The payment database.

Each database can have two modes. A read/write and a read-only mode. A replicated read-only
database can be used to reduce load on the read/write databases.

TransactionMaxRetries
+++++++++++++++++++++

The maximum number of retries on database transactions after which a transaction is 
considered failed.

This usually happens when the database cannot get a lock on a row.

MaxOpenConns
++++++++++++

Each database connection (Principal RW, Principal RO, Payment RW, Payment RO) maintains a
connection pool. This is the maximum number of connections which can be made to the
RDBMS and should match the `max_connections <http://dev.mysql.com/doc/refman/5.5/en/server-system-variables.html#sysvar_max_connections>`_ system variable with a reasonable margin
if other processes are connection to the same RDBMS.

MaxIdleConns
++++++++++++

The connection pools maintain a few open connections to avoid having to reconnect. This
is the maximum number of idle connections allowed.

DSNs
++++

The connection Data Source Names (DSNs) are described at the `MySQL driver library <https://github.com/go-sql-driver/mysql#dsn-data-source-name>`_.

The "Write" DSNs are required. The "ReadOnly" DSNs are optional. If they are ``null``,
only the Read/Write connections will be used.
