# 云IDC供应商`Ucloud`的监控api的exporter采集工具

## 功能：

支持采集`Ucloud`的以下监控指标，并转化为`Prometheus`格式。

- 云主机`uhost`
- 云分布式Redis`umem`
- 云EIP`ueip`


绑定Eip的合并流量
```
sum(ucloud_eip_info{bind_resource_type="uhost"} * on (resource_id) group_left eip_NetworkIn) by (bind_resource_name,bind_resource_id)
```