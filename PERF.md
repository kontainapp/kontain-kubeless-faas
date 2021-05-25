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
function container image is already running (as in 10's of milliseconds vs seconds (pre-pulled container image)
or even 10's of seconds (pull image from registry)).

Users came to realize this and will do things like periodically 'ping' functions to ensure they don't get
shut down. AWS provides 'Provisioned Concurrency for Lambda Functions' which ensures a user's functions
are always loaded and ready to go (for a price).

Open source offering for Kubernetes environments like Kubeless and Open-FAAS implement each function as
one or more dedicated pods listening for calls to that function. Each different function has it's own
set of pods. This is fine when all of the functions are constantly being called, but when a function
is called sporadically this wastes resources.

Some open source FAAS platforms like Open-FAAS will take rarely used functions down to zero running pods
(Scale to Zero) and will restart pods for the function on-demand (Cold Start). Other FAAS platforms like
Kubeless punt on the whole problem and require all functions to have at least one running pod.

As evidenced by commercial FAAS Services like AWS, Scale to Zero is important for controlling resource
utilization and is a critical capability for a service providers.

# How Other FAAS Platforms Handle Cold Start

A common technique used in Kubernetes based FAAS systems is to keep a number of 'embryonic' containers
around that can be adapted to whatever function needs to be started. Open-FAAS claims 1-2 sec function
cold-start times using this technique.

# FAAS: What Kontain Can Do

Previous work we did with Java/SpringBoot showed 20ms start-to response time with Kontain Snaphots.
This is fast enough to support a FAAS server where each function call is handled by a ultra-ephemeral
container dedicated to that single function call. When the function call is complete, the container
goes away. If a new call to the same function is made, a new container is used.

From a security point of view, a single container per function call model like this provides excellent
isolation between function calls since the container for any given function call cannot, by definition,
have been compromised by previous calls to that function.

From a resource usage point of view, a single container per function call model like this provides
close to optimal behavior with resources being consumed for a function only when the function is
doing work for a user.

# Kontain FAAS Server - Proof of Concept

Kontain engineering built a Proof of Concept (POC) implementation to test our hypothesis that a single
container per function call model is feasible using KRUN and the Kontain Monitor (KM). The POC was a
GO-based HTTP server (`kontain-faas-server`) that picks a container to use for a function call based
on the HTTP path in the request URL.

The function call container images are stored on a local filesystem in OCI Runtime Bundle
format. When a function call is received and a Runtime Bundle directory for that function
exists, the server calls KRUN to start the function container. When KRUN terminates, the result
of the function is returned to caller via an HTTP response.

The `kontain-faas-server` implementation was stright forward and naive. No performance tuning was
done on it before is was measured.

# How Perfomance Was Analyzed

There were two pieces of performance that were analyzed:

* Compare `kontain-faas-server` against Kubeless for a Python-based user function.

* Find the lower limit of single container per function response time.

The first, comparing with a python based user function, compares The Kubeless `hellowithdata.py` example
against a `kontain-faas-server` running a python function that we wrote.

The second, finding the lower limit for response time, uses a C langauge based function. While we don't
expect users will write their function in C, this allowed us to measure against a function with minimal
overhead.

The test is serially running `curl` multiple time to create a function call request load. Linux `time` measures
In both cases, the target was running under `minikube` on a local developer workstation: `Intel(R) Core(TM) i7-8700 CPU @ 3.20GHz 32GB Memory`


# Results - Python Based User Functions

The following table shows the results of running 10 serial `curl` calls.
Time is in seconds.

| | Kubeless | KM Python | KM Py Snap | KM C |
|- | -------- | -------| -------| - |
| mean  | 0.023300 | 0.118800 | 0.051000 | 0.038000 |
| std   | 0.011166 | 0.012354 | 0.012953 | 0.007746 |
| min   | 0.018000 | 0.112000 | 0.040000 | 0.027000 |
| 25%   | 0.019000 | 0.113000 | 0.042000 | 0.031250 |
| 50%   | 0.019000 | 0.114500 | 0.047000 | 0.040000 |
| 75%   | 0.019000 | 0.118000 | 0.053250 | 0.044000 |
| max   | 0.054000 | 0.153000 | 0.081000 | 0.049000 |

While all off the runs show high tail latencies, all have fairly reasonable behavior
out to the 75-percentile. As expected the KM cases show worse response time than native
Kubeless.

The first KM case is a full cold start with the user function starting from zero. Every call
involves language initialization and module imports. In perspective, less than 200ms to cold
start a container is an impressive result.

The second KM case uses a KM snapshot of the process image from the first KM case. These results show
that `km-faas-server` can process each request which a fresh container with only a small increase in
response time (20 - 25 milliseconds).

The final KM case shows the minimal C function and is used to estimate the best possible case
for `kontain-faas-server`.

# Takeaways

In the real world, user function will do more than simply echo the input. In particular we
expect that the work we measured here will only be a small piece of real user functions and
the difference between 20 ms and 40 ms will be lost in the noise.  Additionally, we haven't
done anything to performance tune `kontain-faas-server`.

The numbers we gathered validate the Kontain's fast container start claims. And we believe
that what we've demonstrated here can be added to existing FAAS offerings or could be the
basis of a Kontain FAAS product.

## Appendix: Commands used

hellowithdata.py
* `kubeless function deploy hellowithdata --runtime python3.7 --from-file hellowithdata.py --handler hellowithdata.handler`
* `kubeless trigger http create hellowithdata --function-name=hellowithdata --path=kubeless/hellowithdata`
* `time curl --data '{"Another": "Echo"}' --header "Host: hellowithdata.192.168.49.2.nip.io" --header "Content-Type:application/json" 192.168.49.2/kubeless/hellowithdata`
