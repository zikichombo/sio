// +build cgo

package libsio

// this file is for testing only, but _test.go files don't support cgo,
// so you have to compile this every time :(

// #include "cb.h"
// #include <stdlib.h>
// #include <pthread.h>
//
// typedef struct runcb {
//     Cb * cb;
//     int n;
//     void * buf;
//     int bf;
//     int bpf;
// } runcb;
//
// runcb * newruncb(Cb *cb, int n, int bf, int bpf) {
//   runcb * rcb = (runcb *) malloc(sizeof(runcb));
//   if (rcb == 0) {
//       return rcb;
//   }
//
//   rcb->cb = cb;
//   rcb->n = n;
//   rcb->bf = bf;
//   rcb->bpf = bpf;
//   rcb->buf = malloc(bf *bpf);
//   if (rcb->buf == 0) {
//      free(rcb);
//      return 0;
//   }
//   return rcb;
// }
//
// void freercb(runcb *rcb) {
//     free(rcb->buf);
//     free(rcb);
// }
//
//
// void * kickoff(void *p) {
//     runcb *rcb = (runcb *) p;
//     struct timespec sleepTime;
//     sleepTime.tv_sec = 0;
//     sleepTime.tv_nsec = 500000L;
//     for (int i = 0; i < rcb->n; i++) {
//         inCb(rcb->cb, rcb->buf, rcb->bf);
//         nanosleep(&sleepTime, NULL);
//     }
//     return NULL;
// }
//
// int runcbs(Cb *cb, int n, int b, int bps) {
//
//    runcb * rcb = newruncb(cb, n, b, bps);
//    if (rcb == 0) {
//        return -1;
//    }
//
//    pthread_t hwa_emu;
//    int ret = 0;
//
//    if (pthread_create(&hwa_emu, NULL, kickoff, rcb)) {
//        freercb(rcb);
//        ret = 1;
//        goto cleanup;
//    }
//
//    if (pthread_join(hwa_emu, NULL)) {
//        ret = 2;
//        goto cleanup;
//    }
//
//    cleanup:
//    freercb(rcb);
//    return ret;
// }
import "C"

func runcbs(cb *Cb, n, bf, spf int) {
	C.runcbs(cb.c, C.int(n), C.int(bf), C.int(spf))
}
