#include <linux/module.h>
#include <linux/version.h>
#include <net/genetlink.h>

#include "genl.h"
#include "genl_pdr.h"
#include "genl_far.h"
#include "genl_qer.h"
#include "genl_bar.h"
#include "genl_urr.h"
#include "genl_report.h"
#include "genl_version.h"

static const struct nla_policy gtp5g_genl_pdr_policy[GTP5G_PDR_ATTR_MAX + 1] = {
    [GTP5G_PDR_ID]                              = { .type = NLA_U32, },
    [GTP5G_PDR_PRECEDENCE]                      = { .type = NLA_U32, },
    [GTP5G_PDR_PDI]                             = { .type = NLA_NESTED, },
    [GTP5G_OUTER_HEADER_REMOVAL]                = { .type = NLA_U8, },
    [GTP5G_PDR_FAR_ID]                          = { .type = NLA_U32, },
    [GTP5G_PDR_QER_ID]                          = { .type = NLA_U32, },
};

static const struct nla_policy gtp5g_genl_far_policy[GTP5G_FAR_ATTR_MAX + 1] = {
    [GTP5G_FAR_ID]                              = { .type = NLA_U32, },
    [GTP5G_FAR_APPLY_ACTION]                    = { .type = NLA_U8, },
    [GTP5G_FAR_FORWARDING_PARAMETER]            = { .type = NLA_NESTED, },
};

static const struct nla_policy gtp5g_genl_qer_policy[GTP5G_QER_ATTR_MAX + 1] = {
    [GTP5G_QER_ID]                              = { .type = NLA_U32, },
    [GTP5G_QER_GATE]                            = { .type = NLA_U8, },
    [GTP5G_QER_MBR]                             = { .type = NLA_NESTED, },
    [GTP5G_QER_GBR]                             = { .type = NLA_NESTED, },
    [GTP5G_QER_CORR_ID]                         = { .type = NLA_U32, },
    [GTP5G_QER_RQI]                             = { .type = NLA_U8, },
    [GTP5G_QER_QFI]                             = { .type = NLA_U8, },
    [GTP5G_QER_PPI]                             = { .type = NLA_U8, },
    [GTP5G_QER_RCSR]                            = { .type = NLA_U8, },
};

static const struct nla_policy gtp5g_genl_urr_policy[GTP5G_URR_ATTR_MAX + 1] = {
    [GTP5G_URR_ID]                              = { .type = NLA_U32, },
    [GTP5G_URR_MEASUREMENT_METHOD]              = { .type = NLA_U8, },
    [GTP5G_URR_REPORTING_TRIGGER]               = { .type = NLA_U32, },
    [GTP5G_URR_MEASUREMENT_PERIOD]              = { .type = NLA_U32, },
    [GTP5G_URR_MEASUREMENT_INFO]                = { .type = NLA_U8, },
    [GTP5G_URR_SEID]                            = { .type = NLA_U64, },
    [GTP5G_URR_VOLUME_THRESHOLD]                = { .type = NLA_NESTED, },
    [GTP5G_URR_VOLUME_QUOTA]                    = { .type = NLA_NESTED, },
};

static const struct genl_ops gtp5g_genl_ops[] = {
    {
        .cmd = GTP5G_CMD_ADD_PDR,
        // .validate = GENL_DONT_VALIDATE_STRICT | GENL_DONT_VALIDATE_DUMP,
        .doit = gtp5g_genl_add_pdr,
        // .policy = gtp5g_genl_pdr_policy,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_DEL_PDR,
        // .validate = GENL_DONT_VALIDATE_STRICT | GENL_DONT_VALIDATE_DUMP,
        .doit = gtp5g_genl_del_pdr,
        // .policy = gtp5g_genl_pdr_policy,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_GET_PDR,
        // .validate = GENL_DONT_VALIDATE_STRICT | GENL_DONT_VALIDATE_DUMP,
        .doit = gtp5g_genl_get_pdr,
        .dumpit = gtp5g_genl_dump_pdr,
        // .policy = gtp5g_genl_pdr_policy,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_ADD_FAR,
        // .validate = GENL_DONT_VALIDATE_STRICT | GENL_DONT_VALIDATE_DUMP,
        .doit = gtp5g_genl_add_far,
        // .policy = gtp5g_genl_far_policy,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_DEL_FAR,
        // .validate = GENL_DONT_VALIDATE_STRICT | GENL_DONT_VALIDATE_DUMP,
        .doit = gtp5g_genl_del_far,
        // .policy = gtp5g_genl_far_policy,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_GET_FAR,
        // .validate = GENL_DONT_VALIDATE_STRICT | GENL_DONT_VALIDATE_DUMP,
        .doit = gtp5g_genl_get_far,
        .dumpit = gtp5g_genl_dump_far,
        // .policy = gtp5g_genl_far_policy,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_ADD_QER,
        // .validate = GENL_DONT_VALIDATE_STRICT | GENL_DONT_VALIDATE_DUMP,
        .doit = gtp5g_genl_add_qer,
        // .policy = gtp5g_genl_qer_policy,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_DEL_QER,
        // .validate = GENL_DONT_VALIDATE_STRICT | GENL_DONT_VALIDATE_DUMP,
        .doit = gtp5g_genl_del_qer,
        // .policy = gtp5g_genl_qer_policy,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_GET_QER,
        // .validate = GENL_DONT_VALIDATE_STRICT | GENL_DONT_VALIDATE_DUMP,
        .doit = gtp5g_genl_get_qer,
        .dumpit = gtp5g_genl_dump_qer,
        // .policy = gtp5g_genl_qer_policy,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_ADD_URR,
        .doit = gtp5g_genl_add_urr,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_DEL_URR,
        .doit = gtp5g_genl_del_urr,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_GET_URR,
        .doit = gtp5g_genl_get_urr,
        .dumpit = gtp5g_genl_dump_urr,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_ADD_BAR,
        .doit = gtp5g_genl_add_bar,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_DEL_BAR,
        .doit = gtp5g_genl_del_bar,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_GET_BAR,
        .doit = gtp5g_genl_get_bar,
        .dumpit = gtp5g_genl_dump_bar,
        .flags = GENL_ADMIN_PERM,
    }, 
    {
        .cmd = GTP5G_CMD_GET_VERSION,
        .doit = gtp5g_genl_get_version,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_GET_REPORT,
        .doit = gtp5g_genl_get_usage_report,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_GET_MULTI_REPORTS,
        .doit = gtp5g_genl_get_multi_usage_reports,
        .flags = GENL_ADMIN_PERM,
    },
    {
        .cmd = GTP5G_CMD_GET_USAGE_STATISTIC,
        .doit = gtp5g_genl_get_usage_statistic,
        .flags = GENL_ADMIN_PERM,
    },
};

static const struct genl_multicast_group gtp5g_genl_mcgrps[] = {
	[GTP5G_GENL_MCGRP] = { .name = "gtp5g" },
};

struct genl_family gtp5g_genl_family __ro_after_init = {
    .name       = "gtp5g",
    .version    = 0,
    .hdrsize    = 0,
    .maxattr    = GTP5G_ATTR_MAX,
    .netnsok    = true,
    .module     = THIS_MODULE,
    .ops        = gtp5g_genl_ops,
    .n_ops      = ARRAY_SIZE(gtp5g_genl_ops),
    .mcgrps     = gtp5g_genl_mcgrps,
    .n_mcgrps   = ARRAY_SIZE(gtp5g_genl_mcgrps),
#if LINUX_VERSION_CODE >= KERNEL_VERSION(6, 0, 0)
    .resv_start_op = GTP5G_ATTR_MAX,
#endif
};
