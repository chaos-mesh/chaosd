# chaosd

Chaosd is an easy-to-use Chaos Engineering tool. This tool is used to inject failures to the physic node, such as kill process, network failure, CPU burn, memory burn, etc.

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

You can use Chaosd to inject failures with [Command mode](#comamnd-mode) or [Server mode](#server-mode).

### Command mode

Using Chaosd as a command-line tool.

#### Process attack

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

#### Network attack

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

#### Stress attack

* CPU stress

```bash
$ chaosd attack stress cpu -l 100 -w 2
```

* Memory stress

```bash
$ chaosd attack stress mem -w 2 # stress 2 CPU and each cpu loads 100%
```

#### Recover attack

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

* kill process

```bash
curl -X POST "127.0.0.1:31767/api/attack/process" -H "Content-Type: application/json"  -d '{"process": "{pid}", "signal": 9}' # set pid or pod name
{"status":200,"message":"attack successfully","uid":"e6d01a30-4528-4c70-b4fb-4dc47c4d39be"}
```

* stop process

```bash
curl -X POST "127.0.0.1:31767/api/attack/process" -H "Content-Type: application/json"  -d '{"process": "{pid}", "signal": 15}' # set pid or pod name
{"status":200,"message":"attack successfully","uid":"ecf3f564-c4c0-4aaf-83c6-4b511a6e3a85"}
```

#### Network attack

* delay network packet

```bash
$ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "ens33", "ipaddress": "172.16.4.4", "action": "delay", "latency": "10ms", "jitter": "10ms", "correlation": "0"}'
```

* loss network packet

```bash
$ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "ens33", "ipaddress": "172.16.4.4", "action": "loss", "percent": "50", "correlation": "0"}'
```

* corrupt network packet

```bash
$ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "ens33", "ipaddress": "172.16.4.4", "action": "corrupt", "percent": "50",  "correlation": "0"}'
```

* duplicate network packet

```bash
$ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "ens33", "ipaddress": "172.16.4.4", "action": "duplicate", "percent": "50", "correlation": "0"}'
```

#### Stress attack

* CPU stress

```bash
$ curl -X POST 127.0.0.1:31767/api/attack/stress -H "Content-Type:application/json" -d '{"action":"cpu", "load": 100, "worker": 2}'
```

* Memory stress

```bash
$ curl -X POST 127.0.0.1:31767/api/attack/stress -H "Content-Type:application/json" -d '{"action":"mem", "worker": 2}'
```

#### Recover attack

```bash
$ curl -X DELETE "127.0.0.1:31767/api/attack/20df86e9-96e7-47db-88ce-dd31bc70c4f0"
```