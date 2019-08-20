---
title: Error when trying to access Application Registry
type: Troubleshooting
---

## Application Registry - No certificate

While trying to access Application Registry, you may get an error like this:
```
error:1401E410:SSL routines:CONNECT_CR_FINISHED:sslv3 alert handshake failure
```
It means that you are trying to access Application Registry without a client certificate. 

To access Application Registry you need to pass a client certificate with the request.  
To generate a client certificate, see [this tutorial](#tutorials-get-the-client-certificate). 

## Application Registry - Expired certificate

While trying to access Application Registry, you may also get an error like this:
```
error:1401E415:SSL routines:CONNECT_CR_FINISHED:sslv3 alert certificate expired
```
It means that you are trying to access Application Registry using an expired client certificate.

To access Application Registry you need to pass a valid client certificate with the request.  
To generate a new client certificate, see [this tutorial](#tutorials-get-the-client-certificate).  

## Application Registry - Invalid subject

While trying to access Application Registry, you may get this response:
```
{"code":403,"error":"No valid subject found"}
```
Make sure that your certificate is generated for the Application that you are trying to access.  
If it is not, generate a new certificate for your Application. 

To generate a new client certificate, see [this tutorial](#tutorials-get-the-client-certificate).
