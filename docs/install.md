# Installation

Latest Version: **GITHUB_VERSION**

?> Before starting Screego you may read [Configuration](config.md).

!> TLS is required for Screego to work. Either enable TLS inside Screego or 
   use a reverse proxy to serve Screego via TLS.

## Docker

Setting up Screego with docker is pretty easy, you basically just have to start the docker container, and you are ready to go:

The [screego/server](https://hub.docker.com/r/screego/server) docker images are multi-arch docker images. 
This means the image will work for `amd64`, `i386`, `ppc64le` (power pc), `arm64`, `arm v7` (Raspberry PI).

When using [TURN](nat-traversal.md), Screego will allocate ports for relay connections.

The ports must be mapped that the host system forwards them to the screego container.

By default, Screego runs on port 5050.

?> Replace `YOUREXTERNALIP` with your external IP. One way to find your external ip is with ipify.
   ```bash
   $ curl 'https://api.ipify.org'
   ```

### Network Host (recommended)

```bash
$ docker run --net=host -e SCREEGO_EXTERNAL_IP=YOUREXTERNALIP screego/server:GITHUB_VERSION
```

#### docker-compose.yml

```yaml
version: "3.7"
services:
  screego:
    image: screego/server:GITHUB_VERSION
    network_mode: host
    environment:
      SCREEGO_EXTERNAL_IP: "YOUREXTERNALIP"
```

### Port Range

`SCREEGO_TURN_STRICT_AUTH` should only be disabled if you enable TLS inside
Screego and not use a reverse proxy with `SCREEGO_TRUST_PROXY_HEADERS=true`.


```bash
$ docker run \
    -e SCREEGO_TURN_PORT_RANGE=50000:50100 \
    -e SCREEGO_TURN_STRICT_AUTH=false \
    -e SCREEGO_EXTERNAL_IP=YOUREXTERNALIP \
    -p 5050:5050 \
    -p 3478:3478 \
    -p 50000-50100:50000-50100/udp \
    screego/server:GITHUB_VERSION
```

#### docker-compose.yml

```yaml
version: "3.7"
services:
  screego:
    image: screego/server:GITHUB_VERSION
    ports:
      - 5050:5050
      - 3478:3478
      - 50000-50100:50000-50100/udp
    environment:
      SCREEGO_TURN_PORT_RANGE: "50000:50100"
      SCREEGO_EXTERNAL_IP: "YOUREXTERNALIP"
      SCREEGO_TURN_STRICT_AUTH: "false"
```

## Binary

### Supported Platforms:

* linux_amd64 (64bit)
* linux_i386 (32bit)
* armv7 (32bit used for Raspberry Pi)
* armv6
* arm64 (ARMv8)
* ppc64
* ppc64le
* windows_i386.exe (32bit)
* windows_amd64.exe (64bit)

Download the zip with the binary for your platform from [screego/server Releases](https://github.com/screego/server/releases).

```bash
$ wget https://github.com/screego/server/releases/download/vGITHUB_VERSION/screego_GITHUB_VERSION_{PLATFORM}.tar.gz
```

Unzip the archive.

```bash
$ tar xvf screego_GITHUB_VERSION_{PLATFORM}.tar.gz
```

Make the binary executable (linux only).

```bash
$ chmod +x screego
```

Execute screego:

```bash
$ ./screego
# on windows
$ screego.exe
```

## Arch-Linux(aur)

TODO

## Source

TODO
