Introduction
============

The primary goal of :term:`paymentd` is to provide a layer, which communicates with
multiple :term:`Payment Service Providers (PSPs) <PSP>` through a single API.

:term:`paymentd` therefore will be a part of a bigger billing stack inside your
operations.

Its domain consists of:

* Managing payments.
* Communicating with the :term:`PSPs (Payment Service Providers) <PSP>`.
* Managing payment ledgers.
* Providing an event system for payment-related events.

:term:`paymentd` is not a full billing system. A full billing process depends largely
on the type of business, its already established processes and operating departments.

We would rather see :term:`paymentd` as one service in a hetegoreneous environment, than
trying to solve too many different problems in one monolithic system.

Philosophy
----------

.. topic:: Focus

	:term:`paymentd` focuses on maintaining reliable and secure interaction with
	:term:`PSP` APIs.

	Any data and functionality of :term:`paymentd` will therefore be focused on
	dealing with payments.

.. topic:: Accountability

	Any action, transaction, event and flow of :term:`paymentd` should be accounted
	for. As long as security considerations permit it, every action should be logged
	extensively. The full history of any events should be kept.

	Whenever possible API methods should be idempotent.

	Payment operations requires that the whole flow can be reproduced.

Requirements
------------

A MySQL-compatible RDBMS (Relational Database Management System) is required to store 
transactional data. :term:`paymentd` was developed against `MariaDB`_, but it should work 
with any MySQL-compatible RDBMS.

.. links

.. _MariaDB: https://mariadb.com/
