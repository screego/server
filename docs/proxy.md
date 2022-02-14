# Proxy

!> When using a proxy enable `SCREEGO_TRUST_PROXY_HEADERS`. See [Configuration](config.md).

## nginx

### At root path

```nginx
upstream screego {
  # Set this to the address configured in
  # SCREEGO_SERVER_ADDRESS. Default 5050
  server 127.0.0.1:5050;
}

server {
  listen 80;

  # Here goes your domain / subdomain
  server_name screego.example.com;

  location / {
    # Proxy to screego
    proxy_pass         http://screego;
    proxy_http_version 1.1;

    # Set headers for proxying WebSocket
    proxy_set_header   Upgrade $http_upgrade;
    proxy_set_header   Connection "upgrade";
    proxy_redirect     http:// $scheme://;

    # Set proxy headers
    proxy_set_header   X-Real-IP $remote_addr;
    proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header   X-Forwarded-Proto http;

    # The proxy must preserve the host because screego verifies it with the origin
    # for WebSocket connections
    proxy_set_header   Host $http_host;
  }
}
```

### At a sub path

```nginx
upstream screego {
  # Set this to the address configured in
  # SCREEGO_SERVER_ADDRESS. Default 5050
  server 127.0.0.1:5050;
}

server {
  listen 80;

  # Here goes your domain / subdomain
  server_name screego.example.com;

  location /screego/ {
    rewrite ^/screego(/.*) $1 break;
  
    # Proxy to screego
    proxy_pass         http://screego;
    proxy_http_version 1.1;

    # Set headers for proxying WebSocket
    proxy_set_header   Upgrade $http_upgrade;
    proxy_set_header   Connection "upgrade";
    proxy_redirect     http:// $scheme://;

    # Set proxy headers
    proxy_set_header   X-Real-IP $remote_addr;
    proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header   X-Forwarded-Proto http;

    # The proxy must preserve the host because screego verifies it with the origin
    # for WebSocket connections
    proxy_set_header   Host $http_host;
  }
}
```

## Apache (httpd)

The following modules are required:

* mod_proxy
* mod_proxy_wstunnel
* mod_proxy_http

### At root path

```apache
<VirtualHost *:80>
    ServerName screego.example.com
    Keepalive On

    # The proxy must preserve the host because screego verifies it with the origin
    # for WebSocket connections
    ProxyPreserveHost On

    # Replace 5050 with the port defined in SCREEGO_SERVER_ADDRESS.
    # Default 5050

    # Proxy web socket requests to /stream
    ProxyPass "/stream" ws://127.0.0.1:5050/stream retry=0 timeout=5

    # Proxy all other requests to /
    ProxyPass "/" http://127.0.0.1:5050/ retry=0 timeout=5

    ProxyPassReverse / http://127.0.0.1:5050/
</VirtualHost>
```

### At a sub path

```apache
<VirtualHost *:80>
    ServerName screego.example.com
    Keepalive On

    Redirect 301 "/screego" "/screego/"

    # The proxy must preserve the host because screego verifies it with the origin
    # for WebSocket connections
    ProxyPreserveHost On

    # Proxy web socket requests to /stream
    ProxyPass "/screego/stream" ws://127.0.0.1:5050/stream retry=0 timeout=5

    # Proxy all other requests to /
    ProxyPass "/screego/" http://127.0.0.1:5050/ retry=0 timeout=5
    #                 ^- !!trailing slash is required!!

    ProxyPassReverse /screego/ http://127.0.0.1:5050/
</VirtualHost>
```
