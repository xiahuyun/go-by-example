# 运行 prometheus

```aiignore
prometheus --config.file=prometheus.yml
```

# 运行 alertmanager

```aiignore
alertmanager --config.file=alertmanager.yml
```

# 运行 grafana

```aiignore
/opt/homebrew/opt/grafana/bin/grafana server --config /opt/homebrew/etc/grafana/grafana.ini --homepath /opt/homebrew/opt/grafana/share/grafana --packaging\=brew cfg:default.paths.logs\=/opt/homebrew/var/log/grafana cfg:default.paths.data\=/opt/homebrew/var/lib/grafana cfg:default.paths.plugins\=/opt/homebrew/var/lib/grafana/plugins
```

