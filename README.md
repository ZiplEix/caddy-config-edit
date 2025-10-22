# 🧩 Caddy Config CLI

A simple Go-based command-line tool to manage Caddy reverse-proxy configurations stored as multiple `.caddy` label files.
It helps automate the creation, modification, and reload of Caddy configurations running inside a Docker container.

## 🚀 Overview

Instead of editing your main Caddyfile manually, this CLI lets you manage each domain or group of domains as separate files inside /etc/caddy/sites/.
Each file (called a label) contains one or more reverse-proxy entries.

Example structure:

```
/etc/caddy/Caddyfile
/etc/caddy/sites/
├── coolify.caddy
├── your-site.caddy
└── media.caddy
```

Your main `Caddyfile` just imports them:

```Caddyfile
import /etc/caddy/sites/*.caddy
```

## ⚙️ Installation

Download the file:

```bash
sudo wget -O /usr/local/bin/caddy-config-edit https://github.com/ZiplEix/caddy-config-edit/raw/refs/heads/master/bin/caddy-config-edit
sudo chmod +x /usr/local/bin/caddy-config-edit
```

or with curl:

```bash
sudo curl -L \
  -o /usr/local/bin/caddy-config-edit \
  https://github.com/ZiplEix/caddy-config-edit/raw/refs/heads/master/bin/caddy-config-edit
sudo chmod +x /usr/local/bin/caddy-config-edit
```

with go (requires Go installed):

```bash
git clone https://github.com/yourname/caddy-config-cli.git
cd caddy-config-cli
go build -o caddy-config-edit
```

or run directly with go:

```bash
go run .
```

## 🧰 Commands

1. `label`

    Create an empty label file sued to group multiple entries.

    ```bash
    caddy-config-edit label <name> [flags]
    ```

    Flags:

    | Flag       | Shorthand | Description | Default |
    |------------|-----------|-------------|---------|
    | --dir      | -d        | Directory where the label file will be created | `/srv/proxy/sites` |
    | --ext      |           | File extension | `.caddy` |
    | --force    | -f        | Overwrite the file if it already exists | `false` |

    Example:

    ```bash
    caddy-config-edit label coolify
    # → creates /srv/proxy/sites/coolify.caddy
    ```

2. `newEntry`

    Append or replace a reverse-proxy entry in a given label file.

    ```bash
    caddy-config-edit newEntry <label> <host> <ip[:port]> [flags]
    ```

    Flags:

    | Flag       | Shorthand | Description | Default |
    |------------|-----------|-------------|---------|
    | --dir      | -d        | Directory where the label file resides | `/srv/proxy/sites` |
    | --ext      |           | File extension for the label | `.caddy` |
    | --force    | -f        | Replace the entry if it already exists | `false` |

    Behavior:
    - If the label file doesn’t exist, it’s created automatically.
    - If the host already exists, the command fails unless `--force` is provided.
    - Warns if the same IP is already used by another host in the same label.

    Example:

    ```bash
    # Add a new reverse proxy block
    caddy-config-edit newEntry coolify admin.example.com 10.10.0.20:3001

    # Replace existing entry
    caddy-config-edit newEntry coolify admin.example.com 10.10.0.20:4000 --force
    ```

    Resulting file `/srv/proxy/sites/coolify.caddy`:

    ```Caddyfile
    admin.example.com {
        import common
        reverse_proxy 10.10.0.20:4000
    }
    ```

3. `reload`

    Format, validate, and reload the Caddy configuration inside Docker.

    ```bash
    caddy-config-edit reload [flags]
    ```

    Flags:

    | Flag       | Shorthand | Description | Default |
    |------------|-----------|-------------|---------|
    | --container | -c       | Docker container name running Caddy | `caddy` |
    | --config   | -f        | Path to Caddyfile inside the container | `/etc/caddy/Caddyfile` |
    | --no-tty   |           | Disable TTY mode when executing docker commands | `false` |
    | --quiet    | -q        | Reduce output verbosity (only errors) | `false` |

    What is runs:

    ```bash
    docker exec -it caddy caddy fmt --overwrite /etc/caddy/Caddyfile
    docker exec -it caddy caddy validate --config /etc/caddy/Caddyfile
    docker exec -it caddy caddy reload   --config /etc/caddy/Caddyfile
    ```

    Example:

    ```bash
    # Reload the config after adding new entries
    caddy-config-edit reload
    ```

## 🐳 Docker setup

```yml
services:
  caddy:
    image: caddy:2
    container_name: caddy
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - ./sites:/etc/caddy/sites:rw
      - caddy_data:/data
      - caddy_config:/config

volumes:
  caddy_data:
  caddy_config:
```

Your main `Caddyfile` should import all labels:

```
import /etc/caddy/sites/*.caddy
```

## 🧩 Example workflow

```bash
# 1. Create a label file
caddy-config-edit label coolify

# 2. Add domains
caddy-config-edit newEntry coolify admin.example.com 10.10.0.20:3001
caddy-config-edit newEntry coolify api.example.com   10.10.0.20:3002

# 3. Reload Caddy
caddy-config-edit reload
```

## 🧪 Development

Each command is implemented as a Cobra subcommand inside `cmd/`:

```
cmd/
├── label.go       # Create label files
├── newEntry.go    # Add/replace reverse proxy entries
├── reload.go      # Run fmt/validate/reload via Docker
├── root.go        # Root command definition
└── utils.go       # Helper functions (run, isSafeFilename)
```

## 🧠 Design notes

- Uses **Go + Cobra** for clean, modular CLI design.
- Keeps all per-site configs isolated (`/etc/caddy/sites/*.caddy`).
- Integrates directly with Docker to safely reload Caddy without manual edits.
- Fails gracefully with meaningful errors and warnings.

## 📜 License

MIT — free to use and modify.
Built for self-hosted environments with ❤️.

## Changelog

- 0.1.0 - Initial release
