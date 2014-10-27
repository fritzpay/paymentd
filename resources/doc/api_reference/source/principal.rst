Principal API
=============

Create a new principal
----------------------

.. http:put:: /v1/principal

	Crate a new principal

	**Example request**:

	.. sourcecode:: http

		PUT /principal
		Host: example.com
		Accept: application/json
		Authorization: Basic cm9vdDpyb290

	.. sourcecode:: http


		{
        	"Name": "acme_corporation",
			"CreatedBy": "Jane Doe"
		}

	:reqheader Authorization: HTTP Basic Auth

	**Example response**:

		Accept: application/json
		Authorization: Basic cm9vdDpyb290
		Content-Type: application/json

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

	.. sourcecode:: http
		    "Status": "success",
		    "Info": "principal acme_corporation found",
		    "Response": {
					"ID":1,
					"Created":"2014-10-17T14:12:11Z",
					"CreatedBy":"Jane Doe",
					"Name":"acme_corporation",
					"Metadata":null
				},
		    "Error": null

 
	
	:statuscode 200: No error, principal data served.
	:statuscode 400: The request was malformed; the princial data could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 409: principal with given id already exists

Change an existing principal
--------------------------

.. http:put:: /v1/principal

	Change an existing principal

	**Example request**:

	.. sourcecode:: http

		POST /principal
		Host: example.com
		Accept: application/json
		Authorization: Basic cm9vdDpyb290

	.. sourcecode:: http

		{
			"PrincipalID":"1",
			"Name":"DifferentName",
			"CreatedBy":"Dohn Joe"
		}

	**Example reponse**:

	.. sourcecode:: http

		Accept: application/json
		Authorization: Basic cm9vdDpyb290
		HTTP/1.1 200 OK
		Content-Type: application/json

	.. sourcecode:: http

		{
			"ID":1,
			"PrincipalID":"1",
			"Name":"DifferentName",
			"CreatedBy":"John Doe",
			"Created":"2014-10-17T14:12:11Z",
			"Metadata":null
		}

	:reqheader Authorization: HTTP Basic Auth

	:statuscode 200: No error, principal data changed.
	:statuscode 400: The request was malformed; the provided parameters could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 404: principal with given id was not found 

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

	:param name: The principal name.
	:reqheader Authorization: HTTP Basic Auth
	
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

	
	
	:statuscode 200: No error, principal data served.
	:statuscode 400: The request was malformed; the given princial name could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 404: principal with given name could not be found
