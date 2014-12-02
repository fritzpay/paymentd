Introduction
============

:term:`paymentd` is a program which provides payment related services for your
billing stack.

Its domain consists of:

* Managing payments.
* Communicating with the :term:`PSPs (Payment Service Providers) <PSP>`.
* Managing payment ledgers.
* Providing an event system for payment-related events.

:term:`paymentd` is a focused service and will not include features, which are considered
out of scope. It's designed to work well in a hetegoreneous environment of varying services.

Requirements
------------

A MySQL-compatible RDBMS (Relational Database Management System) is required to store 
transactional data. :term:`paymentd` was developed against `MariaDB`_, but it should work 
with any MySQL-compatible RDBMS.

.. links

.. _MariaDB: https://mariadb.com/
