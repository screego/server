# Installation

Latest Version: **GITHUB_VERSION**

?> Before starting Screego you may read [Configuration](config.md).

!> TLS is required for Screego to work. Either enable TLS inside Screego or 
   use a reverse proxy to serve Screego via TLS.

## Docker

Setting up Screego with docker is pretty easy, you basically just have to start the docker container, and you are ready to go:

[ghcr.io/screego/server](https://github.com/orgs/screego/packages/container/package/server) and
[screego/server](https://hub.docker.com/r/screego/server)
docker images are multi-arch docker images.
This means the image will work for `amd64`, `i386`, `ppc64le` (power pc), `arm64`, `armv7` (Raspberry PI) and `armv6`.

When using [TURN](nat-traversal.md), Screego will allocate ports for relay
connections, this currently only works with network mode host inside docker.
See [#56](https://github.com/screego/server/issues/56)

By default, Screego runs on port 5050.

?> Replace `EXTERNALIP` with your external IP. One way to find your external ip is with ipify.

   ```bash
   $ curl 'https://api.ipify.org'
   ```

### Network Host

```bash
$ docker run --net=host -e SCREEGO_EXTERNAL_IP=EXTERNALIP ghcr.io/screego/server:GITHUB_VERSION
```

#### docker-compose.yml

```yaml
version: "3.7"
services:
  screego:
    image: ghcr.io/screego/server:GITHUB_VERSION
    network_mode: host
    environment:
      SCREEGO_EXTERNAL_IP: "EXTERNALIP"
```

## Binary

### Supported Platforms:

- linux_amd64 (64bit)
- linux_i386 (32bit)
- armv7 (32bit used for Raspberry Pi)
- armv6
- arm64 (ARMv8)
- ppc64
- ppc64le
- windows_i386.exe (32bit)
- windows_amd64.exe (64bit)

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

!> Maintenance of the AUR Packages is not performed by the Screego team.
   You should always check the PKGBUILD before installing an AUR package.

Screego's latest release is available in the AUR as [screego-server](https://aur.archlinux.org/packages/screego-server/) and [screego-server-bin](https://aur.archlinux.org/packages/screego-server-bin/).
The development-version can be installed with [screego-server-git](https://aur.archlinux.org/packages/screego-server-git/).

## FreeBSD

!> Maintenance of the FreeBSD Package is not performed by the Screego team.
   Check yourself, if you can trust it.

```bash
$ pkg install screego
```

## Source

[See Development#build](development.md#build)
