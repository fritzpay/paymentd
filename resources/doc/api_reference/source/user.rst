User API
========

.. http:post:: /user/(user_name)/credentials

	Send user credentials to receive an Authorization token.

	**Example request**:

	.. sourcecode:: http

		POST /user/root/credentials HTTP/1.1
		Host: example.com
		Accept: application/json
		Content-Type: application/json

		{
			"type": "password",
			"password": "mysecretpassword"
		}

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Authorization": "dEFFEFeddedeGGEGMceokr353521234"
		}

	:param user_name: the username

	:statuscode 200: no error, credentials accepted
	:statuscode 401: unauthorized, either the username does not exist or the credentials
	                 were incorrect