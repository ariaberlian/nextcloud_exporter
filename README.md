# Nextcloud Prometheus Exporter

Fetches Nextcloud server info and exposes metrics for Prometheus.

## Build

```bash
# For Linux
GOOS=linux GOARCH=amd64 go build -o nextcloud-exporter .
```

## Configuration

| Flag | Env Variable | Description | Default |
|------|--------------|-------------|---------|
| `-url` | `NEXTCLOUD_URL` | Nextcloud base URL | (required) |
| `-token` | `NC_TOKEN` | NC-Token header value | (required) |
| `-listen` | `LISTEN_ADDR` | Listen address | `:9205` |

## Usage

```bash
# Using flags
./nextcloud-exporter \
  -url "https://your-nextcloud.com" \
  -token "your-token"

# Using environment variables
export NEXTCLOUD_URL="https://your-nextcloud.com"
export NC_TOKEN="your-token"
./nextcloud-exporter
```

## Metrics

Available at `http://localhost:9205/metrics`

- `nextcloud_status_info` - Status info (version, productname, edition)
- `nextcloud_status_installed` - Installation status (0/1)
- `nextcloud_status_maintenance` - Maintenance mode (0/1)
- `nextcloud_status_needs_db_upgrade` - DB upgrade needed (0/1)
- `nextcloud_status_extended_support` - Extended support (0/1)
- `nextcloud_system_info` - Version info
- `nextcloud_system_freespace_bytes` - Free disk space
- `nextcloud_system_cpuload` - CPU load (1m, 5m, 15m)
- `nextcloud_system_mem_total_bytes` / `_free_bytes` - Memory
- `nextcloud_system_swap_total_bytes` / `_free_bytes` - Swap
- `nextcloud_apps_installed_total` - Installed apps count
- `nextcloud_apps_updates_available_total` - Available updates
- `nextcloud_update_available` - Nextcloud update available (0/1)
- `nextcloud_users_total` - Total users
- `nextcloud_files_total` - Total files
- `nextcloud_shares_*` - Share statistics
- `nextcloud_php_*` - PHP settings and opcache stats
- `nextcloud_database_size_bytes` - Database size
- `nextcloud_active_users{period}` - Active users by period
- `nextcloud_scrape_success` - Scrape status (0/1)
