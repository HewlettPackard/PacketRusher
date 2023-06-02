#ifndef __GENL_BAR_H__
#define __GENL_BAR_H__

#include "genl.h"

/* BAR attributes */
enum gtp5g_bar_attrs {
    GTP5G_BAR_ID = 3,
    GTP5G_DOWNLINK_DATA_NOTIFICATION_DELAY,
    GTP5G_BUFFERING_PACKETS_COUNT,
    GTP5G_BAR_SEID,

    __GTP5G_BAR_ATTR_MAX,
};
#define GTP5G_BAR_ATTR_MAX (__GTP5G_BAR_ATTR_MAX - 1)

struct buffer_action {
    u64 seid;
    u16 notification_delay;
    u32 buffer_packet_count;
} __attribute__((packed));

/* for kernel */
extern int gtp5g_genl_add_bar(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_del_bar(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_get_bar(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_dump_bar(struct sk_buff *, struct netlink_callback *);

#endif // __GENL_BAR_H__
