# Ports

## Status
The status of ports is kept in the main module [README](README.md).

## Overview
ZikiChombo sio code has a porting process that is unconventional due to the
nature sound support on host systems.  Usually, to port a system, on takes some
desired cross-host functionality, such as a filesystem, and implements it with
the OS or OS/hardware combination of interest and that's that (to make a long
story short).

For sound, the problem is that hosts usually have multiple different software
components as entry points to playing and capturing sound, and moreover these
entry points have some mutually exclusive properties, such as lower bounds on
latency, availability of duplex synchronised by hardware audio clock, ability
to be shared accross many users or applications, etc.

Although the main CPU hardware architectures can certainly be a concern, it has
not been so far in our experience.  So we focus on organizing around host,
which is taken to be the value of runtime.GOOS in a Go program.

# Hosts
A host often has a hardware abstraction layer, hardware drivers, and some
levels of interacting with these layers to coordinate and make secure demands
related to playing and capturing sound on a device.

To simplify the end use case, we want to make all this transparent to the
consumer of sio, and let them simply work with sound.{Source,Sink,Duplex}.

For each Host, ZikiChombo defines a default entry point to accomplish this.

All entry points are available both to the consumer of ZikiChombo and either
with ZikiChombo or as 3rd party implementations.  So if you're working with
VoIP and have precise duplex echo cancellation needs, or a music app that
listens and plays in real time with feedback, or writing a PulseAudio
replacement, you're more likely to be interested in using a specialized entry
point than the default.

There is a directory ZikiChombo/sio/{runtime.GOOS} for each host. 

# Entry Points
An entry point defines a proper subset of the functionality listed in the main
[README](README.md).  ZikiChombo defines for each host (runtime.GOOS) a list of
entry points which refer to the software layer with which the go program will
communicate.  These are named after the respective entry points in the main
[README](README.md).

Each entry point defines a base set of functionality related to what the entry
point supports:  device scanning, device change notifications, playback,
capture, duplex.  Note the README excludes some functionality from some
entry points.  The functionality is slightly more rich than implementing
sound.{Source,Sink,Duplex} to allow callers to achieve latency or other 
requirements.  Package sio automatically uses the Entry point to apply
defaults so that the caller may simply call Play, Capture, Duplex on 
sound.{Source,Sink,Duplex}.  Package sio also enforces that only one 
entry point may be in use at a time in one Go program.

Entry points can then be registered by a package implementing them in
their init() function.  Consumers of ZikiChombo may optionaly control which 
package implements a given entry point. The default is chosen by 
package initialisation order.

see [entry](http://godoc.org/zikichombo.org/sio/entry) for details.


# Supporting concepts Devices, Inputs, Outputs, Duplex, Packets
To implement an Entry Point, ZikiChombo provides some 
supporting concepts which can make life easier, but they are 
unnecessary to implement an entry point.  see [sio](http://godoc.org/zikichombo.org/sio)
for details.

## Duplex
Duplex support is intended for synchronized input/output.  Systems which simply
buffer underlying independent I+O and synchronize with the slack that results
from the buffering should consider not implementing Duplex and just letting the
caller use sound.{Source,Sink} synchronously.  Ideally, duplex should be audio
hardware clock synchronized as in Apple's Aggregate devices.

## Build tags
Submitted ports and tests which produce or capture sound should
use the build tag "listen" so that they do not run by default
but are easily invokable with 

```
go test zikichombo.org/sio/... -tags listen
```


# 3rd Party Ports
To have an independently distributed port listed here, please file an issue.
We expect listings to work with the listed versons

| Port | zc version | Port version |
------------------------------------
|  a   |  b         |    c         |
------------------------------------


# References
The following may be useful references for those considering sio ports.
[Plan9 Audio](http://man.cat-v.org/plan_9/3/audio)
[PortAudio](http://portaudio.com)
[FreeBSD](https://www.freebsd.org/doc/handbook/sound-setup.html)
[ALSA](https://www.alsa-project.org/main/index.php/Main_Page)
[RtAudio](http://www.music.mcgill.ca/~gary/rtaudio/)
[Web Audio](https://www.w3.org/TR/webaudio/)
[Audio Units](https://developer.apple.com/documentation/audiounit?language=objc)
[VST](https://www.steinberg.net/en/company/technologies/vst3.html)
[Pulse Audio](https://en.wikipedia.org/wiki/PulseAudio)
[Android Audio](https://source.android.com/devices/audio/terminology)
[AudioFlinger](https://android.googlesource.com/platform/frameworks/av/+/109347d421413303eb1678dd9e2aa9d40acf89d2/services/audioflinger/AudioFlinger.cpp)

