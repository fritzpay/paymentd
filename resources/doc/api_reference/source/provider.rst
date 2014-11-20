Provider API
============

Available providers
------------------- 

.. http:get:: /v1/provider

    Retrieve a list all available providers.

    **Example request**:

    .. sourcecode:: http

        GET /v1/provider HTTP/1.1
        Host: example.com
        Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...
        Accept: application/json

    **Example response**:

    .. sourcecode:: http

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "Status": "success",
            "Info": "providers found",
            "Response": [
                {
                    "Name": "fritzpay"
                }
            ],
            "Error": null
        }

    :statuscode 200: No error, provider data served.
    :statuscode 401: Unauthorized, either the username does not exist or the credentials were incorrect.

    :reqheader Authorization: A valid authorization token.

    .. note:: 
    
      This response is just an example, usually it is much longer!

Check provider
--------------

.. http:get:: /v1/provider/(provider)

    Check if a specific provider is available in the system.

    **Example request**:

    .. sourcecode:: http

        GET /v1/provider/fritzpay HTTP/1.1
        Host: example.com
        Authorization: MTQxNTA5NTI5MHxYaCVyOkp7RNaMujhp...
        Accept: application/json

    :param provider: string

    **Example response**:

    .. sourcecode:: http

        HTTP/1.1 200 OK
        Content-Type: application/json

        {
            "Status": "success",
            "Info": "provider fritzpay found",
            "Response": [
                {
                    "Name": "fritzpay"
                }
            ],
            "Error": null
        }

    :statuscode 200: No error, provider data served.
    :statuscode 401: Unauthorized, either the username does not exist or the credentials were incorrect.
    :statuscode 404: provider not found

    :reqheader Authorization: A valid authorization token.