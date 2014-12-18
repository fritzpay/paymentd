Payment Method API
==================

Add a paymentmethod to a project
--------------------------------

.. http:put:: /v1/project/(projectid)/method/

	Add a paymentmethod to a project

	**Example request**:

	.. sourcecode:: http

		PUT /project/1/method/ HTTP/1.1
		Host: example.com
		Accept: application/json
		Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...

		{
			"MethodKey":"dummyprovider",
			"ProviderID":"1",
			"Status":"active"
		}


	:param name: The project id
	:param name: The new payment method key
	:param name: The provider id that should be used for the new payment method

	**Example response**

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Version": "1.2",
			"Status": "success",
			"Info": "created with methodkey dummyprovider",
			"Response": {	
				"ID":1,
				"ProjectID":"1",
				"ProviderID":"1",
				"Provider": {
					"ID": 1,
					"Name": "fritzpay"
				},
				"MethodKey": "dummyprovider",
				"Created": "2014-11-11T12:12:29Z",
				"CreatedBy": "John Doe",
				"Status": "inactive",
				"StatusChanged": "2014-11-11T12:12:29Z",
				"StatusCreatedBy": "John Doe",
				"Metadata": null
			},
			"Error": null
		}



Change an existing a paymentmethod
----------------------------------

.. http:post:: /v1/project/(projectid)/method/(methodkey)

	Change the status of an existing a paymentmethod and adding metadata.

	**Example request**:

	.. sourcecode:: http

		POST /project/1/method/dummyprovider HTTP/1.1
		Host: example.com
		Accept: application/json
		Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...

		{
			"Status":"inactive",
			"ProviderID":1,
			"Metadata":{
				"reason":"offline"
			}
		}


	:param name: The project id
	:param name: The payment method key
	:param name: The provider id of the payment method

	**Example response**

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Version": "1.2",
			"Status": "success",
			"Info": "created with methodkey dummyprovider",
			"Response": {	
				"ID":1,
				"ProjectID":"1",
				"ProviderID":"1",
				"Provider": {
					"ID": 1,
					"Name": "fritzpay"
				},
				"MethodKey": "dummyprovider",
				"Created": "2014-11-11T12:12:29Z",
				"CreatedBy": "John Doe",
				"Status": "inactive",
				"StatusChanged": "2014-11-11T13:00:20Z",
				"StatusCreatedBy": "John Doe",
				"Metadata": {
		 		   "reason":"offline"
				}
			},
			"Error": null
		}


Informational
-------------

.. http:get:: /v1/project/(projectid)/method/(methodkey)/provider/(providerid)

	Change the status of an existing a paymentmethod and adding metadata.

	**Example request**:

	.. sourcecode:: http

		GET /project/1/method/dummyprovider/provider/1 HTTP/1.1
		Host: example.com
		Accept: application/json
		Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...



	:param name: The project id
	:param name: The payment method key
	:param name: The provider id of the payment method

	**Example response**

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Version": "1.2",
			"Status": "success",
			"Info": "created with methodkey dummyprovider",
			"Response": {	
				"ID":1,
				"ProjectID":"1",
				"ProviderID":"1",
				"Provider": {
					"ID": 1,
					"Name": "fritzpay"
				},
				"MethodKey": "dummyprovider",
				"Created": "2014-11-11T12:12:29Z",
				"CreatedBy": "John Doe",
				"Status": "inactive",
				"StatusChanged": "2014-11-11T13:00:20Z",
				"StatusCreatedBy": "John Doe",
				"Metadata": {
		 		   "reason":"offline"
				}
			},
			"Error": null
		}



	:statuscode 200: No error, payment method data served.
	:statuscode 400: The request was malformed; the given params could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials.
	:statuscode 404: Payment method not available.

	:reqheader Authorization: A valid authorization token.
