//go:build ignore

#include "bpf_endian.h"
#include "common.h"
#include "linux/tcp.h"
#include "arpa/inet.h"

char __license[] SEC("license") = "Dual MIT/GPL";

#define MAX_MAP_ENTRIES 16

/* DMap to store the target destination port */
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u16); // Destination port
} target_port_map SEC(".maps");


/*
Compare TCP destination port from the packet with target port.
*/
static __always_inline int check_tcp_dst_port(struct xdp_md *ctx, __u16 *tcp_dst_port) {
	void *data_end = (void *)(long)ctx->data_end;
	void *data     = (void *)(long)ctx->data;

	// First, parse the ethernet header.
	struct ethhdr *eth = data;
	if ((void *)(eth + 1) > data_end) {
		return 0;
	}

	if (eth->h_proto != bpf_htons(ETH_P_IP)) {
		// The protocol is not IPv4.
		return 0;
	}

	// Then parse the IP header.
	struct iphdr *ip = (void *)(eth + 1);
	if ((void *)(ip + 1) > data_end) {
		return 0;
	}

	if (ip->protocol != IPPROTO_TCP) {
		return 0;
	}

	// Then parse the TCP header.
	struct tcphdr *tcp = (void *)(ip + 1);
	if((void *)(tcp + 1) > data_end) {
		// The protocol is not TCP, so we can't parse an TCP destination port.
		return 0;
	}

	// Return the TCP destination port.
	__u16 port = bpf_ntohs(tcp->dest);
	if (port != *tcp_dst_port) {
		return 0;
	}

	return 1; // port matched
}

SEC("xdp")
int xdp_prog_func(struct xdp_md *ctx) {
    __u16 target_port;
    __u32 key = 0;

    __u16 *target_port_ptr = bpf_map_lookup_elem(&target_port_map, &key);

    // Check if target_port_ptr is NULL.
    if (!target_port_ptr) {
        // If the target port hasn't been set by user-space, pass all packets.
        goto done;
    }

    target_port = *target_port_ptr;

	if (!check_tcp_dst_port(ctx, &target_port)) {
		// Port doesn't match, so allow and don't count.
		goto done;
	}

	// Drop the packet
	return XDP_DROP;

done:
	return XDP_PASS;
}