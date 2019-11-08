# HTTP Adapter CloudEvents Spec

# Implemented Spec

`cloudevents-batch+json` not implemented

## source
required, but overwritten with custom value

## specversion
only CE 1.0

# datacontenttype
then using structured mode, 
  
# others
type, datacontenttype, dataschema, subject, time are optional
if provided, they should follow the [CE 1.0 spec](https://github.com/cloudevents/spec/blob/v1.0/spec.md)
  
# CE Spec

## id
Type: String
Constraints:
REQUIRED
MUST be a non-empty string
MUST be unique within the scope of the producer

## source
Type: URI-reference
Constraints:
 REQUIRED
 MUST be a non-empty URI-reference
An absolute URI is RECOMMENDED   

## specversion
Type: String
Constraints:
 REQUIRED
 MUST be a non-empty string	

## type
Type: String
Constraints:
 REQUIRED
 MUST be a non-empty string
 SHOULD be prefixed with a reverse-DNS name. The prefixed domain dictates the organization which defines the semantics of this event type.	

The following attributes are optional but still have validation.


## datacontenttype
Type: String per RFC 2046
Constraints:
 OPTIONAL
 If present, MUST adhere to the format specified in RFC 2046
	
## dataschema
Type: URI
Constraints:
 OPTIONAL
 If present, MUST adhere to the format specified in RFC 3986
    	
## subject
Type: String
Constraints:
 OPTIONAL
 MUST be a non-empty string

## time
Type: Timestamp
Constraints:
 OPTIONAL
 If present, MUST adhere to the format specified in RFC 3339
