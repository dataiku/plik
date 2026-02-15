# Reverse Proxy

Plik can be deployed behind a reverse proxy (Nginx, Apache, Caddy, Traefik, etc.).

## Path Prefix

If Plik is served under a sub-path, set the `Path` config parameter:

```toml
Path = "/plik"
```

## Source IP

Behind a proxy, the source IP will be the proxy's address. Configure the source IP header:

```toml
SourceIpHeader = "X-Forwarded-For"
```

## Nginx Example

```nginx
server {
    listen 443 ssl;
    server_name plik.example.com;

    client_max_body_size 10G;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_request_buffering off;
        proxy_buffering off;
    }
}
```

::: warning
Set `client_max_body_size` to match your `MaxFileSizeStr` setting. For streaming uploads, disable request buffering.
:::

## Apache Example

```apache
<VirtualHost *:443>
    ServerName plik.example.com

    ProxyPreserveHost On
    ProxyPass / http://127.0.0.1:8080/
    ProxyPassReverse / http://127.0.0.1:8080/

    RequestHeader set X-Forwarded-For %{REMOTE_ADDR}s
    RequestHeader set X-Forwarded-Proto https
</VirtualHost>
```

## Disabling Nginx Buffering

By default, Nginx buffers large HTTP requests and responses to a temporary file. This causes unnecessary disk I/O and slower transfers. Disable buffering (requires Nginx ≥1.7.12) for `/file` and `/stream` paths, and increase buffer sizes:

```nginx
proxy_buffering off;
proxy_request_buffering off;
proxy_http_version 1.1;
proxy_buffer_size 1M;
proxy_buffers 8 1M;
client_body_buffer_size 1M;
```

See the [Nginx proxy buffering documentation](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering) for details.

