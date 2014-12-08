Tables and Figures
==================

.. contents::
	:local:

The :term:`paymentd` default configuration JSON
-----------------------------------------------

.. startPaymentdDefaultConfigJSON

::

	{
	  "Payment": {
	    "PaymentIDEncPrime": 982450871,
	    "PaymentIDEncXOR": 123456789
	  },
	  "Database": {
	    "TransactionMaxRetries": 5,
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
	}

.. endPaymentdDefaultConfigJSON

An Example :http:header:`Authorization` Container
-------------------------------------------------

.. startPaymentdAuthContainer

::

	MTQxODA0NjQ4NnxHd+v6FLA0tYWGA+l5v6vZ+t06jARTrYq09PhAFJ3PTa2tVd
	IFl3AbbGRQbi08isTe8CIOCF8DbsV2VX/iH/7OpikAsDW84Azjlb1MZ2GD+hvQ
	azEdtkuqbgG8oya8T+WofA==

.. endPaymentdAuthContainer
