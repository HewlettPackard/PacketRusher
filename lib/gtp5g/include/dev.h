#ifndef __GTP5G_DEV_H__
#define __GTP5G_DEV_H__

#include <linux/netdevice.h>
#include <linux/rculist.h>
#include <linux/socket.h>

struct usage_statistic {
    atomic_t ul_byte;
    atomic_t dl_byte;
    atomic_t ul_pkt;
    atomic_t dl_pkt;
};

struct gtp5g_dev {
    struct list_head list;
    struct sock *sk1u; // UDP socket from user space
    struct net_device *dev;
    unsigned int role;
    unsigned int hash_size;
    struct hlist_head *pdr_id_hash;
    struct hlist_head *far_id_hash;
    struct hlist_head *qer_id_hash;
    struct hlist_head *bar_id_hash;
    struct hlist_head *urr_id_hash;

    struct hlist_head *i_teid_hash; // Used for GTP-U packet detect
    struct hlist_head *addr_hash;   // Used for IPv4 packet detect
    
    /* IEs list related to PDR */
    struct hlist_head *related_far_hash; // PDR list waiting the FAR to handle
    struct hlist_head *related_qer_hash; // PDR list waiting the QER to handle
    struct hlist_head *related_bar_hash;
    struct hlist_head *related_urr_hash;
    
    /* Used by proc interface */
    struct list_head proc_list;

    /* Usage Statistics */
    struct usage_statistic rx, tx;
};

extern const struct net_device_ops gtp5g_netdev_ops;

struct gtp5g_dev *gtp5g_find_dev(struct net *, int, int);
int dev_hashtable_new(struct gtp5g_dev *, int);
void gtp5g_hashtable_free(struct gtp5g_dev *);
void update_usage_statistic(struct gtp5g_dev *, u64, u64, int, uint);

#endif // __GTP5G_DEV_H__
