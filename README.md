# chaosd

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/chaos-mesh/chaosd)

chaosd is an easy-to-use Chaos Engineering tool used to inject failures to a physical node. 

## Types of fault

You can use Chaosd to simulate the following fault types:

- Process: Injects faults into the processes. Operations such as killing the process or stopping the process are supported.
- Network: Injects faults into the network of physical machines. Operations such as increasing network latency, losing packets, and corrupting packets are supported.
- Stress: Injects stress on the CPU or memory of the physical machines.
- Disk: Injects faults into disks of the physical machines. Operations such as increasing disk load of reads and writes, and filling disks are supported.
- Host: Injects faults into the physical machine. Operations such as shutdown the physical machine are supported.

For details about the introduction and usage of each fault type, refer to the related documentation.

## Work modes

You can use Chaosd in the following modes:

- Command-line mode: Run Chaosd directly as a command-line tool to inject and recover faults.

- Service mode: Run Chaosd as a service in the background, to inject and recover faults by sending HTTP requests.

## Prerequisites

Before deploying chaosd, make sure the following items have been installed:

* [tc](https://linux.die.net/man/8/tc)
* [ipset](https://linux.die.net/man/8/ipset)
* [iptables](https://linux.die.net/man/8/iptables)
* [stress-ng](https://wiki.ubuntu.com/Kernel/Reference/stress-ng) (required when install chaosd by building from source code)
* [byteman](https://github.com/chaos-mesh/byteman)(required when install chaosd by building from source code)

## Install

You can either build directly from the source or download the binary to finish the installation.

- Build from source code


    Build chaosd:

    ```bash
    make chaosd
    mv chaosd /usr/local/bin/chaosd
    ```

    Build or download tools related to Chaosd:

    ```bash
    make chaos-tools
    ```

    Put Chaosd into `PATH`:

    ```
    mv ./bin /usr/local/chaosd
    export PATH=$PATH:/usr/local/chaosd
    ```

- Download binary

    Download the latest unstable version by executing the command below:

    ```bash
    curl -fsSL -o chaosd-latest-linux-amd64.tar.gz https://mirrors.chaos-mesh.org/chaosd-latest-linux-amd64.tar.gz
    ```

    If you want to download the release version, you can replace the `latest` in the above command with the version number. For example, download `v1.1.1` by executing the command below:

    ```bash
    curl -fsSL -o chaosd-v1.1.1-linux-amd64.tar.gz https://mirrors.chaos-mesh.org/chaosd-v1.1.1-linux-amd64.tar.gz
    ```

    Then uncompress the archived file, and you can go into the folder and execute chaosd：

    ```bash
    tar zxvf chaosd-latest-linux-amd64.tar.gz && cd chaosd-latest-linux-amd64
    ```

## Document

For details about the introduction and usage of chaosd, refer to the [documentation](https://chaos-mesh.org/docs/chaosd-overview/).
