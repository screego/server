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

```ini
# The external ip of the server.
# Execute the following command on the server you want to host Screego
# to find your external ip.
#   curl 'https://api.ipify.org'
SCREEGO_EXTERNAL_IP=

# A secret which should be unique. Is used for cookie authentication.
SCREEGO_SECRET=

# If TLS should be enabled for HTTP requests. Screego requires TLS,
# you either have to enable this setting or serve TLS via a reverse proxy.
SCREEGO_SERVER_TLS=false
# The TLS cert file (only needed if TLS is enabled)
SCREEGO_TLS_CERT_FILE=
# The TLS key file (only needed if TLS is enabled)
SCREEGO_TLS_KEY_FILE=

# The address the http server will listen on.
SCREEGO_SERVER_ADDRESS=0.0.0.0:5050

# The address the TURN server will listen on.
SCREEGO_TURN_ADDRESS=0.0.0.0:3478

# If reverse proxy headers should be trusted.
# Screego uses ip whitelisting for authentication
# of TURN connections. When behind a proxy the ip is always the proxy server.
# To still allow whitelisting this setting must be enabled and
# the `X-Real-Ip` header must be set by the reverse proxy.
SCREEGO_TRUST_PROXY_HEADERS=false

# Defines when a user login is required
# Possible values:
#   all: User login is always required
#   turn: User login is required for TURN connections
#   none: User login is never required
SCREEGO_AUTH_MODE=turn

# Defines origins that will be allowed to access Screego (HTTP + WebSocket)
# Example Value: https://screego.net,https://sub.gotify.net
SCREEGO_CORS_ALLOWED_ORIGINS=

# Defines the location of the users file.
# File Format:
#   user1:bcrypt_password_hash
#   user2:bcrypt_password_hash
#
# Example:
#   user1:$2a$12$WEfYCnWGk0PDzbATLTNiTuoZ7e/43v6DM/h7arOnPU6qEtFG.kZQy
#
# The user password pair can be created via
#   screego hash --name "user1" --pass "your password"
SCREEGO_USERS_FILE=

# The loglevel (one of: debug, info, warn, error)
SCREEGO_LOG_LEVEL=info

# If screego should expose a prometheus endpoint at /metrics. The endpoint
# requires basic authentication from a user in the users file.
SCREEGO_PROMETHEUS=false
```
