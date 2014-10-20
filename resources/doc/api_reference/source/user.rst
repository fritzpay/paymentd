User API
========

Authentication and Authorization
--------------------------------

..
	TODO replace Authorization example fields with reasonable example values
	Curently dEFFEFeddedeGGEGMceokr353521234 acts as a placeholder

.. http:get:: /v1/authorization/basic

	Receive an authorization token for given basic auth.

	The returned authorization token can be used in subsequent :http:header:`Authorization`
	headers for accessing protected resources.

	**Example request**:

	.. sourcecode:: http

		GET /v1/authorization/basic HTTP/1.1
		Host: example.com
		Accept: application/json
		Authorization: Basic cm9vdDpyb290
		Content-Type: application/json

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Authorization": "dEFFEFeddedeGGEGMceokr353521234"
		}

	:reqheader Authorization: HTTP Basic Auth

	:resjson Authorization: The authorization token, which can be used in the
	                      :http:header:`Authorization` header for subsequent requests.

	:statuscode 200: No error, credentials accepted.
	:statuscode 400: The request was malformed; the provided fields could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	                 were incorrect.

Informational
-------------

.. http:get:: /v1/user

	Retrieve the current state of the user.

	**Example request**:

	.. sourcecode:: http

		GET /v1/user HTTP/1.1
		Host: example.com
		Authorization: dEFFEFeddedeGGEGMceokr353521234

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: text/plain

		root

	:reqheader Authorization: A valid authorization token.

.. _user-credentials:

Credentials
-----------

.. 
	TODO "will support", update as soon as other methods are available
	like key derivation methods

:term:`paymentd` will support multiple methods for accepting and authenticating
credentials.

Currently the following types are available:

+-----------+-------------------------+
|    Type   |       Description       |
+===========+=========================+
| ``plain`` | Password in plain text. |
+-----------+-------------------------+


