The :term:`paymentd` Server
===========================

This section gives a high-level overview of the :term:`paymentd` server and where
it fits in your billing environment.

.. _api_server:

The API Server
--------------

The API Server will listen on its own port. Mutliple :term:`paymentd` processes
can be configured to either serve the API or not.