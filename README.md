# [ZikiChombo](http://zikichombo.org) sio project

[![Build Status](https://travis-ci.com/zikichombo/sio.svg?branch=master)](https://travis-ci.com/zikichombo/sio)

# Ports
For porting, see the [porting guide](Port.md) and [contributing](Contributing.md)

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
* Android
    1. TinyALSA
        1. [ ] Playback
        1. [ ] Capture
        1. [ ] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
    1. AudioFlinger
        1. [ ] Playback
        1. [ ] Capture
        1. [?] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
* Windows
    1. ASIO
        1. [?] Playback
        1. [?] Capture
        1. [?] Duplex
        1. [?] Device Scanning
        1. [?] Device Notification
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



