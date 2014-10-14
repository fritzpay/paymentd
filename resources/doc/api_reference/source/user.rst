User API
========

Authentication and Authorization
--------------------------------

..
	TODO replace Authorization example fields with reasonable example values
	Curently dEFFEFeddedeGGEGMceokr353521234 acts as a placeholder

.. http:get:: /v1/user/credentials/basic

	Receive an authorization token for given basic auth.

	The returned authorization token can be used in subsequent :http:header:`Authorization`
	headers for accessing protected resources.

	**Example request**:

	.. sourcecode:: http

		GET /user/credentials/basic HTTP/1.1
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

.. http:get:: /user/(username)

	Retrieve the current state of the user.

	**Example request**:

	.. sourcecode:: http

		GET /user/root HTTP/1.1
		Host: example.com
		Accept: application/json
		Authorization: dEFFEFeddedeGGEGMceokr353521234

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"ID": "1234",
			"Username": "root",
			"Email": "root@example.com",
			"Created": "2006-01-02T15:04:05Z07:00"
		}

	:param username: The username.

	:reqheader Authorization: A valid authorization token.

	:resjson ID: The user ID as a string-encoded integer.
	:resjson Created: The timestamp when the user was created (:rfc:`3339` timestamp).

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


