// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

#ifndef ZC_LIBSIO_RC_H
#define ZC_LIBSIO_RC_H

#include <stdatomic.h>
#include <stdint.h>
#include <time.h>

// Cb holds data for exchanging information between a C or OS callback on a 
// dedicated thread foreign to the calling application and Go.
// 
typedef struct Cb {
	int bufSz;  // in frames
	_Atomic uint32_t inGo;  // whether we've passed of the buffer to go and stlll await it to finish
	void * in;  // input buffer
	int inF; // input number of sample frames
	void * out; // output buffer
	int outF; // output number of sample frames
	struct timespec time;  // for throttling a bit when there's a long wait.


	// function pointers below are used to give access to callbacks to 
	// other Go packaages.   They can can access a Cb * as a go unsafe.Pointer by importing
	// libsio, and then C code in that package can call these function pointers.
	// It doesn't seem possible to share C function definitions between packages by cgo alone,
	// so we came up with this mechanism.  These fields are populated by newCb.
	void (*inCb)(struct Cb *cb, void *in, int nf);
	void (*outCb)(struct Cb *cb, void *out, int *nf);
	void (*duplexCb)(struct Cb *cb, void *out, int *onf, void *in, int inf);
} Cb;

#ifdef LIBSIO

// only libsio can access these functions via cgo package "C".
Cb * newCb(int bufSz);

void freeCb(Cb *cb);
void closeCb(Cb *cb);

void * getIn(Cb *cb);
int getInF(Cb *cb);
void * getOut(Cb *cb);
int getOutF(Cb *cb);


void inCb(Cb *cb, void *in, int nf);
void outCb(Cb *cb, void *out, int *nf);
void duplexCb(Cb *cb, void *out, int *nf, void *in, int isz);
#endif


#endif
