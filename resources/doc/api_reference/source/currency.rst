Currency API
============

Available currencies
-------------------- 

.. http:get:: /v1/currency

	Retrieve a list all available currencies.

	**Example request**:

	.. sourcecode:: http

		GET /v1/currency HTTP/1.1
		Host: example.com
		Authorization: dEFFEFeddedeGGEGMceokr353521234
		Accept: application/json

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		[
			"EUR",
			"RUB",
			"JPY",
			"USD"
		]		

	:reqheader Authorization: A valid authorization token.

   .. note:: 
    
      This response is just an example, usually it is much longer!

Check currency
--------------

.. http:get:: /v1/currency/(currency)

	Check if a specivic currency is available in the system.

	**Example request**:

	.. sourcecode:: http

		GET /v1/currency/EUR HTTP/1.1
		Host: example.com
		Authorization: dEFFEFeddedeGGEGMceokr353521234
		Accept: application/json

	:param currency: string [A-Z]{3}


	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		"EUR"

	:statuscode 200: No error, currency data served.
	:statuscode 400: The request was malformed; the given currency could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 404: currency not available

	:reqheader Authorization: A valid authorization token.