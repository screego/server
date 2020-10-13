# Config

!> TLS is required for Screego to work. Either enable TLS inside Screego or 
   use a reverse proxy to serve Screego via TLS.

Screego tries to obtain config values from different locations in sequence. 
Properties will never be overridden. Thus, the first occurrence of a setting will be used.

#### Order

* Environment Variables
* `screego.config.local` (in same path as the binary)
* `screego.config` (in same path as the binary)
* `$HOME/.config/screego/server.config`
* `/etc/screego/server.config`

#### Config Example

[screego.config.example](https://raw.githubusercontent.com/screego/server/master/screego.config.example ':include :type=code ini')
