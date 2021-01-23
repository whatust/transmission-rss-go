transmission-rss
================

[![Build status][build-img]][build]
[![Code Coverage][coverage-img]][coverage] 

Installation
------------

### Yay
TODO

### Manual
```sh
git clone  https://github.com/whatust/transmission-rss-go
cd transmission-rss-go
make install
```

Configuration
-------------

### Minimal configuration
```yaml
server:
    host: localhost

login:
    username: transmission
    password: transmission

seenFile: /etc/transmission-rss-conf.yml
rssFile: /etc/transmission-rss-feeds.yml
```

### Complete configuration
```yaml
server:
    host: localhost
    port: 9091
    tls: false
    rpsPath: /tranmission/rpc
    validateCert: true
    saveTorrent: false
    torrentPath: /var/lib/torrents
    proxy: ""

connection:
    retries: 10
    timeout: 10
    waitTime: 3

login:
    username: transmission
    password: transmission

log:
    logPath: /var/log/transmission-rss.log
    level: Info
    maxSize: 10000
    maxBackups: 1
    maxAge: 10
    compress: false
    localTime: true
    formatter: "JSON"

seenFile: /etc/transmission-rss-see.log
rssFile: /etc/transmission-rss-feeds.log
```

### Feed list
```yaml
feeds:
    - url:
        matchers:
            - regex:
                downloadPath:
        validateCert:
```

Daemonized Startup
------------------

### Systemd service

Example service file found in `contrib/transmission-rss.service`.
```ini
[Unit]
Description=Transmission RSS daemon.
After=transmission-daemon.service
Wants=network-online.target

[Service]
Type=forking
ExecStart=/usr/bin/transmission-rss -f
ExecReload=/bin/kill -s HUP $MAINPID

[Install]
WantedBy=multi-user.target
```

Copy it to `/etc/systemd/system/` to add to system service.
Make sure `ExecStart` has the full path to the binary.

### Conjob

Add the line bellow to the crontab file to run transmission-rss every 15 min.
`*/15 * * * * /usr/bin/transmission-rss`

[build-img]: https://www.travis-ci.com/whatust/transmission-rss-go.svg?branch=main
[build]: https://www.travis-ci.com/whatust/transmission-rss-go
[coverage-img]: https://codecov.io/gh/whatust/transmission-rss-go/branch/main/graph/badge.svg?token=QIUAJ9KA3A)
[coverage]: https://codecov.io/gh/whatust/transmission-rss-go