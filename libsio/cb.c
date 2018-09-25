// The goal is to synchronize between Go and C without cgo->go callback
// on a thread created outside of go.
#include <stdlib.h>
#include <stdatomic.h>
#include <string.h>
#include <stdint.h>

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
static int onF(Cb *cb);

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
	_Atomic uint32_t * gp = &cb->inGo;
	uint32_t b; 
	for (;;) {
		b = atomic_load_explicit(gp, memory_order_acquire);
		if (b > 0) { 
			// somehow more than one callback called.
			// can't happen unless the underlying API executes callbacks
			// on more than one thread, as anyhow the current thread
			// is in this function (no setjmp/longjmp).
			continue;
		}
		if (atomic_compare_exchange_weak_explicit(gp, &b, b+1, memory_order_acq_rel, memory_order_relaxed)) {
			break;
		}
	}
	b++;
	uint32_t cmp;
	for (;;) {
		cmp = atomic_load_explicit(gp, memory_order_acquire);
		if (cmp != b) {
			break;
		}
	}
}

// TBD
void closeCb(Cb *cb) {
}




