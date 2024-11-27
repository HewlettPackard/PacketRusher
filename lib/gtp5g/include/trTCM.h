#include <linux/ktime.h>

typedef enum {
    Green,
    Yellow,
    Red
} Color;

typedef struct {
    u64 cbs;
    u64 ebs;
    
    u64 tc;
    u64 te;  

    u64 byteRate;
    u64 lastUpdate;
    u64 refillTokenTime;

    spinlock_t lock;
} TrafficPolicer;

TrafficPolicer* newTrafficPolicer(u64);
void refillTokens(TrafficPolicer*);
Color policePacket(TrafficPolicer*, int);
