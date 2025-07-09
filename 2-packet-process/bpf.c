// +build ignore

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_tracing.h>

char LICENSE[] SEC("license") = "Dual MIT/GPL";

#define EPERM 1
#define TASK_COMM_LEN 16

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u16); // allowed port
} port_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, char[TASK_COMM_LEN]); // target process
} process_map SEC(".maps");

SEC("lsm/socket_accept")
int BPF_PROG(check_accept, struct socket *sock, struct socket *newsock) {
    struct sock *sk;
    __u16 port;
    char comm[TASK_COMM_LEN];
    char proc_name[TASK_COMM_LEN];
    __u16 *allowed_port;
    __u16 family;

    __u32 key = 0;
    bpf_get_current_comm(&comm, sizeof(comm));

    allowed_port = bpf_map_lookup_elem(&port_map, &key);

    char *val = bpf_map_lookup_elem(&process_map, &key);
    if (val) {
        __builtin_memcpy(proc_name, val, TASK_COMM_LEN);
    } else {
        // No target process configured, allow all
        return 0;
    }

    // If current process name doesn't match target proc, allow
    if (__builtin_memcmp(proc_name, comm, TASK_COMM_LEN) != 0) {
        return 0;
    }
    
    sk = sock->sk;
    if (!sk) {
        return 0;
    }

    port = BPF_CORE_READ(sk, __sk_common.skc_num);

    if (allowed_port && port != *allowed_port) {
        return -EPERM; // drop connection
    }

    return 0; // allow connection
}
