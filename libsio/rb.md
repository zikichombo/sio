# Ring Buffer connecting Go and C

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
s, _ := sio.Capture()
s.Receive(d)
```

to be implemented in terms of a callback C API, roughly
```
int cb(void *out, void *in, ...)
```

where cb is called, presumably on a foreign thread in a way
that is clocked to synchronize with the number of frames
in `d` and either 
1. requests the caller to copy buf for subsequent treatment of captured data; or
1. requests the caller to fill buf for subsequent playback; or
1. requests the caller to do both.


In sio, we have assumed the user may only specify the desired buffer size of
`d` in frames.  In practice, behind this, there is some ringbuffer which `buf`
points to with `nF` frames of data available, and the other part of the
ringbuffer being filled or coordinated with the hardware or lower level API.

Depending on the implemented mechanism, the size of the underlying ring buffer
may vary and effect latency.  For example, for double buffering, the size of
the underlying ringbuffer would would be 2 times the size of `d`.  In a less
synchronized context, such as Apple's audio queue services the size might be 3
times the size of `d` so that one buffer may be in transit while both sides
treats the other 2 buffers in parallel.


# Design characteristics

To deal with this, we present a ringbuffer in which the elements are
(conceptually atleast) triples 

```
(out *C.void, in *C.void, pkt *libsio.Packet)
```

These triples maintain the invariant that `pkt` is used exclusively on the Go side
and (`in`, `out`) are C allocated memory provided by the underlying API to `cb` and
passed to the Go side for processing.

The ringbuffer has a write index `wi` and a read index `ri` which must be
coordinated between C and Go.  These indices indicate where the next write or
read will occur.  In this design, `wi` is exclusive to the C side 
and `ri` is exclusive to the Go side.  

There is also a 'size' variable indicating the number of samples between the
`ri` and `wi` (circularly ordered).  The 'size' variable is atomically
synchronised between the two sides and may exceed the capacity, indicating
an xrun.  Neither `ri` nor `wi` ever exceed capacity.

We assume that Go code will encode `pkt` into `out` for playback and encode
`in` into `pkt` for capture and that in a duplex setting the Go code will do
both.  A processing chain may do more, but is outside the scope of sio for the
time being.

There are many combinations worth noting.  In the following, we consider
"reliability" w.r.t. OS scheduling overhead and Go GC.  The latter looks good.
Dealing with the former is the whole reason for designing and implementing
this.  In practice, there is a sweet spot between reliability and latency:  you
can reduce the latency to the level where it is sufficiently reliable for your
use case.  The cases are enumerated below.

1. The ringbuffer is hardware/low level OS priveleged and Go is on a specially
scheduled thread.  In this case, the implementation is the same w.r.t. 
OS scheduling as say AAudio on top of Android HAL or AUHAL on top of CoreAudio's 
Audio Device HAL, or JACK. We would expect equivalent latency reliability modulo Go garbage collection.
1. Same as above, but Go code is not specially OS scheduled for audio.  In this case,
we would expect an increase in glitching under non-dedicated hardware or system load.
1. The ringbuffer or similar is implemented by a higher level C API and coordinated
with a lower level C and/or OS API.  This is the case where Go interfaces AAudio
or Audio Units rather than the respective lower level HAL.
In this case, there would be an extra layer of coordination between Go and the OS, 
but it would not increase latency in terms of buffering of samples, because each encoding
or decoding by Go would happen in the same real time slot as it would in a callback.
The Go code should be scheduled at the same priority as the C API with which it is linked, 
if possible.  If Go code is at a higher priority than the C API, then that might starve
the C level.  Inversely, OS scheduling latency would be less reliable.  
So if the C API  is specially scheduled and the Go code is specially scheduled, then
there might be a decrease in reliability compared to 1 due to the extra coordination efforts,
but the latency in terms of sample buffering would be the same.


## Capture

in C cb, we would have roughly

```
rb->ins[wi] = dat
atomic-incr(size)
incWi(rb)
```

Go side:

```
for {
    for atomic-load(size) == 0 {
        runtime.GoSched() // or sleep w.r.t deadline, but go's sleeps are pretty chaotic.
    }
    if size >= cap {
      // handle overrun
      return
    }
    render from rb->bufs[ri] to the packet.
    // ok, we're the only reader.
    atomic-decr(size)
    incRi(rb)
    // send packet or return to caller.
}
```

### Overruns
It can happen that size is greater than the capacity of rb.  This would
indicate the Go code isn't keeping up.  When this happens, ri should jump
to the index of the oldest packet as if `size == cap - 1` were true and
update size according.

Note that in this design, overruns can happen undetected while Go code is running,
but they will be detected on subsequent runs.  This is not true of playback/capture
where the C callback waits for Go to finish so long as the underlying API
disallows multiple calls to the callback at once.  It may be prudent to 
add this to capture as well, but it would prevent the C API from allowing 
sizes other than 0 or 1.

## Playback
C side:
```
rb->outs[wi] = out;
new = atomic-incr(size);
while(atomic-load(size) == new) {}
incWi(rb);
// dat now contains data for playback, fournished by Go.  Return.
```

Go side:

```
for {
  for atomic-load(size) == 0 {
    runtime.GoSched() // or sleep but go's sleep is quite chaotic
  }
  if size >= cap {
    // handle underrun
    return
  }
  buf := rb->bufs[rb->ri]
  // NB: duplex can read from buf
  // playback fill the buffer
  incRi(rb)
  atomic-decr(size)
}
```

### Underruns
Again, on the Go side, if size exceeds capacity, then there is an underrun.
In this case, `ri` should be set to where it would be if `size == cap - 1` were
true and size should be updated to `cap - 1`

## Duplex
Duplex follows exactly the pattern of playback, except that
- The C code records `in` and `out`.
- The Go code processes both.

### Xruns 
In this case if Go doesn't keep up then the result is an overrun w.r.t. capture
and an underrun w.r.t. playback.



# Discussion
## Latency
To discuss latency, we introduce the notion of _exposed_ and _unexposed_ 
portions of an Rb.  The exposed portion is the part that Go code has
the right to process when there is no xrun.  The unexposed portion is the
rest of the buffer.  For simplicity, we consider these to be disjoint 
sets whose union equals the entire Rb at all points in time.  In practice
this is only true in the critical sections of the C side and Go side, and
thus not true at all points in time.


### Capture
The minimum capture latency between the return of a blocking call and the start
time of the first sample frame returned is 1 buffer worth of time.  For this to
happen, it must be that the computation to synchronise the callback and the Go
code is small enough to represent effectively 0 time with respect to the
time scale granularity. 

The maximum capture latency on non-xrun calls fully determined by the capacity,
or time represented by the exposed part plus time represented by the unexposed
part.  This occurs when the the size atomic variable is at capacity, but not
always.  It may be that the underlying API is operating in parallel with the
Go code.  In this case, the latency is somewhere between the time 
represented by the unexposed part and the total time represented by the Rb.

The analysis with respect to the start time of the blocking call is somewhat
looser since work may not be done by Rb for some time. 

### Playback 
The minimum playback latency between the return of a blocking call and
the output of the first sample is 0.

The maximum playback latency between the return of a blocking call and
the output of the first sample is the time represented by the unexposed
portion of Rb.

The analysis with respect to the start of blocking calls is looser 
as in capture.

### Duplex
Duplex latency should be exactly as capture and playback latency respectively.


## Buffer sizes
Many APIs (eg ALSA, AAudio)  give the user the ability to control something
corresponding to the rb capacity in this document.  (buffer size in ALSA vs
period, AAudio buffer capacity).  

This is complicated and can produce situations where there is a lot of 
computation with no gain in latency.  We simplify the interface so that 
the caller may use larger buffers for more latency and less computation
and smaller buffers for less latency and more computation.  This does
not come at an expense of best case latency, but it does mean the
caller cannot choose the bounds on latency in xrun-free cases.

On the other hand, having a separate parameter for Rb capacity can produce
problems when the underlying API is not capable of providing the requested
buffer size.  TBD: figure out what to do for this case.  In any case, 
we can either effectuate the i/o with more risk of jitter or put the burden
of finding the buffer size on the user, or allow for both.
