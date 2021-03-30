---
title: Certificate-related errors when trying to access the Application Registry
type: Troubleshooting
---

## Application Registry - No certificate

If you try to access the Application Registry without a client certificate, you get this error:

```
error:1401E410:SSL routines:CONNECT_CR_FINISHED:sslv3 alert handshake failure
```

To access the Application Registry, you need to pass a client certificate with the HTTP request.  
To generate a client certificate, see [this tutorial](#tutorials-get-the-client-certificate). 

## Application Registry - Expired certificate

If you try to access the Application Registry using an expired client certificate, you get this error:

```
error:1401E415:SSL routines:CONNECT_CR_FINISHED:sslv3 alert certificate expired
```

To access the Application Registry, you need to pass a valid client certificate with the request.  
To generate a new client certificate, see [this tutorial](#tutorials-get-the-client-certificate).  

## Application Registry - Invalid subject

If you try to access Application Registry with the wrong certificate, you get this error:

```json
{"code":403,"error":"No valid subject found"}
```

Make sure that your certificate is generated for the Application that you are trying to access.  
To get the certificate subject, run:

```bash
openssl req -noout -subject -in {PATH_TO_CSR_FILE}
```

You get the certificate subject as a response:

```
subject=/OU=OrgUnit/O=Organization/L=Waldorf/ST=Waldorf/C=DE/CN={APPLICATION_NAME}
```

Check that the common name `CN` matches the name of your Application.  
If it does not, generate a new certificate for your Application. 

To generate a new client certificate, see [this tutorial](#tutorials-get-the-client-certificate).
