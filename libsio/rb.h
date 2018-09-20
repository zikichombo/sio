#ifndef ZC_LIBSIO_RC_H
#define ZC_LIBSIO_RC_H

#include <stdatomic.h>
#include <stdint.h>

typedef struct Rb {
	int ri;
	int wi;
	_Atomic uint32_t size;
	int nb;
	int bufSz;
	void ** ins;  
	void ** outs;
} Rb;


Rb * newRb(int nb, int bufSz);

void freeRb(Rb *rb);
void * getIn(Rb *rb);
void * getOut(Rb *rb);
void incRi(Rb *rb);
void incWi(Rb *rb);
void inCb(Rb *rb, void *in);
void outCb(Rb *rb, void *out);
void duplexCb(Rb *rb, void *out, void *in);


#endif
