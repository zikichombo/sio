# Callbacks connecting Go and C

this document contains some (still buggy) thoughts on using atomic to
synchronise hardware sound buffers with go slices/libsio.Packet in callback
based sound APIs.

Callback C APIs are very common in sound and very problematic for Go because
they often run on dedicated threads which cause cgo->go callback overhead:  a
lot on first call, and it seems atleast, sigaltstack on all calls, which
invokes a system call and hence is inappropriate.

Go is a nice fit for providing a blocking API.  We want to map a blocking api
for example a call to 

```
var d []float64
// allocate d
s, _ := sio.Capture()
s.Receive(d)
```

or a call to
```
var d []float64
// fill d
s, _ := sio.Player()
s.Send(d)
```

to be implemented in terms of a callback C API, roughly
```
int cb(void *out, void *in, ...)
```

where cb is called, and presumably run on a foreign thread in a way
that is clocked to synchronize with the number of frames
in `d` and the target hardware and
1. requests the caller to copy in for subsequent treatment of captured data; or
1. requests the caller to fill out for subsequent playback; or
1. requests the caller to do both.


In sio, we have assumed the user may only specify the desired buffer size of
`d` in frames.  In practice, behind this, there is some ringbuffer which `in`
or `out` points to, and the other part of the ringbuffer being filled or
coordinated with the hardware or lower level API.  However, sometimes
the ringbuffer isn't exposed and instead the callbacks are only passed 
pointers to unspecified memory (as in AAudio).

Depending on the implemented mechanism, the latency between the end of the
callback and the respective input or output may vary.  For example, if the
underlying API uses double buffering, the size of the underlying ringbuffer
would would be 2 times the size of `d`.  In a less synchronized context, such
as Apple's audio queue services, the size might be 3 times the size of `d` so
that one buffer may be in transit while both sides treats the other 2 buffers
in parallel.  Many APIs, such as ALSA and AAudio allow to set the buffer size,
which changes the lower and upper bounds on latency between callbacks and
physical sound i/o.


# Design characteristics

To deal with this, we present a blocking call mechanism which synchroniss
with C callbacks via atomics.  The blocking caller must meet the following
requirements
1. The caller must always provide buffers of a specified size.
1. To avoid glitching, the caller must call the associated i/o function
before the associated deadline of the first sample/frame to be exchanged.
1. In order to work on all lower level systems, the call must finish before 
the duration corresponding to that represented by the sound buffer finishes.
This last bit is what cb is optimised for.  Some underlying systems 
may be more lax in this regard with respect to isolated calls. However,
in this case, the rule above still holds on average, and in this case
there is more latency and variation in latency between I/O calls and physical I/O.


### OS Scheduling
In the following, we consider reliability w.r.t. OS scheduling overhead,
cgo<->Go, and Go GC.  The latter looks good, currently at 1-2ms pauses for
large heaps in arbitrary programs and probably less for programs designed
to avoid gc overhead.

Dealing with OS scheduling and cgo<->go  is the whole reason for designing and
implementing this. 

We propose a mechanism in which there are 2 threads, one created and managed
by the underlying C API and one in Go.  The two are coordinated by atomics,
so as to avoid syscall based coordination.  Due to the nature of Go atomics,
it must be the case that the C API only every invokes one callback at a time,
so that each invocation occurs on one thread.

The Go level should be put at the same OS scheduling priority as the C thread.

## Drawbacks
- this mechanism will involve a context switch on single core systems.  It is
recommended to use it on multi-core systems.
- this mechanism uses more threads to accomplish the same thing.
- it may be the case that some real time timing information needs to be
coordinated, as the API itself doesn't guarantee that a blocking call will
be invoked or return in time.  It could also be that something kills the
Go code and the C api just waits for a response forever in a callback.
this may be necessary for detecting xruns as well.

## Advantages
- no syscalls or runtime coordination in cgo->go callbacks.


# Discussion
## Latency
To discuss latency, we focus on the latency of the Go code with respect
to that of the underlying API.  In the following "latency" refers to
this _additional_ latency.

If OS scheduling considerations are respected and the system is multi-core
and the Go and C threads are on different processors or cpu cores, and
the underlying API is capable of supplying reguar buffer sizes, then
there is no latency introduced by this design, as the Go code is called 
in the same real time slot as the callback and there is no buffering between
the two.  This is true of capture, playback, and duplex operation.

When the considerations above are not met, there are the following 
considerations
1. If the underlying API cannot regularly supply desired buffer sizes,
then some data will be buffered in capture.  In playback, the underyling
system may need a capacity for buffers exceeding twice the configured
buffer size, introducing latency.
1. If there is only one core or the threads are scheduled on the same 
core, then a context switch will be necessary.  TBD: consider incorporating
sched_yield when this is the case.
1. If the threads are at different OS priorities, this will introduce
unreliability.


## Buffer sizes
Many APIs (eg ALSA, AAudio)  give the user the ability to control something
corresponding to the rb capacity in this document.  (buffer size in ALSA vs
period, AAudio buffer capacity).  

This is complicated and can produce situations where there is a lot of 
computation with no gain in latency.  We simplify the interface so that 
the caller may use larger buffers for more latency and less computation
and smaller buffers for less latency and more computation.  This does
not come at an expense of best worst case latency, but it does mean the
caller cannot choose the variation in latency which is allowed.

Some APIs do not provide a means to set a buffer size, as only
some buffer sizes are supported. This is not true of AAudio nor CoreAudio, 
but it is true of ALSA.  

TBD: make the API friendly for ALSA as well.

