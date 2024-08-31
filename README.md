# icmperf

ICMP based performance tool

![example](./docs/icmperf.gif)

## Isn't ICMP terrible for performance testing?

Yes it is, at least in some cases.
ICMP is a protocol that is used for connectivity diagnostics and is not designed for performance testing.
ICMP packets are often given low priority by network devices and can be dropped or delayed.
This can lead to inaccurate results when using ICMP for performance testing.
ICMP packets can follow a different path than TCP or UDP packets, which can result in misleading figures.

For that reason, TCP or UDP based tools like [iperf](https://iperf.fr/) are better suited for performance testing.
And actually, if you don't fall into one of the cases below, I would recommend using `iperf` over `icmperf`.

## So why icmperf?

There are a few reasons why you might still want to use `icmperf`:
- ICMP is often allowed through firewalls and routers, while other protocols are not.
- Most devices will respond to ICMP packets.
You won't need to worry about setting up a server to respond to your packets. Actually you don't even need to have access to the target.
- If both devices are not busy during the test, there's few chances that ICMP packets will be dropped or delayed.
- If you're using some sort of tunneling, like a VPN, ICMP packets will follow the same path as your other packets.
They will also benefit from the same priority, except at the very ends of the tunnel.

So in short, `icmperf` is a good tool to get a rough idea of the performance of a network link, especially when you can't use other protocols, or don't want to bother setting up a server.

## Usage

```shell
Usage: icmperf <target> [flags]

Arguments:
  <target>    The target host to ping.

Flags:
  -h, --help                   Show context-sensitive help.
  -m, --mtu=1500               The maximum transmission unit of your interface.
  -t, --timeout=5s             The timeout for each ping.
  -d, --duration=30s           The duration of the test.
  -l, --bind-addr="0.0.0.0"    The address to bind the ICMP listener to.
```
