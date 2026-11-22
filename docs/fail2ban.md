# Fail2ban

[Fail2ban](https://github.com/fail2ban/fail2ban) can protect the Screego login (`POST /login`)
from brute-force attacks by banning IPs that repeatedly fail to authenticate.

!> Failed logins are only logged on the `debug` log level.
   Set `SCREEGO_LOG_LEVEL=debug` before configuring Fail2ban.
   See [Configuration](config.md).

A failed login produces an access log line like this:

```text
2026-11-23T04:33:35+05:30 DBG HTTP duration=1.234ms host=screego.example.com ip=203.0.113.5:51234 path=/login size=39 status=401
```

The field names and values may be wrapped in ANSI color escapes. The filter below handles that.

!> When Screego runs behind a reverse proxy, `ip` is the address of the proxy, not the client.
   In that case configure Fail2ban on the access log of the proxy instead
   (see [Reverse Proxy](proxy.md) and [nginx](#nginx) below).

## Filter

Create `/etc/fail2ban/filter.d/screego.conf`:

```ini
[Definition]

# Optional ANSI escape sequence emitted by Screego's console logger around
# field names/values. \\u001b additionally matches the escaped form stored by
# docker's json-file log driver.
_ansi = (?:(?:\x1b|\\u001b)\[[0-9;]*m)?

# A failed login is an HTTP request to /login answered with status 401:
#   DBG HTTP duration=1.234ms host=screego.example.com ip=203.0.113.5:51234 path=/login size=39 status=401
# The port after the remote IP is matched but ignored; IPv6 addresses may
# appear bracketed ([::1]:51234).
failregex = HTTP .*%(_ansi)sip=%(_ansi)s\[?<ADDR>\]?:\d+\s+%(_ansi)spath=%(_ansi)s/login\s+.*%(_ansi)sstatus=%(_ansi)s401\b

ignoreregex =
```

## Jail

Create `/etc/fail2ban/jail.d/screego.conf` (or add the jail to `jail.local`).
The `logpath`/`journalmatch` depends on how Screego's stdout reaches fail2ban.

### Screego running as a systemd service (binary)

```ini
[screego]
enabled = true
filter = screego
backend = systemd
journalmatch = _SYSTEMD_UNIT=screego.service
maxretry = 5
findtime = 10m
bantime = 1h
# Port Screego is served on, used by the ban action.
port = 5050
```

### Screego running in docker

When the docker container uses the `journald` log driver
(`--log-driver=journald` or `log_driver: journald` in docker-compose):

```ini
[screego]
enabled = true
filter = screego
backend = systemd
journalmatch = CONTAINER_NAME=screego
maxretry = 5
findtime = 10m
bantime = 1h
port = 5050
```

With docker's default `json-file` log driver point Fail2ban at the container log files:

```ini
[screego]
enabled = true
filter = screego
logpath = /var/lib/docker/containers/*/*-json.log
maxretry = 5
findtime = 10m
bantime = 1h
port = 5050
```

?> When Screego uses `network_mode: host` (the recommended setup) the default ban
   action works. When ports are published via docker (`-p`), traffic passes the
   `DOCKER-USER` chain instead of `INPUT`; add `chain = DOCKER-USER` to the jail.

If Screego's stdout is redirected to a file (e.g. `docker logs -f screego >> /var/log/screego.log`
or a shell redirect for the binary), use `logpath` with `backend = auto`:

```ini
[screego]
enabled = true
filter = screego
logpath = /var/log/screego.log
maxretry = 5
findtime = 10m
bantime = 1h
port = 5050
```

## Test

Test the filter against a log file or the journal:

```bash
$ fail2ban-regex /var/log/screego.log /etc/fail2ban/filter.d/screego.conf
$ fail2ban-regex systemd-journal /etc/fail2ban/filter.d/screego.conf
```

Enable the jail and check its status:

```bash
$ fail2ban-client reload
$ fail2ban-client status screego
```

## nginx

When Screego is served by nginx, the client IP is available in nginx' access
log. A request to `/login` answered with `401` is a failed login attempt.

Create `/etc/fail2ban/filter.d/nginx-screego.conf`:

```ini
[Definition]
# nginx access log: 203.0.113.5 - - [23/Nov/2026:04:33:35 +0530] "POST /login HTTP/1.1" 401 39 "-" "curl/8"
failregex = ^<ADDR> - - \[.*\] "POST /login [^"]*" 401

ignoreregex =
```

Jail in `/etc/fail2ban/jail.d/screego.conf`:

```ini
[screego]
enabled = true
filter = nginx-screego
logpath = /var/log/nginx/access.log
maxretry = 5
findtime = 10m
bantime = 1h
port = http,https
```
