#include <linux/rculist.h>

#include "dev.h"
#include "qer.h"
#include "pdr.h"
#include "far.h"
#include "seid.h"
#include "hash.h"

static void seid_qer_id_to_hex_str(u64 seid_int, u32 qer_id, char *buff)
{
    seid_and_u32id_to_hex_str(seid_int, qer_id, buff);
}

static void qer_context_free(struct rcu_head *head)
{
    struct qer *qer = container_of(head, struct qer, rcu_head);

    if (!qer)
        return;

    kfree(qer);
}

void qer_context_delete(struct qer *qer)
{
    struct gtp5g_dev *gtp = netdev_priv(qer->dev);
    struct hlist_head *head;
    struct pdr_node *pdr_node;
    char seid_qer_id_hexstr[SEID_U32ID_HEX_STR_LEN] = {0};

    if (!qer)
        return;

    if (!hlist_unhashed(&qer->hlist_id))
        hlist_del_rcu(&qer->hlist_id);

    seid_qer_id_to_hex_str(qer->seid, qer->id, seid_qer_id_hexstr);
    head = &gtp->related_qer_hash[str_hashfn(seid_qer_id_hexstr) % gtp->hash_size];
    hlist_for_each_entry_rcu(pdr_node, head, hlist) {
        if (pdr_node->pdr != NULL &&
            find_qer_id_in_pdr(pdr_node->pdr, qer->id)) {
            unix_sock_client_delete(pdr_node->pdr);
        }
    }

    call_rcu(&qer->rcu_head, qer_context_free);
}

struct qer *find_qer_by_id(struct gtp5g_dev *gtp, u64 seid, u32 qer_id)
{
    struct hlist_head *head;
    struct qer *qer;
    char seid_qer_id_hexstr[SEID_U32ID_HEX_STR_LEN] = {0};

    seid_qer_id_to_hex_str(seid, qer_id, seid_qer_id_hexstr);
    head = &gtp->qer_id_hash[str_hashfn(seid_qer_id_hexstr) % gtp->hash_size];
    hlist_for_each_entry_rcu(qer, head, hlist_id) {
        if (qer->seid == seid && qer->id == qer_id)
            return qer;
    }

    return NULL;
}

void qer_update(struct qer *qer, struct gtp5g_dev *gtp)
{
    struct pdr_node *pdr_node;
    struct hlist_head *head;
    char seid_qer_id_hexstr[SEID_U32ID_HEX_STR_LEN] = {0};

    seid_qer_id_to_hex_str(qer->seid, qer->id, seid_qer_id_hexstr);
    head = &gtp->related_qer_hash[str_hashfn(seid_qer_id_hexstr) % gtp->hash_size];
    hlist_for_each_entry_rcu(pdr_node, head, hlist) {
        if (pdr_node->pdr != NULL && find_qer_id_in_pdr(pdr_node->pdr, qer->id)) {
            unix_sock_client_update(pdr_node->pdr, rcu_dereference(pdr_node->pdr->far));
        }
    }
}

void qer_append(u64 seid, u32 qer_id, struct qer *qer, struct gtp5g_dev *gtp)
{
    u32 i;
    char seid_qer_id_hexstr[SEID_U32ID_HEX_STR_LEN] = {0};

    seid_qer_id_to_hex_str(seid, qer_id, seid_qer_id_hexstr);
    i = str_hashfn(seid_qer_id_hexstr) % gtp->hash_size;
    hlist_add_head_rcu(&qer->hlist_id, &gtp->qer_id_hash[i]);
}

int qer_get_pdr_ids(u16 *ids, int n, struct qer *qer, struct gtp5g_dev *gtp)
{
    struct hlist_head *head;
    struct pdr_node *pdr_node;
    int i;
    char seid_qer_id_hexstr[SEID_U32ID_HEX_STR_LEN] = {0};

    seid_qer_id_to_hex_str(qer->seid, qer->id, seid_qer_id_hexstr);
    head = &gtp->related_qer_hash[str_hashfn(seid_qer_id_hexstr) % gtp->hash_size];
    i = 0;
    hlist_for_each_entry_rcu(pdr_node, head, hlist) {
        if (i >= n)
            break;
        if (pdr_node->pdr != NULL && find_qer_id_in_pdr(pdr_node->pdr, qer->id)) {
            ids[i++] = pdr_node->pdr->id;
        }
    }
    return i;
}

void del_related_qer_hash(struct gtp5g_dev *gtp, struct pdr *pdr)
{
    u32 i, j;
    struct pdr_node *pdr_node = NULL ;
    struct pdr_node *to_be_del = NULL ;
    char seid_qer_id_hexstr[SEID_U32ID_HEX_STR_LEN] = {0};

    for (j = 0; j < pdr->qer_num; j++) {
        to_be_del = NULL;
        seid_qer_id_to_hex_str(pdr->seid, pdr->qer_ids[j], seid_qer_id_hexstr);
        i = str_hashfn(seid_qer_id_hexstr) % gtp->hash_size;
        hlist_for_each_entry_rcu(pdr_node, &gtp->related_qer_hash[i], hlist) {
            if (pdr_node->pdr != NULL &&
                pdr_node->pdr->seid == pdr->seid &&
                pdr_node->pdr->id == pdr->id) {
                to_be_del = pdr_node;
                break;
            }
        }
        if (to_be_del){
            hlist_del(&to_be_del->hlist);
            kfree(to_be_del);
        }
    }
}

int qer_set_pdr(struct pdr *pdr, struct gtp5g_dev *gtp)
{
    char seid_qer_id_hexstr[SEID_U32ID_HEX_STR_LEN] = {0};
    u32 i, j;
    struct pdr_node *pdr_node = NULL;

    if (!pdr)
        return -1;

    // clean old pdr_node
    del_related_qer_hash(gtp, pdr);

    for (j = 0; j < pdr->qer_num; j++) {
        seid_qer_id_to_hex_str(pdr->seid, pdr->qer_ids[j], seid_qer_id_hexstr);
        i = str_hashfn(seid_qer_id_hexstr) % gtp->hash_size;

        pdr_node = kzalloc(sizeof(*pdr_node), GFP_ATOMIC);
        if (!pdr_node) {
            return -ENOMEM;
        }
        pdr_node->pdr = pdr;
        hlist_add_head_rcu(&pdr_node->hlist, &gtp->related_qer_hash[i]);
    }
    return 0;
}
