# Faas server URL decode

Faas server decodes the URL and decides which function to run. 

Faas server expects URL in the following format.

/\<FUNCTION\>/URL

Currently faas server does not understand "/" in function name.

Faas server expects executable with \<FUNCTION\> name in directory /kontain

## processing of request by Faas server

Once a request is received, faas server does the following steps

- Faas server writes request file with necessary infomtaion in /kotnain/\<FUNCTION\>.reqeust
- Runs /kontain/\<FUNCTION\>
- Waits for completion of \<FUNCTION\>
- Reads response file /kontain/\<FUNCTION\>.response
- Sets up the response code and response data.


## Testing a function

Place the function executable in /kontain director. Use curl or browser to access the url

>ex: curl -i -X GET http://127.0.0.1:8080/hello/GoodMorning

in the above example __hello__ is the function name.