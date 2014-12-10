General API Basics
==================

The paymentd API is REST-ful over HTTP using JSON-encoded messages.

Notes on JSON
-------------

When transmitting JSON over HTTP, the ``Content-Type`` Header must be always set to the correct :mimetype:`application/json` value like so::

	Content-Type: application/json

For JSON only the ``Number`` type is defined. As the notion of an integer differs for paymentd, integers *must* be encoded as strings::

	{"Int":"1234"}

The General JSON Response Format
--------------------------------

All JSON responses from paymentd share the same base format (an envelope if you want).

.. include:: /examples.rst
	:start-after: startPaymentdGeneralResponse
	:end-before: endPaymentdGeneralResponse

.. include:: /tables.rst
	:start-after: startPaymentdGeneralJSONResponseFields
	:end-before: endPaymentdGeneralJSONResponseFields

The possible values for the ``Status`` field are listed in the :ref:`paymentd-table-statuses` table.

API Version History
-------------------

****
v1.2
****

* Introduced the ``Version`` field.
* The generic ``Error`` field is deprecated. This field was deemed redundant. Error
  statuses were already covered by the ``Status`` field. Additionally it would
  expand the state matrix considerably. The generic response should be simple and
  unambiguous.
