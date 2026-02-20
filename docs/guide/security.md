# Security

Plik allows users to upload and serve any content as-is. Hosting untrusted content raises security concerns that Plik addresses with several mechanisms.

## Content-Type Override

For security reasons, Plik doesn't trust user-provided MIME types and relies solely on server-side detection. This means some files may not render properly in browsers or embedded viewers that require the correct MIME type.

::: warning Office format detection limitation
Office formats like `.pptx`, `.docx`, and `.xlsx` are ZIP archives internally, so Go's built-in MIME detector (`http.DetectContentType`) identifies them as `application/zip` instead of their proper types (e.g., `application/vnd.openxmlformats-officedocument.presentationml.presentation`).
:::

## Enhanced Web Security

When `EnhancedWebSecurity` is enabled in `plikd.cfg`, Plik sets additional HTTP headers:

- **X-Content-Type-Options**: `nosniff`
- **X-XSS-Protection**: enabled
- **X-Frame-Options**: deny
- **Content-Security-Policy**: restrictive policy disabling resource loading, XHR, iframes
- **Secure Cookies**: session cookies only transmitted over HTTPS

::: warning
Enhanced security will break audio/video playback, PDF rendering, and other rich content features. Disable it if you need those capabilities.
:::

::: danger Authentication requires HTTPS with EnhancedWebSecurity
When `EnhancedWebSecurity` is enabled, session cookies have the `Secure` flag set and can only be transmitted over HTTPS connections. Authentication will not work over plain HTTP.
:::

## Download Domain

It is recommend to serve uploaded files on a separate (sub-)domain to:

- Protect against phishing links using your main domain
- Protect Plik's session cookie from being exposed to uploaded content

Configure with `DownloadDomain` in `plikd.cfg`:

```toml
DownloadDomain = "https://dl.plik.example.com"
```

### Troubleshooting: Redirect Loops

If you see this error:

```
Invalid download domain 127.0.0.1:8080, expected plik.root.gg
```

`DownloadDomain` checks the `Host` header of incoming HTTP requests. By default, reverse proxies like Nginx and Apache do not forward this header. Make sure to configure:

```
Apache mod_proxy: ProxyPreserveHost On
Nginx:            proxy_set_header Host $host;
```

## XSRF Protection

Plik uses a dual-cookie XSRF protection mechanism:

1. The `plik-xsrf` cookie value must be copied into the `X-XSRFToken` HTTP header for all mutating authenticated requests
2. This prevents cross-site request forgery attacks

## Upload Restrictions

### Source IP Whitelist

Restrict uploads and user creation to specific IP ranges:

```toml
UploadWhitelist = ["10.0.0.0/8", "192.168.1.0/24"]
```

### Authentication

Set `FeatureAuthentication = "forced"` to require authentication for all uploads.

### Upload Tokens

Authenticated users can generate upload tokens to link CLI uploads to their account. Tokens are sent via the `X-PlikToken` HTTP header.

## Link Preview Bot Protection

Messaging apps like Slack, Telegram, WhatsApp, Signal, Discord, and others generate link previews by fetching shared URLs. This is problematic for:

- **One-shot uploads**: the bot's preview request counts as the single allowed download, deleting the file before the intended recipient can access it
- **Streaming uploads**: the bot consumes the stream data, leaving nothing for the real downloader

Plik automatically blocks known messaging app link preview bots from downloading one-shot and streaming files, returning a **406 Not Acceptable** response. Normal (multi-download) uploads are not affected — bots can still generate previews for those.

::: tip No configuration needed
This protection is always active and requires no configuration. It uses a hardcoded list of known bot user-agent strings that is maintained with Plik releases.
:::

### Blocked Bots

Slack, Telegram, WhatsApp, Signal, Facebook/Messenger, Discord, Skype, Viber, LinkedIn, Twitter/X, Microsoft Teams, Wire, Mattermost, Rocket.Chat, and Zulip.

