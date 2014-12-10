.. _paymentd_server:

The :term:`paymentd` Server
===========================

This section gives a high-level overview of the :term:`paymentd` server and where
it fits in your billing environment.

.. _api_server:

The API Server
--------------

API endpoints are served by the API server. :term:`paymentd` does not deal with securing
the transport layer of its servers. Typically you would want to serve :term:`paymentd`
through a TLS proxy.

The Web Server can be configured to serve an administrative GUI (see 
:ref:`config_api_admin_gui_pub_www_dir`).

The API server will listen on its own port. Mutliple :term:`paymentd` processes
can be configured to either serve the API or not.

Please refer to the :ref:`API section <config_api>` for API Server related configuration
variables.

**********************
Administrative Methods
**********************

The API server can be configured to serve administrative endpoints (See :ref:`config_api_serve_admin` for the configuration variable).

The administrative methods deal with configuring and modifying :ref:`Prinicpals <principal>`,
:ref:`Projects <project>` and :ref:`Payment Methods <payment_method>` along with their
:ref:`Metadata <metadata>`.

One possible scenario is to serve the administrative API only on a dedicated instance
which is only accessible in the internal network or via a VPN.

***************
Payment Methods
***************

All API servers serve payment related API methods. Those deal with 
:ref:`Initializing Payments <init_payment>`, :ref:`Capturing payments <capture_payment>`, etc.

Your :term:`Order System` in your billing stack would typically communicate with the
payment API endpoints.

.. _web_server:

The Web Server
--------------

At some point during the payment flow, the :term:`customer <Customer>` will have to interact
with the payment system. Either to provide payment-related information or to be
redirected to the :term:`Payment Service Provider (PSP) <PSP>`.

The web server will listen on its dedicated port and will serve the payment endpoint along
with :ref:`Provider Driver <provider_driver>` endpoints as well as static files.

Please refer to the :ref:`WWW section <config_www>` for Web Server related configuration
variables.