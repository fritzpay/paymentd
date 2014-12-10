Glossary
========

.. glossary::
	:sorted:

	paymentd
		The :term:`daemon`, which serves the payment service in the FritzPay stack.

		See the high-level description of the ``paymentd`` service: :ref:`paymentd_server`.

	Customer
		"Customer" means the end-user using :term:`paymentd` to make payments.

	daemon
		A program which runs on a server (usually in the background) and provides
		various services.

	www endpoint
		The HTTP service endpoint, which is used by the customer/end-user to access the
		payment and perform the payment.

		See also the high-level description of the Web Server: :ref:`web_server`.

	Order System
		The service connected to :term:`paymentd`. This system will usually handle
		orders, offer primary interfaces for the customer/end-user and broker API
		calls to other subsystems like fulfillment/delivery.

	PSP
		A Payment Service Provider (PSP) offers services to accept online payments.
