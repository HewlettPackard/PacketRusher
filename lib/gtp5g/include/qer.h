#ifndef __QER_H__
#define __QER_H__

#include <linux/kernel.h>
#include <linux/rculist.h>
#include <linux/net.h>

#include "dev.h"
#include "pdr.h"

// TS 29.244 8.2.7 Gate Status
// 0 OPEN, 1 CLOSED
// 2, 3 For future use. Shall not be sent.
// If received, shall be interpreted as the value "1"
#define QER_UL_GATE_CLOSE (0x1 << 2)
#define QER_DL_GATE_CLOSE (0x1 << 0)

struct qer {
    struct hlist_node hlist_id;
    u64 seid;
    u32 id;
    uint8_t ul_dl_gate;
    struct {
        uint32_t ul_high;
        uint8_t ul_low;
        uint32_t dl_high;
        uint8_t dl_low;
    } mbr;
    u64 ul_mbr;
    u64 dl_mbr;

    struct {
        uint32_t ul_high;
        uint8_t ul_low;
        uint32_t dl_high;
        uint8_t dl_low;
    } gbr;
    uint32_t qer_corr_id;
    uint8_t rqi;
    uint8_t qfi;
    uint8_t ppi;
    uint8_t rcsr;
    struct net_device *dev;
    struct rcu_head rcu_head;

    TrafficPolicer  *ul_policer, *dl_policer;  
};

void qer_context_delete(struct qer *);
struct qer *find_qer_by_id(struct gtp5g_dev *, u64, u32);
void qer_update(struct qer *, struct gtp5g_dev *);
void qer_append(u64, u32, struct qer *, struct gtp5g_dev *);
int qer_get_pdr_ids(u16 *, int, struct qer *, struct gtp5g_dev *);
int qer_set_pdr(struct pdr *, struct gtp5g_dev *);
void del_related_qer_hash(struct gtp5g_dev *, struct pdr *);
void set_pdr_qer_with_rate_null(struct qer *, struct gtp5g_dev *);

#endif // __QER_H__
