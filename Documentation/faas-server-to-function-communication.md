# Communication File format between faas server and function

Information is passed between kontain faas server and function by using \<function\>.request and \<function\>.response.
Request and response files are located in /kontain directory.

Request and response files are in plain text. All information is passed as \<key\>: \<value\> pairs. key is always plain text. One KV per line.

## Value encoding

numeric is decimal.
string type value is base64 encoded.
key value pair is encoded as comma separated base64 values. 
key with multiple values is encoded as comma separated base64 values with first as key and rest as values.

## KEYS supported

Following are the supported KEYs in

### Request file

- METHOD
- URL
- HEADER
- DATA

### Response file

- STATUSCODE
- DATA

