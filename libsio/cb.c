// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// The goal is to synchronize between Go and C without cgo->go callback
// on a thread created outside of go.
#include <stdlib.h>
#include <stdatomic.h>
#include <string.h>
#include <stdint.h>
#include <time.h>
#include <stdio.h>

#include "cb.h"


Cb * newCb(int bufSz) {
	Cb * cb = (Cb*) malloc(sizeof(Cb));
	if (cb == 0) {
		return cb;  // Go code will call and must check.
	}
	cb->bufSz = bufSz;
	cb->in = 0;
	cb->inF = 0;

	cb->out = 0;
	cb->outF = 0;
	cb->time.tv_sec = 0;
	cb->time.tv_nsec = 1000;
	cb->inGo = 0;
	cb->inCb = inCb;
	cb->outCb = outCb;
	cb->duplexCb = duplexCb;
	return cb;
}

void freeCb(Cb *cb) {
	free(cb);
}

void * getIn(Cb *cb) {
	return cb->in;
}

int getInF(Cb *cb) {
	return cb->inF;
}

void * getOut(Cb *cb) {
	return cb->out;
}

int getOutF(Cb *cb) {
	return cb->outF;
}

// forward decl.
static void toGoAndBack(Cb *cb);

/*
 From Ian Lance Taylor: in C11 stdatomic terms Go atomic.CompareAndSwap is like
atomic_compare_exchange_weak_explicit(p, &old, new,
memory_order_acq_rel, memory_order_relaxed).  

This is however undocumented in the language, so from wsc: don't use it to
control an airplane.  But for audio, we'll give it a shot.
*/

/*
 * inCb reads dat into the buffer, advancing the buffer write index.  It
 * spins until there is space.
 *
 * inCb is written to be called in an audio i/o callback API, as described
 * in cb.md, for capture.
 */
void inCb(Cb *cb, void *in, int nF) {
	cb->in = in;
	cb->inF = nF;
	toGoAndBack(cb);
}

void outCb(Cb *cb, void *out, int *nF) {
	cb->out = out;
	cb->outF = *nF;
	toGoAndBack(cb);
	*nF = cb->outF;
}

void duplexCb(Cb *cb, void *out, int *onF, void *in, int inF) {
	cb->in = in;
	cb->inF = inF;
	cb->out = out;
	cb->outF = *onF;
	toGoAndBack(cb);
	*onF = cb->outF;
}


static void toGoAndBack(Cb *cb) {
	_Atomic uint32_t * gp = &(cb->inGo);
	uint32_t b; 
	for (;;) {
		b = atomic_load_explicit(gp, memory_order_acquire);
		if (b > 0) { 
			// somehow more than one callback called.
			// can't happen unless the underlying API executes callbacks
			// on more than one thread, as anyhow the current thread
			// is in this function (no setjmp/longjmp).
			fprintf(stderr, "toGoAndBack: %d > 0, does the API guarantee one callback at a time?\n", b);
			continue;
		}
		if (atomic_compare_exchange_weak_explicit(gp, &b, b+1, memory_order_acq_rel, memory_order_relaxed)) {
			break;
		}
	}
	b++;
	uint32_t cmp;
	int i;
	for (i=1; i<=1000000;i++) {
		cmp = atomic_load_explicit(gp, memory_order_acquire);
		if (cmp != b) {
			return;
		}
		if (i >= 1000 && i%50 == 0) {
			// sleep is 1us, but involves syscall so system latency is involved.
			// avoid as much as possible without entirely eating the CPU.
			// TBD(wsc) make this buffer size real time dependent rather than by cycle
			// counts.  If the buffer time is large, then we can sleep as in 
			// cb.go, otherwise either we're in a slack time or contention is causing the 
			// atomic to fail and it might help to back off.
			nanosleep(&cb->time, NULL);
		}
	}
	// paranoid code to reset state to as if Go was running properly
	// equivalent of libsio.ErrCApiLost
	fprintf(stderr, "atomic failed after 1000000 tries (resetting), did Go die?\n");
	while (b>0) {
		b = atomic_load_explicit(gp, memory_order_acquire);
		if (atomic_compare_exchange_weak_explicit(gp, &b, 0, memory_order_acq_rel, memory_order_relaxed)) {
			break;
		}
	}
}

// TBD
void closeCb(Cb *cb) {
}




