# IP address counter

This is a solution for https://github.com/Ecwid/new-job/blob/master/IP-Addr-Counter-GO.md

## Results
Testing machine: MacBook Air 2020 M1 8 cores, 16 GB RAM

### file.Read, parallel loading into bitmap, 20 workers, 1MB read buffer:
Peak memory uasge 1.6 GB (memory isn't freed to OS immediately by Go, then memory drops to 1.13 GB)
Avg reading 594 MB/sec
CPU bottleneck was reached (~700%)
```
IP addresses were loaded in bitmap for 3m22.510120167s
IP addresses were counted for 58.81475ms
Unique IP addresses count: 1000000000
```