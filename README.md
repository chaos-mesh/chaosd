# chaosd

chaosd is an easy-to-use Chaos Engineering tool used to inject failures to a physical node. Currently, two modes are supported:

- **Command mode** - Using chaosd as a command-line tool. Supported failure types are:
  
    - [Process attack](#process-attack)
    - [Network attack](#network-attack)
    - [Stress attack](#stress-attack)
    - [Disk attack](#disk-attack)
    - [Host attack](#host-attack)
    - [Recover attack](#recover-attack)

- **Server mode** - Running chaosd as a daemon server. Supported failure types are:
    - [Process attack](#process-attack-1)
    - [Network attack](#network-attack-1)
    - [Stress attack](#stress-attack-1)
    - [Disk attack](#disk-attack-1)
    - [Recover attack](#recover-attack-1)

## Prerequisites

Before deploying chaosd, make sure the following items have been installed:

* [tc](https://linux.die.net/man/8/tc)
* [ipset](https://linux.die.net/man/8/ipset)
* [iptables](https://linux.die.net/man/8/iptables)
* [stress-ng](https://wiki.ubuntu.com/Kernel/Reference/stress-ng)

## Install

You can either build directly from the source or download the binary to finish the installation.

- Build from source code

    ```bash
    make chaosd
    chmod +x chaosd && mv chaosd /usr/local/bin/chaosd
    ```

- Download binary

    ```bash
    curl -fsSL -o chaosd https://mirrors.chaos-mesh.org/latest/chaosd
    chmod +x chaosd && mv chaosd /usr/local/bin/chaosd
    ```

## Usages

### Command mode

#### Process attack

Attacks a process according to the PID or process name. Supported tasks are:

- **kill process** 
  
    Description: Kills a process by sending the `SIGKILL` signal
  
    Sample usage:

    ```bash
    $ chaosd attack process kill -p [pid] # set pid or pod name
    # the generated uid is used to recover chaos attack
    Attack network successfully, uid: 2c865e6f-299f-4adf-ab37-94dc4fb8fea6
    ```

- **stop process**

    Description: Kills a process by sending the `SIGKILL` signal

    Sample usage:

    ```bash
    $ chaosd attack process stop -p [pid] # set pid or pod name
    ```

#### Network attack

Attacks the network using `iptables`, `ipset`, and `tc`. Supported tasks are:

- **delay network packet**

    Description: Sends messages with the specified latency

    Sample usage:

    ```bash
    $ chaosd attack network delay -d eth0 -i 172.16.4.4 -l 10ms
    ```

- **lose network packet**

    Description: Drops network packets randomly

    Sample usage:

    ```bash
    $ chaosd attack network loss -d eth0 -i 172.16.4.4 --percent 50%
    ```

- **corrupt network packet**

    Description: Causes packet corruption

    Sample usage:

    ```bash
    $ chaosd attack network corrupt -d eth0 -i 172.16.4.4 --percent 50%
    ```

- **duplicate network packet**

    Description: Sends duplicated packets

    Sample usage:

    ```bash
    $ chaosd attack network duplicate -d eth0 -i 172.16.4.4 --percent 50%
    ```

#### Stress attack

Generates stress on the host. Supported tasks are:

- **CPU stress**

   Description: Generates stress on the host CPU

   Sample usage:

    ```bash
    $ chaosd attack stress cpu -l 100 -w 2
    ```

- **Memory stress**

   Description: Generates stress on the host memory

   Sample usage:

    ```bash
    $ chaosd attack stress mem -w 2 # stress 2 CPU and each cpu loads 100%
    ```

#### Disk attack

Attacks the disk by increasing write/read payload, or filling up the disk. Supported tasks are:

- **add payload**

    Description: Add read/write payload

    Sample usage:

    ```bash
    ./bin/chaosd attack disk add-payload read --path /tmp/temp --size 100
    ```

    ```bash
    ./bin/chaosd attack disk add-payload write --path /tmp/temp --size 100
    ```

- **fill disk**

    Description: Fills up the disk

    Sample usage:


    ```bash
    ./bin/chaosd attack disk fill --fallocate true --path /tmp/temp --size 100   //filling using fallocate
    ```

    ```bash
    ./bin/chaosd attack disk fill --fallocate false --path /tmp/temp --size 100  //filling by writing data to files
    ```

#### Host attack

Shuts down the host

Sample usage:

```bash
./bin/chaosd attack host shutdown
```

> **Note:**
>
> This command will shut down the host. Be cautious when you execute it.

#### Recover attack

Recovers an attack

Sample usage:

```bash
$ chaosd recover 2c865e6f-299f-4adf-ab37-94dc4fb8fea6
```

### Server Mode

To enter server mode, execute the following:

```bash
nohup ./bin/chaosd server > chaosd.log 2>&1 &
```

And then you can inject failures by sending HTTP requests.

> **Note**:
>
> Make sure you are operating with the privileges to run iptables, ipset, etc. Or you can run chaosd with `sudo`.

#### Process attack

Attacks a process according to the PID or process name. Supported tasks are:

- **kill process**

    Description: Kills a process by sending the `SIGKILL` signal
  
    Sample usage:

    ```bash
    curl -X POST "127.0.0.1:31767/api/attack/process" -H "Content-Type: application/json"  -d '{"process": "{pid}", "signal": 9}' # set pid or pod name
    {"status":200,"message":"attack successfully","uid":"e6d01a30-4528-4c70-b4fb-4dc47c4d39be"}
    ```

- **stop process**

    Description: Kills a process by sending the `SIGKILL` signal

    Sample usage:

    ```bash
    curl -X POST "127.0.0.1:31767/api/attack/process" -H "Content-Type: application/json"  -d '{"process": "{pid}", "signal": 15}' # set pid or pod name
    {"status":200,"message":"attack successfully","uid":"ecf3f564-c4c0-4aaf-83c6-4b511a6e3a85"}
    ```

#### Network attack

Attacks the network using `iptables`, `ipset`, and `tc`. Supported tasks are:

- **delay network packet**

    Description: Sends messages with the specified latency

    Sample usage:

    ```bash
    $ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "eth0", "ipaddress": "172.16.4.4", "action": "delay", "latency": "10ms", "jitter": "10ms", "correlation": "0"}'
    ```

- **lose network packet**

    Description: Drops network packets randomly

    Sample usage:

    ```bash
    $ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "eth0", "ipaddress": "172.16.4.4", "action": "loss", "percent": "50", "correlation": "0"}'
    ```

- **corrupt network packet**

    Description: Causes packet corruption

    Sample usage:

    ```bash
    $ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "eth0", "ipaddress": "172.16.4.4", "action": "corrupt", "percent": "50",  "correlation": "0"}'
    ```

- **duplicate network packet**

    Description: Sends duplicated packets

    Sample usage:

    ```bash
    $ curl -X POST "127.0.0.1:31767/api/attack/network" -H "Content-Type: application/json"  -d '{"device": "eth0", "ipaddress": "172.16.4.4", "action": "duplicate", "percent": "50", "correlation": "0"}'
    ```

#### Stress attack

Generates stress on the host. Supported tasks are:

- **CPU stress**

   Description: Generates stress on the host CPU

   Sample usage:

    ```bash
    $ curl -X POST 127.0.0.1:31767/api/attack/stress -H "Content-Type:application/json" -d '{"action":"cpu", "load": 100, "worker": 2}'
    ```

- **Memory stress**

   Description: Generates stress on the host memory

   Sample usage:

    ```bash
    $ curl -X POST 127.0.0.1:31767/api/attack/stress -H "Content-Type:application/json" -d '{"action":"mem", "worker": 2}'
    ```

#### Disk attack

Attacks the disk by increasing write/read payload, or filling up the disk. Supported tasks are:

- Add payload

    Description: Add read/write payload

    Sample usage:

    ```bash
    curl -X POST "127.0.0.1:31767/api/attack/disk" -H "Content-Type: application/json" -d '{"action":"read-payload","size":1024,"path":"temp"}'

    ```bash
    curl -X POST "127.0.0.1:31767/api/attack/disk" -H "Content-Type: application/json" -d '{"action":"write-payload","size":1024,"path":"temp"}'
    ```

- Fill disk

    Description: Fills up the disk

    Sample usage:

    ```bash
    curl -X POST "127.0.0.1:31767/api/attack/disk" -H "Content-Type: application/json" -d '{"action":"fill", "size":1024, "path":"temp", "fill_by_fallocate": true}' //filling using fallocate
    ```

    ```bash
    curl -X POST "127.0.0.1:31767/api/attack/disk" -H "Content-Type: application/json" -d '{"action":"fill", "size":1024, "path":"temp", "fill_by_fallocate": false}' //filling by writing data to files
    ```

#### Recover attack

Recovers an attack

Sample usage:

```bash
$ curl -X DELETE "127.0.0.1:31767/api/attack/20df86e9-96e7-47db-88ce-dd31bc70c4f0"
```
