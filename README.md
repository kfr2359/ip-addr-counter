# IP address counter

This is a solution for https://github.com/Ecwid/new-job/blob/master/IP-Addr-Counter-GO.md

## Results
Testing machine: MacBook Air 2020 M1 8 cores, 16 GB RAM

### file.Read, parallel loading into bitmap, 30 workers, 1MB read buffer:
Peak memory uasge 1.6 GB (memory isn't freed to OS immediately by Go, then memory drops to ~1.1 GB)
Avg reading 745 MB/sec
CPU load ~400%
```
IP addresses were loaded in bitmap for 2m41.468659834s
IP addresses were counted for 40.643666ms
Unique IP addresses count: 1000000000
```