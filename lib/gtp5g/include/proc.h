#ifndef __GTP5G_PROC_H__
#define __GTP5G_PROC_H__

#include <linux/rculist.h>

int create_proc(void);
void remove_proc(void);

void init_proc_gtp5g_dev_list(void);
struct list_head * get_proc_gtp5g_dev_list_head(void);

#endif // __GTP5G_PROC_H__
