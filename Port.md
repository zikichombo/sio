# Ports

## Status
The status of ports is kept in the main module [README](README.md).

## Overview
ZikiChombo sio code has a porting process that is unconventional due to the
nature sound support on host systems.  Usually, to port a system, on takes some
desired cross-host functionality, such as a filesystem or user interface, and
implements it the OS or OS/hardware combination of interest and that's that 
(to make a long story short).

For sound, the problem is that hosts usually have multiple different 
software components as entry points to playing and capturing sound,
and moreover these entry points have some mutually exclusive 
properties, such as lower bounds on latency, availability of 
duplex synchronised by hardware audio clock, ability to be shared
accross many users or applications, etc.

Although the main cpu hardware architectures can certainly be a concern,
it has not so far in our experience.  So we focus on organizing around
host, which is taken to be the value of runtime.GOOS in a Go program.

# Hosts
A host often has a hardware abstraction layer, hardware drivers, and some
levels of interacting with these layers to coordinate and make secure
demands related to playing and capturing sound on a device.

To simplify the end use case, we want to make all this transparent to 
the consumer of sio, and let them simply work with sound.{Source,Sink,Duplex}.

For each Host, ZikiChombo defines a default entry point to accomplish this.

All entry points are available both to the consumer of ZikiChombo and 
either with ZikiChombo or as 3rd party implementations.  So if you're
working with VoIP and have precise duplex echo cancellation needs, 
or a music app that listens and plays in real time with feedback, 
you're more likely to be interested in using a specialized entry point 
than the default.

There is a directory ZikiChombo/sio/{runtime.GOOS} for each host. 
Each directory contains code for all entry points for that host,
and registers the entry points in its package init().

# Entry Point
An entry point defines the following interfaces.


# Supporting concepts Devices, Inputs, Outputs, Duplex, Packets
To implement an Entry Point, ZikiChombo provides some 
supporting concepts which can make life easier, but they are 
unnecessary.





