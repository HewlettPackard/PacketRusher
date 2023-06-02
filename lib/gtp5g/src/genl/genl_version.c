#include "genl_version.h"

static int gtp5g_genl_fill_ver(struct sk_buff *skb, u32 snd_portid, u32 snd_seq,
        u32 type)
{
    void *genlh;

    genlh = genlmsg_put(skb, snd_portid, snd_seq, &gtp5g_genl_family, 0, type);
    if (!genlh)
        goto genlmsg_fail;

    if (nla_put_string(skb, GTP5G_VERSION, DRV_VERSION))
        goto genlmsg_fail;

    genlmsg_end(skb, genlh);
    return 0;

genlmsg_fail:
    genlmsg_cancel(skb, genlh);
    return -EMSGSIZE;
}

int gtp5g_genl_get_version(struct sk_buff *skb, struct genl_info *info)
{
    struct sk_buff *skb_ack;
    int err;

    skb_ack = genlmsg_new(NLMSG_GOODSIZE, GFP_ATOMIC);
    if (!skb_ack) {
        return -ENOMEM;
    }

    err = gtp5g_genl_fill_ver(skb_ack,
            NETLINK_CB(skb).portid,
            info->snd_seq,
            info->nlhdr->nlmsg_type);
    if (err) {
        kfree_skb(skb_ack);
        return err;
    }

    return genlmsg_unicast(genl_info_net(info), skb_ack, info->snd_portid);
}
