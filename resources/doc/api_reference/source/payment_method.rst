Payment Method API
==================

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

    :statuscode 200: No error, payment method data served.
    :statuscode 400: The request was malformed; the given params could not be understood.
    :statuscode 401: Unauthorized, either the username does not exist or the credentials.
    :statuscode 404: Payment method not available.

    :reqheader Authorization: A valid authorization token.