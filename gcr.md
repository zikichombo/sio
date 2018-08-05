# Ring Buffer connecting Go and C

# Design characteristics:

- elements are pairs (*C.char, *Packet) containing per-slice buffer sizes,
rather than individual samples.

- Goal: per-slice buffer size minimum latency for playback as well as capture.

## Data elements

- ri, wi  indices for eading and writing

- ri == wi => empty

- single atomic size type.

## Capture

C: 
render to rb[wi].cBuf,
atomic-incr(size)
wi++


Go side:

- for {
    if atomic-load(size) > 0 {
      break 
    }
    runtime.GoSched() // or sleep w.r.t deadline
  }
  if size > cap {
     error()
  }
  render from rb[ri].cBuf to rb[ri].Packet
  // ok, we're the only reader.
  atomic-decr(size)
  ri++
  // send packet

## Playback

"size" is actually the size of available buffers to send to app for filling
initially it is the size of the queue.

This needs initialization detection added on C side

C side:
  copy buf to driver buf
  zero buf
  atomic-incr(size)

Go side:
  atomic-load(size) // record starting size, at first, it is silence.
  for {
      send packet
      receive packet
      for {
        if atomic-load(size) != 0 {
          break
        }
        runtime.GoSched()
      }
      encode packet to buf
      atomic decr(size)
  }

## Duplex

2 queues: one for play, one for capture, same size

C callback: in output callback, calls render then
does C side of capture

Go side, waits for available capture, sends out captured packet,
and then waits for corresponding output packet.  Upon receit, it 


