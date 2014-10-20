Principal API
=============

Create a new principal
----------------------

.. http:put:: /v1/principal

	Crate a new principal

	**Example request**:

		PUT /principal
		Host: example.com
		Accept: application/json
		Authorization: Basic cm9vdDpyb290

	.. sourcecode:: http

		Content-Type: application/json

		{
        	"Name": "acme_corporation",
			"CreatedBy": "Jane Doe"
		}

**Example reponse**:

		Accept: application/json
		Authorization: Basic cm9vdDpyb290

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json


		{
			"ID":1,
			"Created":"2014-10-17T14:12:11Z",
			"CreatedBy":"Jane Doe",
			"Name":"acme_corporation",
			"Metadata":null
		}
 


Informational
-------------

.. http:get:: /principal/(name)

	Retrieve the given principal.

	**Example request**:

	.. sourcecode:: http

		GET /principal/acme_corporation
		Host: example.com
		Accept: application/json
		Authorization: dEFFEFeddedeGGEGMceokr353521234

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"ID":1,
			"Created":"2014-10-17T14:12:11Z",
			"CreatedBy":"Jane Doe",
			"Name":"acme_corporation",
			"Metadata":null
		}

	:param name: The principal name.
	
