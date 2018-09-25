#ifndef ZC_LIBSIO_RC_H
#define ZC_LIBSIO_RC_H

#include <stdatomic.h>
#include <stdint.h>

typedef struct Cb {
	int bufSz;
	_Atomic uint32_t inGo;
	void * in;  
	int inF;
	void * out;
	int outF;
} Cb;


Cb * newCb(int bufSz);

void freeCb(Cb *cb);
void * getIn(Cb *cb);
int getInF(Cb *cb);
void * getOut(Cb *cb);
int getOutF(Cb *cb);
void inCb(Cb *cb, void *in, int nf);
void outCb(Cb *cb, void *out, int *nf);
void duplexCb(Cb *cb, void *out, int *nf, void *in, int isz);
void closeCb(Cb *cb);


#endif
