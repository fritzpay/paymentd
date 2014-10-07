User API
========

.. http:post:: /user/(user_name)/credentials

	Send user credentials to receive an authorization token.

	The returned authorization token can be used in subsequent :http:header:`Authorization`
	headers for accessing protected resources.

	**Example request**:

	.. sourcecode:: http

		POST /user/root/credentials HTTP/1.1
		Host: example.com
		Accept: application/json
		Content-Type: application/json

		{
			"method": "plain",
			"password": "mysecretpassword"
		}

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Authorization": "dEFFEFeddedeGGEGMceokr353521234"
		}

	:param user_name: The username.

	:<json method: The credential method. Please refer to :ref:`user-credentials` for a
	               list of supported methods.

	:>json Authorization: The authorization token, which can be used in the
	                      :http:header:`Authorization` header for subsequent requests.

	:statuscode 200: No error, credentials accepted.
	:statuscode 400: The request was malformed; the provided fields could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	                 were incorrect.

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


