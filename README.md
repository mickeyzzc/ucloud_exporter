# 云IDC供应商`Ucloud`的监控api的exporter采集工具

## 功能：

支持采集`Ucloud`自身监控数据的以下监控指标，并转化为`Prometheus`格式。

- 云主机`uhost`
- 云分布式内存数据库`umem`
- 主备版Redis`uredis`
- 云EIP`ueip`
- 云磁盘`udisk`
- 云数据库`udb`，包括`nosql`,`sql`,`postgresql`

----
```
  -log.level string
        'info','debug' (default "info")
  -log.path string
        log path (default "./ucloud_exporter.log")
  -ucloud.interval int
        Ucloud time interval.  (default 3600)
  -ucloud.privatekey string
        Ucloud api  PrivateKey.
  -ucloud.publickey string
        Ucloud api  PublicKey.
  -web.listen-port int
        An port to listen on for web interface and telemetry. (default 58086)
  -web.telemetry-path string
        A path under which to expose metrics. (default "/metrics")
```


绑定Eip的合并流量
```
sum(ucloud_eip_info{bind_resource_type="uhost"} * on (resource_id) group_left eip_NetworkIn) by (bind_resource_name,bind_resource_id)
```