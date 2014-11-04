Principal API
=============

Create a new principal
----------------------

.. http:put:: /v1/principal

	Create a new principal resource.

	**Example request**:

	.. sourcecode:: http

		PUT /principal HTTP/1.1
		Host: example.com
		Content-Type: application/json
		Accept: application/json
		Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...

		{
			"Name": "acme_corporation"
		}

	:reqheader Authorization: HTTP Basic Auth

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Version": "1.2",
			"Status": "success",
			"Info": "principal acme_corporation created",
			"Response": {
				"ID": "3",
				"Created": "2014-11-04T09:59:28Z",
				"CreatedBy": "Jane Joe",
				"Name": "acme_corporation",
				"Metadata": null
			},
			"Error": null
		}

	
	:statuscode 200: No error, principal data served.
	:statuscode 400: The request was malformed; the princial data could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 409: principal with given id already exists

Change an existing principal
----------------------------

.. http:post:: /v1/principal

	Change an existing principal.

	**Example request**:

	.. sourcecode:: http

		POST /principal HTTP/1.1
		Host: example.com
		Content-Type: application/json
		Accept: application/json
		Authorization: QxNTA4NTO7MHxYaCVyOkp7RNaMujhpMT...

		{
			"ID":"1",
			"Metadata": {
				"city":"munic"
			}
		}

	**Example reponse**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Version": "1.2",
			"Status": "success",
			"Info": "principal acme_corporation changed",
			"Response": {
				"ID": "1",
				"Created": "2014-11-04T14:07:49Z",
				"CreatedBy": "Dan Done",
				"Name": "acme_corporation",
				"Metadata": {
					"city": "munic"
				},
			"Error": null
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

		GET /principal/acme_corporation  HTTP/1.1
		Host: example.com
		Content-Type: application/json
		Accept: application/json
		Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...

	:param name: The principal name.
	
	**Example response**:

	.. sourcecode:: http
		
		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Version": "1.2",
			"Status": "success",
			"Info": "principal acme_corporation found",
			"Response": {
				"ID":1,
				"Created": "2014-11-04T14:07:49Z",
				"CreatedBy": "Dan Done",
				"Name": "acme_corporation",
				"Metadata": {
					"city": "munic",
				},
			"Error": null
		}

	
	:statuscode 200: No error, principal data served.
	:statuscode 400: The request was malformed; the given princial name could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 404: principal with given name could not be found
