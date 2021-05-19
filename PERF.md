# FAAS: The Promise

To users, FAAS offerings such as AWS Lambda, Kubeless, Open-FAAS, etc. provide a simple environment where
all the user has to do is define an event and the code to run when the event happens and the system takes
care of the rest. Users are freed from dealing with low level constructs like containers, virtual machines
and horizontal scaling.

# FAAS: The Reality

Under the covers, the user's function is packaged inside a container image and that image is run inside
of virtual machines when the function is called. Since virtual machines and container images are traditionally
heavy weight things to start, the best performance is seen when the user function container image is already
running.

As the total number of user functions grow, so does the number of container images and, at least in a naive
implementation, so does the number of machine resources required to run the images. Of course, the commercial
FAAS providers recognized this and had internal algorithms to shut rarely used functions so they don't use
resources when they are idle and re-start them on-demand.

Since restarting the function means starting a container image, and potentially starting a virtual machine,
the response time for a call to a function that was restarted is orders of magnitude greater that when the
function container image is already running (as in 10's of milliseconds vs seconds or even 10's of seconds).

Users came to realize this and will do things like periodically 'ping' functions to ensure they don't get
shut down. AWS provides 'Provisioned Concurrency for Lambda Functions' which ensures a user's functions
are always loaded and ready to go (for a price).

Clearly having sporadically executed functions 

# FAAS: What Kontain Can Do

Cold start response time = container startup time + user program initialization time

KRUN/KM provide fast container startup

KM Snapshots allow program initialization to occur off-line.

# FAAS: Some numbers

Plan: Show consistent response time for cold functions in Kontain FAAS server.

