// this file is as-of-yet unused, uncompiled
// prototype for synchronising between Go and C
// with atomics for audio as described in rb.md.
// 
// The goal is to synchronize between Go and C without cgo->go callback
// on a thread created outside of go.
#include <stdlib.h>
#include <stdatomic.h>
#include <string.h>

typedef struct Rb {
	int ri;
	int wi;
	_Atomic int size;
	int nb;
	int bufSz;
	void ** ins;  
	void ** outs;
} Rb;


Rb * newRb(int nb, int bufSz) {
	Rb * rb = (Rb*) malloc(sizeof(Rb));
	if (rb == 0) {
		return rb;  // Go code will call and must check.
	}
	rb->ri = 0;
	rb->wi = 0;
	rb->size = 0;
	rb->nb = nb;
	rb->bufSz = bufSz;
	rb->ins = (void **) calloc(sizeof(void*),  nb);
	if (rb->ins == 0) {
		free(rb);
		return 0;
	}
	rb->outs = (void **) calloc(sizeof(void*), nb);
	if (rb->outs == 0) {
		free(rb->ins);
		free(rb);
		return 0;
	}
	return rb;
}

void freeRb(Rb *rb) {
	free(rb->ins);
	free(rb->outs);
	free(rb);
}

void incWi(Rb *rb) {
	rb->wi++;
	if (rb->wi == rb->nb) {
		rb->wi = 0;
	}
}

void incRi(Rb *rb) {
	rb->ri++;
	if (rb->ri == rb->nb) {
		rb->ri = 0;
	}
}

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
 * in rb.md, for capture.
 */
void inCb(Rb *rb, void *in) {
	rb->ins[rb->wi] = in;
	_Atomic int * szp = &rb->size;
	int sz; 
	for (;;) {
		sz = rb->size;
		if (atomic_compare_exchange_weak_explicit(szp, &sz, sz+1, memory_order_acq_rel, memory_order_relaxed)) {
			break;
		}
	}
	incWi(rb);
}

void outCb(Rb *rb, void *out) {
	rb->outs[rb->wi] = out;
	_Atomic int * szp = &rb->size;
	int sz; 
	for (;;) {
		sz = rb->size;
		if (atomic_compare_exchange_weak_explicit(szp, &sz, sz+1, memory_order_acq_rel, memory_order_relaxed)) {
			break;
		}
	}
	sz++;
	int cmp;
	for (;;) {
		cmp = atomic_load_explicit(szp, memory_order_acquire);
		if (cmp != sz) {
			break;
		}
	}
	incWi(rb);
}

void duplexCb(Rb *rb, void *out, void *in) {
	rb->ins[rb->wi] = in;
	rb->outs[rb->wi] = out;
	_Atomic int * szp = &rb->size;
	int sz; 
	for (;;) {
		sz = rb->size;
		if (atomic_compare_exchange_weak_explicit(szp, &sz, sz+1, memory_order_acq_rel, memory_order_relaxed)) {
			break;
		}
	}
	sz++;
	int cmp;
	for (;;) {
		cmp = atomic_load_explicit(szp, memory_order_acquire);
		if (cmp != sz) {
			break;
		}
	}
	incWi(rb);
}




