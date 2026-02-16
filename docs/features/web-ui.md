# Web UI

Plik serves a web interface on the same port as the API by default.

## Configuration

| Parameter | Default | Description |
|-----------|---------|-------------|
| `NoWebInterface` | `false` | Disable the web UI entirely |
| `WebappDirectory` | `../webapp/dist` | Path to static files |

## Customization

The web interface can be customized:

| File | Purpose |
|------|---------|
| `js/custom.js` | Change the page title |
| `css/custom.css` | Override styles (use `!important`) |
| `img/background.jpg` | Custom background image |
| `favicon.ico` | Custom favicon |

### Docker Customization

When running in Docker, files are at `/home/plik/webapp/dist`:

```bash
docker run -p 8080:8080 \
  -v my_background.jpg:/home/plik/webapp/dist/img/background.jpg \
  rootgg/plik
```

## Features

### Inline File Viewer

The web interface includes an inline file viewer for text files (code, logs, markdown, etc.). 
- **Auto-display**: If an upload contains only one text file, the viewer is displayed by default.
- **Syntax Highlighting**: Automatic detection of hundreds of languages.
- **JSON Formatting**: Pretty-print and validation buttons for JSON files.
