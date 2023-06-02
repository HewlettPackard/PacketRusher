#include <linux/module.h>

#include "genl.h"
#include "link.h"
#include "net.h"
#include "proc.h"
#include "hash.h"
#include "log.h"
#include "genl_version.h"

static int __init gtp5g_init(void)
{
    int err;

    GTP5G_LOG(NULL, "Gtp5g Module initialization Ver: %s\n", DRV_VERSION);

    init_proc_gtp5g_dev_list();

    // set hash initial value
    get_random_bytes(&gtp5g_h_initval, sizeof(gtp5g_h_initval));

    err = rtnl_link_register(&gtp5g_link_ops);
    if (err < 0) {
        GTP5G_ERR(NULL, "Failed to register rtnl\n");
        goto error_out;
    }

    err = genl_register_family(&gtp5g_genl_family);
    if (err < 0) {
        GTP5G_ERR(NULL, "Failed to register generic\n");
        goto unreg_rtnl_link;
    }

    err = register_pernet_subsys(&gtp5g_net_ops);
    if (err < 0) {
        GTP5G_ERR(NULL, "Failed to register namespace\n");
        goto unreg_genl_family;
    }

    err = create_proc();
    if (err < 0) {
        goto unreg_pernet;
    }
    GTP5G_LOG(NULL, "5G GTP module loaded\n");

    return 0;
unreg_pernet:
    unregister_pernet_subsys(&gtp5g_net_ops);
unreg_genl_family:
    genl_unregister_family(&gtp5g_genl_family);
unreg_rtnl_link:
    rtnl_link_unregister(&gtp5g_link_ops);
error_out:
    return err;
}

static void __exit gtp5g_fini(void)
{
    genl_unregister_family(&gtp5g_genl_family);
    rtnl_link_unregister(&gtp5g_link_ops);
    unregister_pernet_subsys(&gtp5g_net_ops);

    remove_proc();

    GTP5G_LOG(NULL, "5G GTP module unloaded\n");
}

late_initcall(gtp5g_init);
module_exit(gtp5g_fini);

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Yao-Wen Chang <yaowenowo@gmail.com>");
MODULE_AUTHOR("Muthuraman <muthuramane.cs03g@g2.nctu.edu.tw>");
MODULE_DESCRIPTION("Interface for 5G GTP encapsulated traffic");
MODULE_VERSION(DRV_VERSION);
MODULE_ALIAS_RTNL_LINK("gtp5g");
MODULE_ALIAS_GENL_FAMILY("gtp5g");
