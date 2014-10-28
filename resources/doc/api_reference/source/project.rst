Project API
===========

Create a new project
--------------------

.. http:put:: /v1/project

	Create a new project

	**Example request**:

	.. sourcecode:: http

		PUT /project
		Host: example.com
		Accept: application/json
		Authorization: Basic cm9vdDpyb290

	.. sourcecode:: http

		{
			"PrincipalID":"1",
			"Name":"Roadrunnergame",
		}

	**Example reponse**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Accept: application/json
		Authorization: Basic cm9vdDpyb290
		Content-Type: application/json

		{
			"Status": "success",
			"Info": "project Roadrunnergame created",
			"Response": {		
				"ID":1,
				"PrincipalID":"1",
				"Name":"Roadrunnergame",
				"CreatedBy":"John Doe",
				"Created":"2014-10-17T14:12:11Z",
				"Metadata":null
			},
			"Error": null
		}

	:reqheader Authorization: HTTP Basic Auth
	
	:statuscode 200: No error, project created.
	:statuscode 400: The request was malformed; the provided fields could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials were incorrect.
 
Change an existing project
--------------------------

.. http:post:: /v1/project

	Change an existing project

	**Example request**:

	.. sourcecode:: http

		POST /project
		Host: example.com
		Accept: application/json
		Authorization: Basic cm9vdDpyb290

	.. sourcecode:: http		
	
		{
			"PrincipalID":"1",
			"Name":"DifferentName",
		}

	**Example reponse**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Accept: application/json
		Authorization: Basic cm9vdDpyb290
		Content-Type: application/json

		{
			"Status": "success",
			"Info": "project Roadrunnergame created",
			"Response": {	
				"ID":1,
				"PrincipalID":"1",
				"Name":"DifferentName",
				"CreatedBy":"John Doe",
				"Created":"2014-10-17T14:12:11Z",
				"Metadata":null
			},
			"Error": null
		}

	:reqheader Authorization: HTTP Basic Auth

	:statuscode 200: No error, project data changed.
	:statuscode 400: The request was malformed; the provided parameters could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 404: project with given id was not found 

Informational
-------------

.. http:get:: /project/(id)

	Retrieve the project data with the given project id.

	**Example request**:

	.. sourcecode:: http

		GET /project/1
		Host: example.com
		Accept: application/json
		Authorization: dEFFEFeddedeGGEGMceokr353521234

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Accept: application/json
		Authorization: Basic cm9vdDpyb290
		Content-Type: application/json

		{
			"Status": "success",
			"Info": "project Roadrunnergame created",
			"Response": {	
				"ID":1,
				"PrincipalID":"1",
				"Name":"Roadrunnergame",
				"CreatedBy":"John Doe",
				"Created":"2014-10-17T14:12:11Z",
				"Metadata":null
			},
			"Error": null
		}

	:param name: The project id

	:reqheader Authorization: HTTP Basic Auth
	
	:statuscode 200: No error, project data served.
	:statuscode 400: The request was malformed; the provided id could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 404: project with given id was not found 

Add a paymentmethod to a project
--------------------------------

.. http:put:: /v1/project/(projectid)/method/

	Add a paymentmethod to a project

	**Example request**:

	.. sourcecode:: http

		PUT /project/1/method/
		Host: example.com
		Accept: application/json
		Authorization: dEFFEFeddedeGGEGMceokr353521234

		{
			"MethodKey":"dummyprovider",
			"ProviderID":"1",
			"Status":"active",
			"Metadata": null
		}

	:param name: The project id
	:param name: The new payment method name
	:param name: The provider id that should be used for the new payment method

	**Example response**

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Accept: application/json
		Authorization: Basic cm9vdDpyb290
		Content-Type: application/json

		{
			"Status": "success",
			"Info": "method dummyprovider created",
			"Response": {	
				"ID":1,
				"ProjectID":"1",
				"ProviderID":"1",
				"MethodKey":"dummyprovider",
				"CreatedBy":"John Doe",
				"Created":"2014-10-17T14:12:11Z",
				"Metadata":null
			},
			"Error": null
		}
