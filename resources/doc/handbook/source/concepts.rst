Basic concepts
==============

This section explains basic concepts of :term:`paymentd`, a description of the 
high-level architecture and the reasoning behind our choices.

.. _principal:

The Principal
-------------

The Principal is the primary resource under which all other resources are organized.

.. _project:

A Project
---------

A project is the primary means of organizing payments.

.. _provider:

A Provider
----------

A Payment Service Provider (:term:`PSP`) offers services to accept online payments.

:term:`paymentd` has drivers, which manage the communication and handling of payment-related
events with the :term:`PSPs <PSP>`. Sometimes this documentation will refer to the
implementation of :term:`PSP` connections with "driver".

.. _payment_method:

A Payment Method
----------------

Some :term:`PSPs <PSP>` act as aggregators and have support payment methods. :term:`paymentd` sees
these payment methods as configuration sets on :term:`PSP` drivers.

Provider drivers may implement their own configuration sets or ways to handle payments.
Those will be represented by Payment Methods. Some examples of ways Provider Drivers could
support multiple methods:

* Aggregator PSPs, i.e. PSPs, which support multiple types of payments (Credit Card, 
  Debit, Prepaid, etc.) Each type would be represented by a payment method.
* Different modes of payment such as: direct payment, auth/capture, deferred, etc.

.. _metadata:

The Metadata
------------

Any resource in :term:`paymentd` can hold generic metadata, which is in essence a versioned
key-value store.

This metadata can be accessed from other services to share information on these resources.

Possible use cases for the metadata system are:

* Keeping order information to communicate between :term:`order system` => :term:`paymentd` => fulfillment.
* Keep information to be used by various fraud prevention services.
