#include <linux/types.h>
#include "api_version.h"

bool api_far_action_u16 = false;

void set_far_action_u16(bool val){
    api_far_action_u16 = val;
}

bool far_action_is_u16(){
    return api_far_action_u16;
}