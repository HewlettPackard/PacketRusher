#include "log.h"

int dbg_trace_lvl = 1;

int get_dbg_lvl(){
    return dbg_trace_lvl;
}

void set_dbg_lvl(int val){
    dbg_trace_lvl = val;
}