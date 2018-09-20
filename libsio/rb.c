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
	volatile int size;
	int cap;
	int bufSz;
	void ** bufs;  // C memory
} Rb;


Rb * newRb(int rbSz, int bufSz) {
	Rb * rb = (Rb*) malloc(sizeof(Rb));
	if (rb == 0) {
		return rb;  // Go code will call and must check.
	}
	rb->ri = 0;
	rb->wi = 0;
	rb->size = 0;
	rb->cap = rbSize;
	rb->bufSz = bufSz;
	rb->bufs = (char **) malloc(sizeof(char*) * rbSz);
	if (rb->bufs == 0) {
		free(rb);
		return 0;
	}
	// TBD: allocate once and compute pointer offsets to avoid fragmentation
	// of memory backing bufs.
	for (int i = 0; i < rbSz; i++) {
		rb->bufs[i] = (char *) calloc(sizeof(char), bufSz);
		if (rb->bufs[i] == 0) {
			for (int j=0; j < i; j++) {
				free(rb->bufs[j]);
			}
			free(rb->bufs);
			free(rb);
			return 0;
		}
	}
	return rb;
}

void freeRb(rb *Rb) {
	for (int i = 0; i < rb->cap; i++) {
		free(rb->bufs[i]);
	}
	free(rb->bufs);
	free(rb);
}

void primePlay(rb *Rb) {
}

void incWi(rb *Rb) {
	rb->wi++;
	if (rb->wi == rb->cap) {
		rb->wi = 0;
	}
}

void incRi(rb *Rb) {
	rb->ri++;
	if (rb->ri == rb->cap) {
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
 * inRb reads dat into the buffer, advancing the buffer write index.  It
 * spins until there is space.
 *
 * inRb is written to be called in an audio i/o callback API, as described
 * in rb.md, for capture.
 */
void inRb(rb *Rb, void *dat) {
	/* TBD: don't allocate memory, just use the passed dat and place it in tb->bufs[rb->wi] */
	void * buf = rb->bufs[rb->wi];
	memcpy(buf, dat, rb->bufSz);
	volatile int * szp = &rb->size;
	int sz; 
	for (;;) {
		sz = rb->size;
		if (atomic_compare_exchange_weak_explicit(szp, &sz, sz+1, memory_order_acq_rel, memory_order_relaxed)) {
			break;
		}
	}
	incWi(rb);
}

void outRb(rb *Rb, void *dat) {
}



