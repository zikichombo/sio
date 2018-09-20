# Ring Buffer connecting Go and C

this document contains some (still buggy) thoughts on
using atomic to synchronise hardware sound buffers with
go slices/libsio.Packet in callback based sound APIs.

Callback C APIs are very common in sound and very problematic
for Go because they often run on dedicated threads which
cause cgo->go callback overhead:  a lot on first call, and 
it seems atleast, sigaltstack on other calls.  

Since callbacks are on real-time, sigaltstack is inappropriate.

Go is a nice fit for providing a blocking API.  We want to
map a blocking api for example a call to 

```
var d []float64
s, _ := sio.Capture()
s.Receive(d)
```

to be implemented in terms of a callback C API, roughly
```
int cb(void *buf, ...)
```

where cb is called, presumably on a foreign thread in a way
that is clocked to synchronize with the number of frames
in `d` and either 
1. requests the caller to copy buf for subsequent treatment of captured data; or
1. requests the caller to fill buf for subsequent playback; or
1. requests the caller to do both, in order.


In sio, we have assumed the user may only specify the desired
buffer size of `d` in frames.  In practice, behind this, there
is some ringbuffer which `buf` points to with `nF` frames
of data available, and the other part of the ringbuffer 
being filled or coordinated with the hardware or lower level API.

Depending on the implemented mechanism, the size of the underlying
ring buffer may vary and effect latency.  For example, for double
buffering, the size of the underlying ringbuffer would would be 2 times
the size of `d`.  In a less synchronized context, the size might be 3 times
the size of `d` so that one buffer may be in transit while both sides treats
the other 2 buffers in parallel.


# Design characteristics:

To deal with this, we present a ringbuffer in which the elements
are pairs 

```
(buf *C.char, pkt *libsio.Packet)
```

The ringbuffer has a write index `wi` and a read index `ri` which must be coordinated between C and Go.
These indices indicate where the next write or read will occur.  There is also a 'size' variable
indicating the number of samples between the `ri` and `wi` (circularly ordered).
The 'size' variable is atomically synchronised between the two sides.  The read side 
keeps `ri` and the write side keeps `wi`.

We assume that Go code will encode `pkt` into `buf` for playback and `buf` into `pkt`
for capture.  The duplex case is similar.  A processing chain may do more, but is
outside the scope of sio for the time being.

There are many combinations worth noting.  In the following, we consider "reliability" 
w.r.t. OS scheduling overhead and Go GC.  The latter looks good.  Dealing with the
former is the whole reason for designing and implementing this.  In practice, there
is a sweet spot between reliability and latency:  you can reduce the latency to the
level where it is sufficiently reliable for your use case.  The cases are enumerated
below.

1. The ringbuffer is hardware/low level OS priveleged and Go is on a specially
scheduled thread.  In this case, the implementation is the same w.r.t. 
OS scheduling as say AAudio on top of Android HAL or AUHAL on top of CoreAudio's 
Audio Device HAL, and we would expect equivalent latency reliability modulo Go garbage collection.
1. Same as above, but Go code is not specially OS scheduled for audio.  In this case,
we would expect an increase in glitching under non-dedicated hardware or system load.
1. The ringbuffer or similar is implemented by a higher level C API and coordinated
with a lower level C and/or OS API.
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

in C, we would have roughly

```
render to rb[wi].cBuf,
atomic-incr(size)
wi++
```

Go side:

```
for {
    for atomic-load(size) == 0 {
        runtime.GoSched() // or sleep w.r.t deadline, but go's sleeps are pretty chaotic.
    }
    if size > cap {
         error()
    }
    render from rb[ri].cBuf to rb[ri].Packet
    // ok, we're the only reader.
    atomic-decr(size)
    ri++
    // send packet
}
```


### Memory copies
Note this requires 2 copies/moves of memory rather than 1 if the data is not in
floating point format.  However, extra functionality is also provided: we get
floats for and sample format/codec.  In systems which do processing, floats are
the most likely representation and would have to be done eventually anyway.  By
including the conversion here, we guarantee that no latency overhead would be
needed to do it later.  As sio is for sound processing and i/o, this seems
reasonable.

TBD: see if we can just reference the underlying buffer directly.


## Playback

"size" is actually the size of available buffers to send to app for filling
initially it is the size of the queue. 

This needs initialization detection added on C side

C side:
  copy buf to driver buf
  zero buf
  atomic-incr(size)

Go side:

```
atomic-load(size) // record starting size, at first, it is silence.
for {
  send packet
  receive packet
  for {
    if atomic-load(size) != 0 {
      break
    }
    runtime.GoSched() // or sleep but go's sleep is quite chaotic
  }
  encode packet to buf
  atomic decr(size)
}
```

## Duplex
in C, we would have roughly

```
render to rb[wi].cBuf,
atomic-incr(size)
wi++
```

Go side:

```
for {
    for atomic-load(size) == 0 {
        runtime.GoSched() // or sleep w.r.t deadline, but go's sleeps are pretty chaotic.
    }
    if size > cap {
         error()
    }
    render from rb[ri].cBuf to rb[ri].Packet
    // ok, we're the only reader.
    atomic-decr(size)
    ri++
    // send packet
}
```




