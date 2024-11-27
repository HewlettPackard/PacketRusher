#ifndef __BAR_H__
#define __BAR_H__

#include <linux/kernel.h>
#include <linux/rculist.h>
#include <linux/net.h>

#include "dev.h"

#define SEID_U32ID_HEX_STR_LEN 24

struct bar {
    struct hlist_node hlist_id;
    u64 seid;
    u8 id;
    uint8_t delay;
    uint16_t count;
    struct net_device *dev;
    struct rcu_head rcu_head;
};

void bar_context_delete(struct bar *);
struct bar *find_bar_by_id(struct gtp5g_dev *, u64, u32);
void bar_update(struct bar *, struct gtp5g_dev *);
void bar_append(u64, u32, struct bar *, struct gtp5g_dev *);
int bar_get_far_ids(u32 *, int, struct bar *, struct gtp5g_dev *);
void bar_set_far(u64, u32, struct hlist_node *, struct gtp5g_dev *);

#endif // __BAR_H__
