#ifndef __GTP5G_PROC_H__
#define __GTP5G_PROC_H__

#include <linux/rculist.h>

extern int create_proc(void);
extern void remove_proc(void);

extern void init_proc_gtp5g_dev_list(void);
extern struct list_head * get_proc_gtp5g_dev_list_head(void);

#endif // __GTP5G_PROC_H__
