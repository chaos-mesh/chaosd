# chaosd

Chaosd is an easy-to-use Chaos Engineering tool. This tool is used to inject failures to the physic node, such as kill process, network failure, CPU burn, memory burn and etc.

## Prerequisites

Before deploying Chaosd, make sure the following items have been installed:

* [tc](https://linux.die.net/man/8/tc)
* [ipset](https://linux.die.net/man/8/ipset)
* [iptables](https://linux.die.net/man/8/iptables)

## Install

### Build from source code

```bash
make chaosd
chmod +x chaosd && mv chaosd /usr/local/bin/chaosd
```

### Download binary

```bash
curl -fsSL -o chaosd https://mirrors.chaos-mesh.org/latest/chaosd
chmod +x chaosd && mv chaosd /usr/local/bin/chaosd
```

## Usage

### Process attack

* kill process

```bash
$ chaosd attack process kill -p [pid] # set pid or pod name
# remember the generated uid, we need this uid to recover chaos attack
Attack network successfully, uid: 2c865e6f-299f-4adf-ab37-94dc4fb8fea6
```

* stop process

```bash
$ chaosd attack process stop -p [pid] # set pid or pod name
```

### Network attack

* delay network packet

```bash
$ chaosd attack network delay -d eth0 -i 172.16.4.4 -l 10ms
```

* loss network packet

```bash
$ chaosd attack network loss -d eth0 -i 172.16.4.4 --percent 50%
```

* corrupt network packet

```bash
$ chaosd attack network corrupt -d eth0 -i 172.16.4.4 --percent 50%
```

* duplicate network packet

```bash
$ chaosd attack network duplicate -d eth0 -i 172.16.4.4 --percent 50%
```

### Recover network

```bash
$ chaosd recover 2c865e6f-299f-4adf-ab37-94dc4fb8fea6
```
