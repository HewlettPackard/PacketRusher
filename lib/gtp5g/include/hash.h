#ifndef __HASH_H__
#define __HASH_H__

#include <linux/string.h>

#include <linux/jhash.h>

extern u32 gtp5g_h_initval;

static inline u32 u32_hashfn(u32 val)
{
    return jhash_1word(val, gtp5g_h_initval);
}

static inline u32 str_hashfn(char *str)
{
    return jhash(str, strlen(str), 0);
}

static inline u32 ipv4_hashfn(__be32 ip)
{
    return jhash_1word((__force u32)ip, gtp5g_h_initval);
}

#endif // __HASH_H__