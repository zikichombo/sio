# [ZikiChombo](http://zikichombo.org) sio project

[![Build Status](https://travis-ci.com/zikichombo/sio.svg?branch=master)](https://travis-ci.com/zikichombo/sio)

# Usage
If you are using sio for sound capture and playback, only the [sio](http://godoc.org/zikichombo.org/sio)
package is needed.  For device scanning and APIs, the [host](http://godoc.org/zikichombo.org/sio/host) 
package provides the necessary support.


# Ports
For porting, see the [porting guide](Porting.md) and [contributing](Contributing.md).

## Status

Below is the status of sio ports.  Items marked with an "X" in plain text (checked off
in html rendered markdown) are incorporated into sio, potentially with alpha status.
Items marked with a "?" indicates we do not yet have sufficient knowledge to judge 
whether or not the item is a TODO.  Related discussion on the issue tracker is welcome.
Items marked with "-" are those for which we think the functionality is not relevant or 
not sufficiently supported by the external software interface to add to sio.

In the event there are opinions about the content of the list itself, such as whether 
to support JACK, whether to interface with Android HAL, the issue tracker is our best
means of coordinating the discussion.


* Linux
    1. ALSA (cgo)
        1. [X] Playback
        1. [X] Capture
        1. [ ] Duplex
        1. [ ] Device Scanning
        1. [ ] Device Notification
    1. TinyALSA (cgo)
        1. [ ] Playback
        1. [ ] Capture
        1. [ ] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
    1. ALSA (no cgo)
        1. [?] Playback
        1. [?] Capture
        1. [?] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
    1. Pulse Audio
        1. [ ] Playback
        1. [ ] Capture
        1. [?] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
* Darwin/iOS
    1. Audio Queue Services
        1. [X] Playback
        1. [X] Capture
        1. [-] Duplex
        1. [-] Device Scanning
        1. [ ] Test for iOS
    1. AUHAL
        1. [ ] Playback
        1. [ ] Capture
        1. [ ] Duplex
        1. [X] Device Scanning
        1. [ ] Test for iOS via RemoteIO replacing AUHAL.
    1. VPIO [?]
* Android
    1. TinyALSA
        1. [ ] Playback
        1. [ ] Capture
        1. [ ] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
    1. Android Audio HAL [?]
    1. AudioFlinger
        1. [ ] Playback
        1. [ ] Capture
        1. [?] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
    1. AAudio
        1. [ ] Playback
        1. [ ] Capture
        1. [?] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
    1. OpenSL ES
        1. [ ] Playback
        1. [ ] Capture
        1. [?] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
* Windows
    1. Direct Sound
        1. [?] Playback
        1. [?] Capture
        1. [?] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
    1. WASAPI
        1. [?] Playback
        1. [?] Capture
        1. [?] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
* js
    1. Web Audio
        1. [ ] Playback
        1. [ ] Capture
        1. [-] Duplex
        1. [ ] Device Scanning
        1. [?] Device Notification

* plan9 [?]
* netbsd [?]
* freebsd [?]
* openbsd [?]
* dragonfly [?]



