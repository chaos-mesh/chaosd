# chaosd

Chaosd is an easy-to-use Chaos Engineering tool used to inject failures to a physical node.

## Prerequisites

Before deploying Chaosd, make sure the following items have been installed:

* [tc](https://linux.die.net/man/8/tc)
* [ipset](https://linux.die.net/man/8/ipset)
* [iptables](https://linux.die.net/man/8/iptables)
* [stress-ng](https://wiki.ubuntu.com/Kernel/Reference/stress-ng)

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

Chaosd supports two modes of failure injection:

-  **Command mode** - using Chaosd in the command line mode
-  **Server mode** - running Chaosd as a daemon server

- [Command mode](#command-mode)
    - [Process attack](#process-attack)
    - [Network attack](#network-attack)
    - [Stress attack](#stress-attack)
    - [Disk attack](#disk-attack)
    - [Host attack](#host-attack)
    - [Recover attack](#recover-attack)
- [Server mode](#server-mode)
    - [Process attack](#process-attack-1)
    - [Network attack](#network-attack-1)
    - [Stress attack](#stress-attack-1)
    - [Disk attack](#disk-attack-1)
    - [Recover attack](#recover-attack-1)

### Command mode

Using Chaosd as a command-line tool.

#### Process attack

Attack a process according to the PID or process name.

* kill process

Kill a process by sending `SIGKILL` signal:

```bash
$ chaosd attack process kill -p [pid] # set pid or pod name
# remember the generated uid, we need this uid to recover chaos attack
Attack network successfully, uid: 2c865e6f-299f-4adf-ab37-94dc4fb8fea6
```

* stop process

Stop a process by sending `SIGSTOP` signal:

```bash
$ chaosd attack process stop -p [pid] # set pid or pod name
```

#### Network attack

Attack the network by set the `iptables`, `ipset` and `tc`.

* delay network packet

Send message with specified latency:

```bash
$ chaosd attack network delay -d eth0 -i 172.16.4.4 -l 10ms
```

* loss network packet

Drop network packets randomly:

```bash
$ chaosd attack network loss -d eth0 -i 172.16.4.4 --percent 50%
```

* corrupt network packet

Make packet corruption:

```bash
$ chaosd attack network corrupt -d eth0 -i 172.16.4.4 --percent 50%
```

* duplicate network packet

Send packet duplicated:

```bash
$ chaosd attack network duplicate -d eth0 -i 172.16.4.4 --percent 50%
```

#### Stress attack

Generate plenty of stresses on the host:

* CPU stress

Generate CPU stresses on the host:

```bash
$ chaosd attack stress cpu -l 100 -w 2
```

* Memory stress

Generate memory stresses on the host:

```bash
$ chaosd attack stress mem -w 2 # stress 2 CPU and each cpu loads 100%
```

#### Disk attack

Attack the disk by increasing write/read payload, or filling the disk.

* Add payload

Add read payload:

```bash
./bin/chaosd attack disk add-payload read --path /tmp/temp --size 100
```

Add write payload:

```bash
./bin/chaosd attack disk add-payload write --path /tmp/temp --size 100
```

* Fill disk

Fill disk by fallocate:

```bash
./bin/chaosd attack disk fill --fallocate true --path /tmp/temp --size 100
```

Fill disk by write data to file:

```bash
./bin/chaosd attack disk fill --fallocate false --path /tmp/temp --size 100
```

#### Host attack

Shutdown the host:

```bash
./bin/chaosd attack host shutdown
```

> **Note:**
>
> This command will shutdown the host, make sure you know what the command means before you execute it.

#### Recover attack

Recover a attack:

```bash
$ chaosd recover 2c865e6f-299f-4adf-ab37-94dc4fb8fea6
```

### Server Mode

Run Chaosd as a daemon server:

```bash
nohup ./bin/chaosd server > chaosd.log 2>&1 &
```

And then you can inject failures by sending HTTP requests.

> **Note**:
>
> Make sure the user has the privileges to run iptables, ipset, etc. Or you can run chaosd with `sudo`.

#### Process attack

Attack a process according to the PID or process name.

* kill process

Kill a process by sending `SIGKILL` signal:

```bash
curl -X POST "127.0.0.1:31767/api/attack/process" -H "Content-Type: application/json"  -d '{"process": "{pid}", "signal": 9}' # set pid or pod name
{"status":200,"message":"attack successfully","uid":"e6d01a30-4528-4c70-b4fb-4dc47c4d39be"}
```

* stop process

Stop a process by sending `SIGSTOP` signal:

```bash
curl -X POST "127.0.0.1:31767/api/attack/process" -H "Content-Type: application/json"  -d '{"process": "{pid}", "signal": 15}' # set pid or pod name
{"status":200,"message":"attack successfully","uid":"ecf3f564-c4c0-4aaf-83c6-4b511a6e3a85"}
```

#### Network attack

Attack the network by set the `iptables`, `ipset` and `tc`.

* delay network packet

Send message with specified latency:

```bash
$ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "eth0", "ipaddress": "172.16.4.4", "action": "delay", "latency": "10ms", "jitter": "10ms", "correlation": "0"}'
```

* loss network packet

Drop network packets randomly:

```bash
$ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "eth0", "ipaddress": "172.16.4.4", "action": "loss", "percent": "50", "correlation": "0"}'
```

* corrupt network packet

Make packet corruption:

```bash
$ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "eth0", "ipaddress": "172.16.4.4", "action": "corrupt", "percent": "50",  "correlation": "0"}'
```

* duplicate network packet

Send packet duplicated:

```bash
$ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "eth0", "ipaddress": "172.16.4.4", "action": "duplicate", "percent": "50", "correlation": "0"}'
```

#### Stress attack

Generate plenty of stresses on the host:

* CPU stress

Generate CPU stresses on the host:

```bash
$ curl -X POST 127.0.0.1:31767/api/attack/stress -H "Content-Type:application/json" -d '{"action":"cpu", "load": 100, "worker": 2}'
```

* Memory stress

Generate memory stresses on the host:

```bash
$ curl -X POST 127.0.0.1:31767/api/attack/stress -H "Content-Type:application/json" -d '{"action":"mem", "worker": 2}'
```

#### Disk attack

Attack the disk by increasing write/read payload, or filling the disk.

* Add payload

Add read payload:

```bash
curl -X POST "127.0.0.1:31767/api/attack/disk" -H "Content-Type: application/json" -d '{"action":"read-payload","size":1024,"path":"temp"}'
```

Add write payload:

```bash
curl -X POST "127.0.0.1:31767/api/attack/disk" -H "Content-Type: application/json" -d '{"action":"write-payload","size":1024,"path":"temp"}'
```

* Fill disk

Fill disk by fallocate:

```bash
curl -X POST "127.0.0.1:31767/api/attack/disk" -H "Content-Type: application/json" -d '{"action":"fill", "size":1024, "path":"temp", "fill_by_fallocate": true}'
```

Fill disk by write data to file:

```bash
curl -X POST "127.0.0.1:31767/api/attack/disk" -H "Content-Type: application/json" -d '{"action":"fill", "size":1024, "path":"temp", "fill_by_fallocate": false}'
```

#### Recover attack

Recover a attack:

```bash
$ curl -X DELETE "127.0.0.1:31767/api/attack/20df86e9-96e7-47db-88ce-dd31bc70c4f0"
```