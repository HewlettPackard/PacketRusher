#ifndef __GTP5G_LOG_H__
#define __GTP5G_LOG_H__

#include <linux/netdevice.h>

#define DBG(level, dev, fmt, args...) do {      \
    if (level <= get_dbg_lvl()) {               \
        if (dev)                                \
            printk_ratelimited("%s:[gtp5g] %s: "fmt, netdev_name(dev), __func__, ##args);   \
        else                                    \
            printk_ratelimited("[gtp5g] %s: " fmt, __func__, ##args);       \
    } \
} while(0)

#define GTP5G_LOG(dev, fmt, args...) DBG(0, dev, fmt, ##args)
#define GTP5G_ERR(dev, fmt, args...) DBG(1, dev, fmt, ##args)
#define GTP5G_WAR(dev, fmt, args...) DBG(2, dev, fmt, ##args)
#define GTP5G_INF(dev, fmt, args...) DBG(3, dev, fmt, ##args)
#define GTP5G_TRC(dev, fmt, args...) DBG(4, dev, fmt, ##args)

extern int get_dbg_lvl(void);
extern void set_dbg_lvl(int);

#endif // __GTP5G_LOG_H__