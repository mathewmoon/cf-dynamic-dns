# cf-dynamic-dns
Dynamic DNS update using Cloudflare APIv4
Config should be placed in /etc/cf_dynamic_updater/updater.conf
cfupdater is an init script for systemv

TODO:
  Accept cmdline args for log and config locations
  Set a better name for the default config path
  Init script should check process table against PID file
