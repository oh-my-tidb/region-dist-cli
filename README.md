region-dist-cli
--

Visualize region distribution of a [TiDB](https://github.com/pingcap/tidb) Cluster.

![Screenshot](screenshot.png)

Screen record:
https://asciinema.org/a/hwmxyMkpoQioTqEtmCHAtiUPM

Usage:

```bash
> go get github.com/disksing/region-dist-cli
> region-dist-cli -pd="127.0.0.1:2379"
```

Watch:
```bash
> watch -c region-dist-cli
```
