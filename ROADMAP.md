# Chaosd Roadmap

This document defines the roadmap for Chaosd development.

## v2.0

- [ ] Support HTTP attack to simulate HTTP faults, such as abort connection, delay, etc.
- [ ] Support IO attack to simulate file system faults, such as IO delay and read/write errors.
- [ ] Support workflow to manage a group of chaos experiments.
- [ ] Support use Dashboard to manage chaos experiments.
- [ ] Support time skew.
- [ ] JVM Attack supports fault injection for applications including MySQL, PostgreSQL, RoketMQ, etc.
- [ ] Support file attack, including deleting files, appending data to files, renaming files, modifying files' privilege.
- [ ] Support inject faults into Kafka, including high payload, the message queue is full, unable to write into, etc.
- [ ] Support inject faults into Zookeeper, including fail to receive publisher's message, fail to update config, fail to notify the update, etc.
- [ ] Support inject faults into Redis, including switch primary and secondary, cache hotspot, the sentinel is unavailable, sentinel restart, etc.