#include <linux/version.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/rculist.h>

#include "proc.h"
#include "log.h"
#include "pdr.h"
#include "far.h"

#include "util.h"

struct list_head proc_gtp5g_dev;
struct proc_gtp5g_pdr {
    u16     id;
    u64     seid;
    u32     precedence;
    u8      ohr;
    u32     role_addr4;

    u32     pdi_ue_addr4;
    u32     pdi_fteid;
    u32     pdi_gtpu_addr4;
    
    u32     far_id;
    u32     *qer_ids;
    u32     qer_num;

    u32     *urr_ids;
    u32     urr_num;

    u64     ul_drop_cnt;
    u64     dl_drop_cnt;

    /* Packet Statistics */
    u64     ul_pkt_cnt;
    u64     dl_pkt_cnt;
    u64     ul_byte_cnt;
    u64     dl_byte_cnt;
};

struct proc_gtp5g_far {
    u32     id;
    u64     seid;
    u16     action;

    //OHC
    u16     description;
    u32     teid; 
    u32     peer_addr4;
};

struct proc_gtp5g_qer {
    u32     id;
    u64     seid;
    u8      qfi;
};

struct proc_gtp5g_urr {
    u32     id;
    u64     seid;
    u8      method;
    u32     trigger;
    u32     period;
    u8      info;

    u8      volth_flag;
    u64     volth_tolvol;
    u64     volth_ulvol;
    u64     volth_dlvol;

    u8      volqu_flag;
    u64     volqu_tolvol;
    u64     volqu_ulvol;
    u64     volqu_dlvol;

    s64     start_time;
    s64     end_time;
};

struct proc_gtp5g_qos 
{
    bool qos_enable;
};

struct proc_gtp5g_seq
{
    bool seq_enable;
};

struct proc_dir_entry *proc_gtp5g = NULL;
struct proc_dir_entry *proc_gtp5g_dbg = NULL;
struct proc_dir_entry *proc_gtp5g_pdr = NULL;
struct proc_dir_entry *proc_gtp5g_far = NULL;
struct proc_dir_entry *proc_gtp5g_qer = NULL;
struct proc_dir_entry *proc_gtp5g_urr = NULL;
struct proc_dir_entry *proc_gtp5g_qos = NULL;
struct proc_dir_entry *proc_gtp5g_seq = NULL;
struct proc_gtp5g_pdr proc_pdr;
struct proc_gtp5g_far proc_far;
struct proc_gtp5g_qer proc_qer;
struct proc_gtp5g_urr proc_urr;
struct proc_gtp5g_qos proc_qos;
struct proc_gtp5g_seq proc_seq;

u64 proc_seid = 0;
u16 proc_pdr_id = 0;
u32 proc_far_id = 0;
u32 proc_qer_id = 0;
u32 proc_urr_id = 0;

struct list_head * get_proc_gtp5g_dev_list_head(){
    return &proc_gtp5g_dev;
}

void init_proc_gtp5g_dev_list(){
    INIT_LIST_HEAD(&proc_gtp5g_dev);
}

static int gtp5g_dbg_read(struct seq_file *s, void *v) 
{
    seq_printf(s, "gtp5g kerenl debug level range: 0~4\n");
    seq_printf(s, "\t 0 -> Logging\n");
    seq_printf(s, "\t 1 -> Error(default)\n");
    seq_printf(s, "\t 2 -> Warning\n");
    seq_printf(s, "\t 3 -> Information\n");
    seq_printf(s, "\t 4 -> Trace\n");
    seq_printf(s, "Current: %d\n", get_dbg_lvl());
    return 0;
}

static ssize_t proc_dbg_write(struct file *filp, const char __user *buffer,
    size_t len, loff_t *dptr) 
{
    char buf[16];
    unsigned long buf_len = min(len, sizeof(buf) - 1);
    int dbg;

    if (copy_from_user(buf, buffer, buf_len)) {
        GTP5G_ERR(NULL, "Failed to read buffer: %s\n", buffer);
        goto err;
    }
    
    buf[buf_len] = 0;
    if (sscanf(buf, "%d", &dbg) != 1) {
        GTP5G_ERR(NULL, "Failed to read debug level: %s\n", buffer);
        goto err;
    }
  
    if (dbg < 0 || dbg > 4) {
        GTP5G_ERR(NULL, "Failed to set debug level: %d <0 or >4\n", dbg);
        goto err;
    }
    
    set_dbg_lvl(dbg);
    return strnlen(buf, buf_len);
err:
    return -1;
}

static int proc_dbg_read(struct inode *inode, struct file *file)
{
    return single_open(file, gtp5g_dbg_read, NULL);
}

static void set_pdr_qer_ids(char *pdr_qer_ids, struct proc_gtp5g_pdr *proc_pdr)
{
    int i = 0, len = 0;
    for (i = 0; i < proc_pdr->qer_num; i++) {
        len += sprintf(&pdr_qer_ids[len], "0x%x, ", proc_pdr->qer_ids[i]);
    }
}

static void set_pdr_urr_ids(char *pdr_urr_ids, struct proc_gtp5g_pdr *proc_pdr)
{
    int i = 0, len = 0;
    for (i = 0; i < proc_pdr->urr_num; i++) {
        len += sprintf(&pdr_urr_ids[len], "0x%x, ", proc_pdr->urr_ids[i]);
    }
}

static int gtp5g_pdr_read(struct seq_file *s, void *v) 
{
    char role_addr[35];
    char pdi_ue_addr[35];
    char pdu_gtpu_addr[35];
    char pdr_qer_ids[64];
    char pdr_urr_ids[64];

    if (!proc_pdr_id) {
        seq_printf(s, "Given PDR ID does not exists\n");
        return -1;
    }
    
    seq_printf(s, "PDR: \n");
    seq_printf(s, "\t SEID : %llu\n", proc_pdr.seid);
    seq_printf(s, "\t ID : %u\n", proc_pdr.id);
    seq_printf(s, "\t Precedence: %u\n", proc_pdr.precedence);
    seq_printf(s, "\t OHR: %u\n", proc_pdr.ohr);
    ip_string(role_addr, proc_pdr.role_addr4);
    seq_printf(s, "\t Role Addr4: %s(%#08x)\n", role_addr, ntohl(proc_pdr.role_addr4));
    ip_string(pdi_ue_addr, proc_pdr.pdi_ue_addr4);
    seq_printf(s, "\t PDI UE Addr4: %s(%#08x)\n", pdi_ue_addr, ntohl(proc_pdr.pdi_ue_addr4));
    seq_printf(s, "\t PDI TEID: %#08x\n", ntohl(proc_pdr.pdi_fteid));
    ip_string(pdu_gtpu_addr, proc_pdr.pdi_gtpu_addr4);
    seq_printf(s, "\t PDU GTPU Addr4: %s(%#08x)\n", pdu_gtpu_addr, ntohl(proc_pdr.pdi_gtpu_addr4));
    seq_printf(s, "\t FAR ID: %u\n", proc_pdr.far_id);
    set_pdr_qer_ids(pdr_qer_ids, &proc_pdr);
    seq_printf(s, "\t QER IDs: %s\n", pdr_qer_ids);
    set_pdr_urr_ids(pdr_urr_ids, &proc_pdr);
    seq_printf(s, "\t URR IDs: %s\n", pdr_urr_ids);
    seq_printf(s, "\t UL Drop Count: %#llx\n", proc_pdr.ul_drop_cnt);
    seq_printf(s, "\t DL Drop Count: %#llx\n", proc_pdr.dl_drop_cnt);
    seq_printf(s, "\t UL Packet Count: %llu\n", proc_pdr.ul_pkt_cnt);
    seq_printf(s, "\t DL Packet Count: %llu\n", proc_pdr.dl_pkt_cnt);
    seq_printf(s, "\t UL Byte Count: %llu\n", proc_pdr.ul_byte_cnt);
    seq_printf(s, "\t DL Byte Count: %llu\n", proc_pdr.dl_byte_cnt);
    return 0;
}

static int gtp5g_far_read(struct seq_file *s, void *v) 
{
    if (!proc_far_id) {
        seq_printf(s, "Given FAR ID does not exists\n");
        return -1;
    }
    
    seq_printf(s, "FAR: \n");
    seq_printf(s, "\t SEID : %llu\n", proc_far.seid);
    seq_printf(s, "\t ID : %u\n", proc_far.id);
    seq_printf(s, "\t Apply Action: %#x\n", proc_far.action);
    seq_printf(s, "\t OHC Description: %#x\n", proc_far.description);
    seq_printf(s, "\t OHC TEID : %#08x\n", ntohl(proc_far.teid));
    seq_printf(s, "\t OHC Peer Addr4: %#08x\n", ntohl(proc_far.peer_addr4));
    return 0;
}

static int gtp5g_qer_read(struct seq_file *s, void *v) 
{
    if (!proc_qer_id) {
        seq_printf(s, "Given QER ID does not exists\n");
        return -1;
    }
    
    seq_printf(s, "QER: \n");
    seq_printf(s, "\t SEID : %llu\n", proc_qer.seid);
    seq_printf(s, "\t ID : %u\n", proc_qer.id);
    seq_printf(s, "\t QFI: %u\n", proc_qer.qfi);
    return 0;
}

static int gtp5g_urr_read(struct seq_file *s, void *v)
{
    if (!proc_urr_id) {
        seq_printf(s, "Given URR ID does not exists\n");
        return -1;
    }

    seq_printf(s, "URR: \n");
    seq_printf(s, "\t SEID : %llu\n", proc_urr.seid);
    seq_printf(s, "\t ID : %u\n", proc_urr.id);
    seq_printf(s, "\t Measurement method : 0x%02x\n", proc_urr.method);
    seq_printf(s, "\t Reporting trigger : 0x%06x\n", proc_urr.trigger);
    seq_printf(s, "\t Measurement period : %d\n", proc_urr.period);
    seq_printf(s, "\t Measurement information : 0x%02x\n", proc_urr.info);
    seq_printf(s, "\t Volume threshold flag: 0x%02x\n", proc_urr.volth_flag);
    seq_printf(s, "\t Volume threshold toltal volume: %llu\n", proc_urr.volth_tolvol);
    seq_printf(s, "\t Volume threshold uplink volume: %llu\n", proc_urr.volth_ulvol);
    seq_printf(s, "\t Volume threshold downlink volume: %llu\n", proc_urr.volth_dlvol);

    seq_printf(s, "\t Volume quota flag: %02x\n", proc_urr.volqu_flag);
    seq_printf(s, "\t Volume quota toltal volume: %llu\n", proc_urr.volqu_tolvol);
    seq_printf(s, "\t Volume quota uplink volume: %llu\n", proc_urr.volqu_ulvol);
    seq_printf(s, "\t Volume quota downlink volume: %llu\n", proc_urr.volqu_dlvol);

    seq_printf(s, "\t Start time: %lld\n", proc_urr.start_time);
    seq_printf(s, "\t End time: %lld\n", proc_urr.end_time);

    return 0;
}

static int gtp5g_qos_read(struct seq_file *s, void *v)
{
    GTP5G_TRC(NULL, "gtp5g_qos_read");
    seq_printf(s, "QoS Enable: %d\n", get_qos_enable());
    return 0;
}

static ssize_t proc_qos_write(struct file *filp, const char __user *buffer,
    size_t len, loff_t *dptr) 
{
    char buf[16];
    unsigned long buf_len = min(len, sizeof(buf) - 1);
    int qos_enable;

    if (copy_from_user(buf, buffer, buf_len)) {
        GTP5G_ERR(NULL, "Failed to read buffer: %s\n", buffer);
        goto err;
    }
    
    buf[buf_len] = 0;
    if (sscanf(buf, "%d", &qos_enable) != 1) {
        GTP5G_ERR(NULL, "Failed to read qos enable setting: %s\n", buffer);
        goto err;
    }
     
    set_qos_enable(qos_enable);
    GTP5G_TRC(NULL, "qos enable:%d", get_qos_enable());
    return strnlen(buf, buf_len);
err:
    return -1;
}


static int gtp5g_seq_read(struct seq_file *s, void *v)
{
    GTP5G_TRC(NULL, "gtp5g_seq_read");
    seq_printf(s, "Sequence Number Enable: %d\n", get_seq_enable());
    return 0;
}

static ssize_t proc_seq_write(struct file *filp, const char __user *buffer,
    size_t len, loff_t *dptr) 
{
    char buf[16];
    unsigned long buf_len = min(len, sizeof(buf) - 1);
    int seq_enable;

    if (copy_from_user(buf, buffer, buf_len)) {
        GTP5G_ERR(NULL, "Failed to read buffer: %s\n", buffer);
        goto err;
    }
    
    buf[buf_len] = 0;
    if (sscanf(buf, "%d", &seq_enable) != 1) {
        GTP5G_ERR(NULL, "Failed to read seq enable setting: %s\n", buffer);
        goto err;
    }
     
    set_seq_enable(seq_enable);
    GTP5G_TRC(NULL, "seq enable:%d", get_seq_enable());
    return strnlen(buf, buf_len);
err:
    return -1;
}

static ssize_t proc_pdr_write(struct file *filp, const char __user *buffer,
    size_t len, loff_t *dptr)
{
    char buf[128], dev_name[32];
    u8 found = 0;
    unsigned long buf_len = min(sizeof(buf) - 1, len);
    struct pdr *pdr;
    struct gtp5g_dev *gtp;

    if (copy_from_user(buf, buffer, buf_len)) {
        GTP5G_ERR(NULL, "Failed to read buffer: %s\n", buf);
        goto err;
    }

    buf[buf_len] = 0;
    if (sscanf(buf, "%s %llu %hu", dev_name, &proc_seid, &proc_pdr_id) != 3) {
        GTP5G_ERR(NULL, "proc write of PDR Dev & ID: %s is not valid\n", buf);
        goto err;
    }

    list_for_each_entry_rcu(gtp, &proc_gtp5g_dev, proc_list) {
        if (strcmp(dev_name, netdev_name(gtp->dev)) == 0) {
            found = 1;
            break;
        }
    }
    if (!found) {
        GTP5G_ERR(NULL, "Given dev: %s not exists\n", dev_name);
        goto err;
    }

    pdr = find_pdr_by_id(gtp, proc_seid, proc_pdr_id);
    if (!pdr) {
        GTP5G_ERR(NULL, "Given SEID : %llu PDR ID : %u not exists\n", proc_seid, proc_pdr_id);
        goto err;
    }

    memset(&proc_pdr, 0, sizeof(proc_pdr));
    proc_pdr.id = pdr->id;
    proc_pdr.seid = pdr->seid;
    proc_pdr.precedence = pdr->precedence;

    if (pdr->outer_header_removal) 
        proc_pdr.ohr = *pdr->outer_header_removal;
    
    if (pdr->role_addr_ipv4.s_addr)
        proc_pdr.role_addr4 = pdr->role_addr_ipv4.s_addr;
    
    if (pdr->pdi) {
        if (pdr->pdi->ue_addr_ipv4) 
            proc_pdr.pdi_ue_addr4 = pdr->pdi->ue_addr_ipv4->s_addr;
        if (pdr->pdi->f_teid) {
            proc_pdr.pdi_fteid = pdr->pdi->f_teid->teid;
            proc_pdr.pdi_gtpu_addr4 = pdr->pdi->f_teid->gtpu_addr_ipv4.s_addr;
        }
    }

    if (pdr->far_id)
        proc_pdr.far_id = *pdr->far_id;

    if (pdr->qer_ids) {
        proc_pdr.qer_ids = pdr->qer_ids;
        proc_pdr.qer_num = pdr->qer_num;
    }

    if (pdr->urr_ids) {
        proc_pdr.urr_ids = pdr->urr_ids;
        proc_pdr.urr_num = pdr->urr_num;
    }

    proc_pdr.ul_drop_cnt = pdr->ul_drop_cnt;
    proc_pdr.dl_drop_cnt = pdr->dl_drop_cnt;

    proc_pdr.ul_pkt_cnt = pdr->ul_pkt_cnt;
    proc_pdr.dl_pkt_cnt = pdr->dl_pkt_cnt;
    proc_pdr.ul_byte_cnt = pdr->ul_byte_cnt;
    proc_pdr.dl_byte_cnt = pdr->dl_byte_cnt;

    return strnlen(buf, buf_len);
err:
    proc_pdr_id = 0;
    return -1;
}

static ssize_t proc_far_write(struct file *filp, const char __user *buffer,
    size_t len, loff_t *dptr) 
{
    char buf[128], dev_name[32];
    u8 found = 0;
    unsigned long buf_len = min(sizeof(buf) - 1, len);
    struct far *far;
    struct gtp5g_dev *gtp;
    struct forwarding_parameter *fwd_param;

    if (copy_from_user(buf, buffer, buf_len)) {
        GTP5G_ERR(NULL, "Failed to read buffer: %s\n", buf);
        goto err;
    }

    buf[buf_len] = 0;
    if (sscanf(buf, "%s %llu %u", dev_name, &proc_seid, &proc_far_id) != 3) {
        GTP5G_ERR(NULL, "proc write of FAR Dev & ID: %s is not valid\n", buf);
        goto err;
    }

    list_for_each_entry_rcu(gtp, &proc_gtp5g_dev, proc_list) {
        if (strcmp(dev_name, netdev_name(gtp->dev)) == 0) {
            found = 1;
            break;
        }
    }
    if (!found) {
        GTP5G_ERR(NULL, "Given dev: %s not exists\n", dev_name);
        goto err;
    }

    far = find_far_by_id(gtp, proc_seid, proc_far_id);
    if (!far) {
        GTP5G_ERR(NULL, "Given FAR ID : %u not exists\n", proc_far_id);
        goto err;
    }

    memset(&proc_far, 0, sizeof(proc_far));
    proc_far.id = far->id;
    proc_far.seid = far->seid;
    proc_far.action = far->action;

    fwd_param = rcu_dereference(far->fwd_param);
    if (fwd_param) {
        if (fwd_param->hdr_creation) {
            proc_far.description = fwd_param->hdr_creation->description;
            proc_far.teid = fwd_param->hdr_creation->teid;
            proc_far.peer_addr4 = fwd_param->hdr_creation->peer_addr_ipv4.s_addr;
        }
    }

    return strnlen(buf, buf_len);
err:
    proc_far_id = 0;
    return -1;
}

static ssize_t proc_qer_write(struct file *filp, const char __user *buffer,
    size_t len, loff_t *dptr) 
{
    char buf[128], dev_name[32];
    u8 found = 0;
    unsigned long buf_len = min(sizeof(buf) - 1, len);
    struct qer *qer;
    struct gtp5g_dev *gtp;

    if (copy_from_user(buf, buffer, buf_len)) {
        GTP5G_ERR(NULL, "Failed to read buffer: %s\n", buf);
        goto err;
    }

    buf[buf_len] = 0;
    if (sscanf(buf, "%s %llu %u", dev_name, &proc_seid, &proc_qer_id) != 3) {
        GTP5G_ERR(NULL, "proc write of QER Dev & ID: %s is not valid\n", buf);
        goto err;
    }

    list_for_each_entry_rcu(gtp, &proc_gtp5g_dev, proc_list) {
        if (strcmp(dev_name, netdev_name(gtp->dev)) == 0) {
            found = 1;
            break;
        }
    }
    if (!found) {
        GTP5G_ERR(NULL, "Given dev: %s not exists\n", dev_name);
        goto err;
    }

    qer = find_qer_by_id(gtp, proc_seid, proc_qer_id);
    if (!qer) {
        GTP5G_ERR(NULL, "Given QER ID : %u not exists\n", proc_qer_id);
        goto err;
    }

    memset(&proc_qer, 0, sizeof(proc_qer));
    proc_qer.id = qer->id;
    proc_qer.seid = qer->seid;
    proc_qer.qfi = qer->qfi;

    return strnlen(buf, buf_len);
err:
    proc_qer_id = 0;
    return -1;
}

static ssize_t proc_urr_write(struct file *filp, const char __user *buffer,
    size_t len, loff_t *dptr) 
{
    char buf[128], dev_name[32];
    u8 found = 0;
    unsigned long buf_len = min(sizeof(buf) - 1, len);
    struct urr *urr;
    struct gtp5g_dev *gtp;

    if (copy_from_user(buf, buffer, buf_len)) {
        GTP5G_ERR(NULL, "Failed to read buffer: %s\n", buf);
        goto err;
    }

    buf[buf_len] = 0;
    if (sscanf(buf, "%s %llu %u", dev_name, &proc_seid, &proc_urr_id) != 3) {
        GTP5G_ERR(NULL, "proc write of URR Dev & ID: %s is not valid\n", buf);
        goto err;
    }

    list_for_each_entry_rcu(gtp, &proc_gtp5g_dev, proc_list) {
        if (strcmp(dev_name, netdev_name(gtp->dev)) == 0) {
            found = 1;
            break;
        }
    }
    if (!found) {
        GTP5G_ERR(NULL, "Given dev: %s not exists\n", dev_name);
        goto err;
    }

    urr = find_urr_by_id(gtp, proc_seid, proc_urr_id);
    if (!urr) {
        GTP5G_ERR(NULL, "Given URR ID : %u not exists\n", proc_urr_id);
        goto err;
    }

    memset(&proc_urr, 0, sizeof(proc_urr));
    proc_urr.id = urr->id;
    proc_urr.seid = urr->seid;
    proc_urr.method = urr->method;
    proc_urr.trigger = urr->trigger;
    proc_urr.period = urr->period;
    proc_urr.info = urr->info;

    proc_urr.volth_flag = urr->volumethreshold.flag;
    proc_urr.volth_tolvol = urr->volumethreshold.totalVolume;
    proc_urr.volth_ulvol = urr->volumethreshold.uplinkVolume;
    proc_urr.volth_dlvol = urr->volumethreshold.downlinkVolume;

    proc_urr.volqu_flag = urr->volumequota.flag;
    proc_urr.volqu_tolvol = urr->volumequota.totalVolume;
    proc_urr.volqu_ulvol = urr->volumequota.uplinkVolume;
    proc_urr.volqu_dlvol = urr->volumequota.downlinkVolume;

    proc_urr.start_time = urr->start_time;
    proc_urr.end_time = urr->end_time;

    return strnlen(buf, buf_len);
err:
    proc_urr_id = 0;
    return -1;
}

static int proc_pdr_read(struct inode *inode, struct file *file)
{
    return single_open(file, gtp5g_pdr_read, NULL);
}

static int proc_far_read(struct inode *inode, struct file *file)
{
    return single_open(file, gtp5g_far_read, NULL);
}

static int proc_qer_read(struct inode *inode, struct file *file)
{
    return single_open(file, gtp5g_qer_read, NULL);
}

static int proc_urr_read(struct inode *inode, struct file *file)
{
    return single_open(file, gtp5g_urr_read, NULL);
}

static int proc_qos_read(struct inode *inode, struct file *file)
{
    return single_open(file, gtp5g_qos_read, NULL);
}

static int proc_seq_read(struct inode *inode, struct file *file)
{
    return single_open(file, gtp5g_seq_read, NULL);
}

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 6, 0)
static const struct proc_ops proc_gtp5g_dbg_ops = {
    .proc_open = proc_dbg_read,
    .proc_read = seq_read,
    .proc_write = proc_dbg_write,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};
#else
static const struct file_operations proc_gtp5g_dbg_ops = {
    .owner      = THIS_MODULE,
    .open       = proc_dbg_read,
    .read       = seq_read,
    .write      = proc_dbg_write,
    .llseek     = seq_lseek,
    .release    = single_release,
};
#endif

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 6, 0)
static const struct proc_ops proc_gtp5g_pdr_ops = {
    .proc_open = proc_pdr_read,
    .proc_read = seq_read,
    .proc_write = proc_pdr_write,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};
#else
static const struct file_operations proc_gtp5g_pdr_ops = {
    .owner      = THIS_MODULE,
    .open       = proc_pdr_read,
    .read       = seq_read,
    .write      = proc_pdr_write,
    .llseek     = seq_lseek,
    .release    = single_release,
};
#endif

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 6, 0)
static const struct proc_ops proc_gtp5g_far_ops = {
    .proc_open = proc_far_read,
    .proc_read = seq_read,
    .proc_write = proc_far_write,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};
#else
static const struct file_operations proc_gtp5g_far_ops = {
    .owner      = THIS_MODULE,
    .open       = proc_far_read,
    .read       = seq_read,
    .write      = proc_far_write,
    .llseek     = seq_lseek,
    .release    = single_release,
};
#endif

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 6, 0)
static const struct proc_ops proc_gtp5g_qer_ops = {
    .proc_open = proc_qer_read,
    .proc_read = seq_read,
    .proc_write = proc_qer_write,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};
#else
static const struct file_operations proc_gtp5g_qer_ops = {
    .owner      = THIS_MODULE,
    .open       = proc_qer_read,
    .read       = seq_read,
    .write      = proc_qer_write,
    .llseek     = seq_lseek,
    .release    = single_release,
};
#endif

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 6, 0)
static const struct proc_ops proc_gtp5g_urr_ops = {
    .proc_open = proc_urr_read,
    .proc_read = seq_read,
    .proc_write = proc_urr_write,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};
#else
static const struct file_operations proc_gtp5g_urr_ops = {
    .owner      = THIS_MODULE,
    .open       = proc_urr_read,
    .read       = seq_read,
    .write      = proc_urr_write,
    .llseek     = seq_lseek,
    .release    = single_release,
};
#endif

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 6, 0)
static const struct proc_ops proc_gtp5g_qos_ops = {
    .proc_open = proc_qos_read,
    .proc_read = seq_read,
    .proc_write = proc_qos_write,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};
#else
static const struct file_operations proc_gtp5g_qos_ops = {
    .owner      = THIS_MODULE,
    .open       = proc_qos_read,
    .read       = seq_read,
    .write      = proc_qos_write,
    .llseek     = seq_lseek,
    .release    = single_release,
};
#endif

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 6, 0)
static const struct proc_ops proc_gtp5g_seq_ops = {
    .proc_open = proc_seq_read,
    .proc_read = seq_read,
    .proc_write = proc_seq_write,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};
#else
static const struct file_operations proc_gtp5g_seq_ops = {
    .owner      = THIS_MODULE,
    .open       = proc_seq_read,
    .read       = seq_read,
    .write      = proc_seq_write,
    .llseek     = seq_lseek,
    .release    = single_release,
};
#endif

int create_proc(void)
{
    proc_gtp5g = proc_mkdir("gtp5g", NULL);
    if (!proc_gtp5g) {
        GTP5G_ERR(NULL, "Failed to create /proc/gtp5g\n");
    }

    proc_gtp5g_dbg = proc_create("dbg", (S_IFREG | S_IRUGO | S_IWUGO),
        proc_gtp5g, &proc_gtp5g_dbg_ops);
    if (!proc_gtp5g_dbg) {
        GTP5G_ERR(NULL, "Failed to create /proc/gtp5g/dbg\n");
        goto remove_gtp5g_proc;
    }

    proc_gtp5g_pdr = proc_create("pdr", (S_IFREG | S_IRUGO | S_IWUGO),
        proc_gtp5g, &proc_gtp5g_pdr_ops);
    if (!proc_gtp5g_pdr) {
        GTP5G_ERR(NULL, "Failed to create /proc/gtp5g/pdr\n");
        goto remove_dbg_proc;
    }

    proc_gtp5g_far = proc_create("far", (S_IFREG | S_IRUGO | S_IWUGO),
        proc_gtp5g, &proc_gtp5g_far_ops);
    if (!proc_gtp5g_far) {
        GTP5G_ERR(NULL, "Failed to create /proc/gtp5g/far\n");
        goto remove_pdr_proc;
    }

    proc_gtp5g_qer = proc_create("qer", (S_IFREG | S_IRUGO | S_IWUGO),
        proc_gtp5g, &proc_gtp5g_qer_ops);
    if (!proc_gtp5g_qer) {
        GTP5G_ERR(NULL, "Failed to create /proc/gtp5g/qer\n");
        goto remove_far_proc;
    }

    proc_gtp5g_urr = proc_create("urr", (S_IFREG | S_IRUGO | S_IWUGO),
        proc_gtp5g, &proc_gtp5g_urr_ops);
    if (!proc_gtp5g_urr) {
        GTP5G_ERR(NULL, "Failed to create /proc/gtp5g/urr\n");
        goto remove_qer_proc;
    }

    proc_gtp5g_qos = proc_create("qos", (S_IFREG | S_IRUGO | S_IWUGO),
        proc_gtp5g, &proc_gtp5g_qos_ops);
    if (!proc_gtp5g_qos) {
        GTP5G_ERR(NULL, "Failed to create /proc/gtp5g/qos\n");
        goto remove_urr_proc;
    }

    proc_gtp5g_seq = proc_create("seq", (S_IFREG | S_IRUGO | S_IWUGO),
        proc_gtp5g, &proc_gtp5g_seq_ops);
    if (!proc_gtp5g_seq) {
        GTP5G_ERR(NULL, "Failed to create /proc/gtp5g/seq\n");
        goto remove_qos_proc;
    }

    return 0;

    remove_qos_proc:
        remove_proc_entry("qos", proc_gtp5g);
    remove_urr_proc:
        remove_proc_entry("urr", proc_gtp5g);
    remove_qer_proc:
        remove_proc_entry("qer", proc_gtp5g);
    remove_far_proc:
        remove_proc_entry("far", proc_gtp5g);
    remove_pdr_proc:
        remove_proc_entry("pdr", proc_gtp5g);
    remove_dbg_proc:
        remove_proc_entry("dbg", proc_gtp5g);
    remove_gtp5g_proc:
        remove_proc_entry("gtp5g", NULL);
    return -1;
}

void remove_proc()
{
    remove_proc_entry("qos", proc_gtp5g);
    remove_proc_entry("seq", proc_gtp5g);
    remove_proc_entry("urr", proc_gtp5g);
    remove_proc_entry("qer", proc_gtp5g);
    remove_proc_entry("far", proc_gtp5g);
    remove_proc_entry("pdr", proc_gtp5g);
    remove_proc_entry("dbg", proc_gtp5g);
    remove_proc_entry("gtp5g", NULL);
}
