Administrative API
==================

.. contents::
	:local:

User API
--------

The User API deals with user authentication and authorization. All administrative
API methods require a valid :http:header:`Authorization` header to be present.

Those can be obtained using the User API methods. Other services in your stack could
act as a authentication/authorization service and provide the correct
auth schemes for integrating the administrative API into your environment.

***************************
The Authorization Container
***************************

The Authorization Container is self-contained and holds all required information
for authenticating and authorizing the bearer of the container.

It is base64-encoded and encrypted using the Keys
configured in :ref:`config_api_auth_keys`.

An Authorization Container provided by :term:`paymentd` has a fixed expiry of 
**15 minutes**.

.. note::

	The expiry is subject to change. Later versions will allow configuring
	this variable.

Here is an example authorization container:

.. include:: /examples.rst
	:start-after: startPaymentdAuthContainer
	:end-before: endPaymentdAuthContainer

.. note::

	In the future, the Authorization Container structure will be replaced by
	`Macaroons <http://theory.stanford.edu/~ataly/Papers/macaroons.pdf>`_.

.. _system_user:

***************
The System User
***************

:term:`paymentd` has the notion of a system user. This unique user (in fact this is 
the only "user"-like entity present in the application), has full read/write access
on every aspect of :term:`paymentd`. This system user is similar in concept to the
UNIX ``root`` user.

***********
Cookie Auth
***********

When cookie auth is enabled (:ref:`config_api_cookie_allow_cookie_auth`)
requests will accept the ``auth`` cookie, with the authorization token as a value.

All responses containing a new authorization container will have a matching
:http:header:`Set-Cookie` header.

********************************
Authentication and Authorization
********************************

..
	TODO replace Authorization example fields with reasonable example values
	Curently dEFFEFeddedeGGEGMceokr353521234 acts as a placeholder

.. http:get:: /v1/authorization/basic
	:synopsis: Receive an authorization token for given basic auth.

	Receive an authorization token for given basic auth.

	The password must match the :ref:`system_user` password. The returned authorization
	container will identify the bearer as the :ref:`system_user`.

	The returned authorization token can be used in subsequent :http:header:`Authorization`
	headers for accessing protected resources.

	**Example request**:

	.. sourcecode:: http

		GET /v1/authorization/basic HTTP/1.1
		Host: example.com
		Accept: application/json
		Authorization: Basic cm9vdDpyb290
		Content-Type: application/json

	**Example response**: 

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json
		Set-Cookie: auth=MTQxODA0NjQ4NnxHd+v...; Path=/v1; Expires=Mon, 08 
		            Dec 2014 15:43:00 UTC; HttpOnly

		{
			"Authorization": "MTQxODA0NjQ4NnxHd+v..."
		}

	:reqheader Authorization: HTTP Basic Auth

	:resheader Set-Cookie: Present when :ref:`config_api_cookie_allow_cookie_auth`
	                       is enabled.
	:resjson Authorization: The authorization token, which can be used in the
	                      :http:header:`Authorization` header for subsequent requests.

	:statuscode 200: No error, credentials accepted.
	:statuscode 400: The request was malformed; the provided fields could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials were incorrect.

.. http:post:: /v1/authorization/text
	:synopsis: :http:method:`POST` the password in plaintext to receive an 
	           authorization container.

	:http:method:`POST` the password in plaintext to receive an authorization container.

	The password should be UTF-8 encoded and sent in plaintext in the request body.

	**Example request**:

	.. sourcecode:: http

		POST /v1/authorization/text HTTP/1.1
		Host: example.com
		Content-Type: text/plain; charset=utf-8

		root

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json
		Set-Cookie: auth=MTQxODA0NjQ4NnxHd+v...; Path=/v1; Expires=Mon, 08 
		            Dec 2014 15:43:00 UTC; HttpOnly

		{
			"Authorization": "MTQxODA0NjQ4NnxHd+v..."
		}

	:reqheader Authorization: HTTP Basic Auth

	:resheader Set-Cookie: Present when :ref:`config_api_cookie_allow_cookie_auth`
	                       is enabled.
	:resjson Authorization: The authorization token, which can be used in the
	                      :http:header:`Authorization` header for subsequent requests.

	:statuscode 200: No error, credentials accepted.
	:statuscode 400: The request was malformed; the provided fields could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials were incorrect.

**********************
Renew an authorization
**********************

.. http:get:: /v1/authorization
	:synopsis: Renew an authorization.

	Renew an authorization.

	Passing a valid authorization container will return a new container, extending
	the expiry.

	**Example request**:

	.. sourcecode:: http

		GET /v1/authorization HTTP/1.1
		Host: example.com
		Authorization: MTQxODA0NjQ4NnxHd+v...
		Cookie: auth=MTQxODA0NjQ4NnxHd+v...

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Authorization": "MTQxODA0NjQ4NnxHd+v..."
		}

	:reqheader Authorization: A valid authorization token.
	:reqheader Cookie: Accepted when :ref:`config_api_cookie_allow_cookie_auth`
	                   is enabled.

	:resjson Authorization: The authorization token, which can be used in the
	                      :http:header:`Authorization` header for subsequent requests.

	:statuscode 200: No error, credentials accepted.
	:statuscode 400: The request was malformed; the provided fields could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials were incorrect.

******************
Set a new password
******************

.. http:put:: /v1/authorization
	:synopsis: Set a new system user password.

	Set a new system user password.

	**Example Request**:

	.. sourcecode:: http

		PUT /v1/authorization HTTP/1.1
		Host: example.com
		Authorization: MTQxODA0NjQ4NnxHd+v...
		Cookie: auth=MTQxODA0NjQ4NnxHd+v...
		Content-Type: text/plain

		newpassword

	**Example Response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: text/plain; charset=utf-8

	:reqheader Authorization: A valid authorization token.
	:reqheader Cookie: Accepted when :ref:`config_api_cookie_allow_cookie_auth`
	                   is enabled.

	:statuscode 200: No error, password changed.
	:statuscode 400: The request was malformed; the provided fields could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials were incorrect.


*************************
Retrieve the current user
*************************

.. http:get:: /v1/user

	Retrieve the current state of the user.

	**Example request**:

	.. sourcecode:: http

		GET /v1/user HTTP/1.1
		Host: example.com
		Accept: application/json
		Content-Type: application/json
		Authorization: MTQxODA0NjQ4NnxHd+v...
		Cookie: auth=MTQxODA0NjQ4NnxHd+v...

	**Example response**:

	.. sourcecode:: http

		HTTP/1.1 200 OK
		Content-Type: application/json
	
		{
			"Version": "1.2",
			"Status": "success",
			"Info": "user id",
			"Response": "root",
			"Error": null
		}

	:reqheader Authorization: A valid authorization token.
	:reqheader Cookie: Accepted when :ref:`config_api_cookie_allow_cookie_auth`
	                   is enabled.

Principal API
-------------

These methods deal with the administration of :ref:`Principals <principal>`.

**********************
Create a new principal
**********************

.. http:put:: /v1/principal

	Create a new principal resource.

	**Example request**:

	.. sourcecode:: http

		PUT /principal HTTP/1.1
		Host: example.com
		Content-Type: application/json
		Accept: application/json
		Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...
		Cookie: auth=MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...

		{
			"Name": "acme_corporation",
			"Metadata": {
				"MyMeta": "Value"
			}
		}

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
				"CreatedBy": "Jane Doe",
				"Name": "acme_corporation",
				"Metadata": {
					"MyMeta": "Value"
				}
			},
			"Error": null
		}

	:reqheader Authorization: A valid authorization token.
	:reqheader Cookie: Accepted when :ref:`config_api_cookie_allow_cookie_auth`
	                   is enabled.

	:resjson string Response.ID: The Principal ID
	:resjson string Response.Created: The Created :rfc:`3339` Timestamp
	:resjson string Response.CreatedBy: The user who created the principal
	:resjson string Response.Name: The principal name
	:resjson Object Metadata: The Metadata or ``null``

	:statuscode 200: No error, current principal state returned.
	:statuscode 400: The request was malformed; the princial data could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	                 were incorrect.
	:statuscode 409: Principal with given name already exists.

*********************
Change principal data
*********************

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
			"Metadata": {
				"city":"munich"
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
				"ID": "3",
				"Created": "2014-11-04T14:07:49Z",
				"CreatedBy": "Dan Done",
				"Name": "acme_corporation",
				"Metadata": {
					"MyMeta": "Value",
					"city": "munich"
				},
			"Error": null
		}

	:reqheader Authorization: A valid authorization token.

	:statuscode 200: No error, principal data changed.
	:statuscode 400: The request was malformed; the provided parameters could not be
	                 understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	                 were incorrect.
	:statuscode 404: Principal with given id was not found 

********************
Retrieve a principal
********************

.. http:get:: /principal/(name)

	Retrieve the given principal.

	**Example request**:

	.. sourcecode:: http

		GET /principal/acme_corporation  HTTP/1.1
		Host: example.com
		Content-Type: application/json
		Accept: application/json
		Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...

	**Example response**:

	.. sourcecode:: http
		
		HTTP/1.1 200 OK
		Content-Type: application/json

		{
			"Version": "1.2",
			"Status": "success",
			"Info": "principal acme_corporation found",
			"Response": {
				"ID": "3",
				"Created": "2014-11-04T14:07:49Z",
				"CreatedBy": "Dan Done",
				"Name": "acme_corporation",
				"Metadata": {
					"MyMeta": "Value",
					"city": "munich"
				},
			"Error": null
		}

	:param name: The principal name.

	:reqheader Authorization: A valid authorization token.
	
	:statuscode 200: No error, principal data served.
	:statuscode 400: The request was malformed; the given princial name could not be understood.
	:statuscode 401: Unauthorized, either the username does not exist or the credentials
	:statuscode 404: principal with given name could not be found
