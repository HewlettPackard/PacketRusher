#include <linux/time.h>
#include <linux/slab.h>
#include <linux/random.h>

#include "trTCM.h"
#include "log.h"

#define MTU 1500
#define MILLISECONDS_PER_SECOND 1000
#define NANOSECONDS_PER_SECOND 1000000000
#define AVG_WINDOW 1000 // ms
#define BURST_DURATION 200 // ms

TrafficPolicer* newTrafficPolicer(u64 kbitRate) {
    TrafficPolicer* p = (TrafficPolicer*)kmalloc(sizeof(TrafficPolicer), GFP_KERNEL);
    if (p == NULL) {
        GTP5G_ERR(NULL, "traffic policer memory allocation error\n");
        return NULL;
    }
    
    spin_lock_init(&p->lock);
    
    p->byteRate = kbitRate * 125 ; // Kbit/s to byte/s (*1000/8)

    // CBS size = CIR * AVG_WINDOW
    p->cbs = p->byteRate * (AVG_WINDOW / MILLISECONDS_PER_SECOND); // bytes
    if (p->cbs < MTU) {
        p->cbs = MTU;
    }

    // EBS size = CIR * BURST_DURATION
    p->ebs = p->byteRate * BURST_DURATION / MILLISECONDS_PER_SECOND; // bytes

    // fill buckets at the begining
    p->tc = p->cbs; 
    p->te = p->ebs; 

    p->lastUpdate = ktime_get_ns();

    p->refillTokenTime = 0;

    return p;
} 

Color policePacket(TrafficPolicer* p, int pktLen) {
    u64 refillTokens = 0;
    u64 tc, te = 0;
    u64 elapsed = 0;
    u64 now = 0;

    spin_lock(&p->lock); 

    now = ktime_get_ns();
    elapsed = now - p->lastUpdate;
    p->lastUpdate = now;

    // the total time elapsed since the last token refill
    p->refillTokenTime = p->refillTokenTime + elapsed;

    refillTokens = p->byteRate * p->refillTokenTime / NANOSECONDS_PER_SECOND;

    if (refillTokens > 0) {
        u64 remainTime = p->refillTokenTime -
            (refillTokens * NANOSECONDS_PER_SECOND / p->byteRate);
        p->refillTokenTime = remainTime;
    }
 
    tc = p->tc + refillTokens;
    te = p->te;
    if (tc > p->cbs) {
        te += (tc - p->cbs);
        if (te > p->ebs){
            te = p->ebs;
        }
        tc = p->cbs; 
    }
   
    if (tc >= pktLen) {
        p->tc = tc - pktLen;
        p->te = te;
        spin_unlock(&p->lock);
        return Green;
    }

    if (te >= pktLen) {
        p->tc = tc;
        p->te = te - pktLen;
        spin_unlock(&p->lock);
        return Yellow;
    }

    p->tc = tc;
    p->te = te;

    spin_unlock(&p->lock); 
    return Red;
}
