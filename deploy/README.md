# gscale-zebra Linux release

This package is built for Linux and is tested for Ubuntu/Arch style hosts.

## Contents

- `bin/scale` - scale + zebra workflow worker
- `bin/bot` - telegram + ERP worker
- `bin/zebra` - zebra diagnostic utility
- `config/*.env.example` - config templates
- `systemd/*.service` - service templates
- `install.sh` - install helper

## Quick install

```bash
tar -xzf gscale-zebra-<version>-linux-<arch>.tar.gz
cd gscale-zebra-<version>-linux-<arch>
sudo ./install.sh
```

Then set real credentials:

- `config/bot.env` (token + ERP creds)
- `config/scale.env` (device paths)

Start services:

```bash
sudo systemctl restart gscale-scale.service gscale-bot.service
sudo systemctl status gscale-scale.service gscale-bot.service
```
