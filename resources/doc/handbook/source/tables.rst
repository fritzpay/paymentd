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

paymentd General JSON response fields
-------------------------------------

.. startPaymentdGeneralJSONResponseFields

.. table:: General JSON response fields

	======== ========================================================
	Field    Explanation
	======== ========================================================
	Version  The API version served.
	Status   A defined status field as a string.
	Info     A human readable explanation about the response.
	Response A response object, which is defined by the request type.
	Error    A generic value. It's ``null`` if no error occured.
	         *Deprecated in version 1.2.*
	======== ========================================================

.. endPaymentdGeneralJSONResponseFields

.. _paymentd-table-statuses:

paymentd API response status codes
----------------------------------

.. tabularcolumns:: |p{5cm}|L|
.. table:: A list of JSON response statuses currently in use.

	======================= ==============================================================
	Status                  Meaning
	======================= ==============================================================
	``success``             The request was successfully processed.
	``error``               There was an error processing the request.
	``unauthorized``        The request could not be processed due to wrong credentials or
	                        missing rights.
	``implementationError`` There was an error mostly due to wrongly formatted request
	                        fields, missing required fields or conflicts.
	======================= ==============================================================

.. _paymentd-table-payment-status-codes:

paymentd Payment status codes
-----------------------------

.. startPaymentStatusCodes

.. tabularcolumns:: |p{5cm}|L|
.. table:: A list payment statuses currently in use.

	+----------------+----------------------------------------------------------------------+
	|     Status     |                               Meaning                                |
	+================+======================================================================+
	| ``open``       | The Payment was accessed by the customer/end-user and is ready to be |
	|                | processed.                                                           |
	+----------------+----------------------------------------------------------------------+
	| ``paid``       | The Payment was succesfully paid.                                    |
	+----------------+----------------------------------------------------------------------+
	| ``cancelled``  | The customer/end-user deliberately cancelled the Payment.            |
	+----------------+----------------------------------------------------------------------+
	| ``chargeback`` | There was a chargeback and the payment was reversed.                 |
	|                |                                                                      |
	|                | This usually happens when an account does not have the required      |
	|                | funds.                                                               |
	+----------------+----------------------------------------------------------------------+

.. endPaymentStatusCodes
