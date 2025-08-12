# IP address counter

This is a solution for https://github.com/Ecwid/new-job/blob/master/IP-Addr-Counter-GO.md

## Solution description
The solution is to create a bitmap for all possible IP addresses where each bit represents 1 unique IP address, then count ones in it.

Total number of IPv4 addresses is 2^32, that means size of the bitmap is 2^29 bytes ~ 537 MB. This amount of memory is occupied in any case.

Loading of 1 IP address takes O(1) - calculates offset in bitmap and set 1 bit. Parallel loading is possible as long as the access to single bitmap element (uint64, machine word) is atomic.

Counting addresses after loading takes O(1) - iterate the bitmap and count ones in each byte. The constant here is big but still a constant, moreover counting bits in sequental bytes is trivial operation, asm instruction including vector ones can be utilized.

## Results
Testing machine: MacBook Air 2020 M1 8 cores, 16 GB RAM
Input file: https://ecwid-vgv-storage.s3.eu-central-1.amazonaws.com/ip_addresses.zip (120 GB uncompressed, was taken from the task)

### file.Read, parallel loading into bitmap, 30 workers, 1MB read buffer:
Peak memory uasge 1.6 GB (memory isn't freed to OS immediately by Go, then memory drops to ~1.1 GB)
Avg reading 784 MB/sec
CPU load ~400%
```
IP addresses were loaded in bitmap for 2m33.461633542s
IP addresses were counted for 30.329083ms
Unique IP addresses count: 1000000000
```