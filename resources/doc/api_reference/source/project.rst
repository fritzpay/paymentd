Project API
===========

Create a new project
--------------------

.. http:put:: /v1/project

	Create a new project

	**Example request**:

	.. sourcecode:: http

		PUT /project HTTP/1.1
		Host: example.com
		Content-Type: application/json
		Accept: application/json
		Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...

		{
			"PrincipalID":"1",
			"Name":"Roadrunnergame",
			"Metadata": {
				"Version":"Singleplay"
			}
		}

	**Example reponse**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
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
				"Metadata":{
					"Version":"Singleplay"
				}
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

		POST /project HTTP/1.1
		Host: example.com
		Content-Type: application/json
		Accept: application/json
		Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...
	
		{
			"PrincipalID":"1",
			"ID":"1",
			"Metadata": {
				"Type": "Game",
				"Version":"1"
			}
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
				"Metadata": {
					"Type": "Game",
					"Version":"1"
				}
			},
			"Error": null
		}

	:statuscode 200: No error, project data changed.
	:statuscode 400: The request was malformed; the provided parameters could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 404: project with given id was not found 

Informational
-------------

.. http:get:: /project/(id)?principalid=(principalid)

	Retrieve the project data with the given project id.

	**Example request**:

	.. sourcecode:: http

		GET /project/1?principalid=1 HTTP/1.1
		Host: example.com
		Accept: application/json
		Authorization: dEFFEFeddedeGGEGMceokr353521234

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Accept: application/json
		Content-Type: application/json

		{
			"Version": "1.2",
			"Status": "success",
			"Info": "project Roadrunnergame found",
			"Response": {
				"ID": "1",
				"PrincipalID": "1",
				"Name": "Roadrunnergame",
				"Created": "2014-10-17T14:12:11Z",
				"CreatedBy": "John Doe",
				"Config": {
					"WebURL": null,
					"CallbackURL": null,
					"CallbackAPIVersion": null,
					"ProjectKey": null,
					"ReturnURL": null
				},
				"Metadata": {
					"Type": "Game",
					"Version": "1"
				}
			},
			"Error": null
		}

	:param name: The project id
	:param name: The principal id
	
	:statuscode 200: No error, project data served.
	:statuscode 400: The request was malformed; the provided id could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 404: project with given id was not found 

