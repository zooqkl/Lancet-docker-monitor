# Lancet

## 说明
本工具是为了监控docker容器的状态。状态包含 cpu、内存、网络实时速度统计
（NETInSpeed、NETOutSpeed）、硬盘写/读 累加总量。
并将统计的数据制成图表。

Lancet.resultData.ChartFile: 生成的图表文件放在此处。

Lancet.resultData.ExcelFile: 统计的监控数据文件放在此处。


## 使用方法

### 修改docker.service 文件
##### ① 查看配置文件docker.service位于哪里

systemctl show --property=FragmentPath docker

返回结果：
FragmentPath=/lib/systemd/system/docker.service

##### ② 编辑配置文件内容，接收所有ip请求

sudo vim  /lib/systemd/system/docker.service

ExecStart=/usr/bin/dockerd -H unix:///var/run/docker.sock -H tcp://0.0.0.0:2375
##### ③ 重新加载配置文件，重启docker daemon，并查看dockerAPI版本
sudo systemctl daemon-reload

sudo systemctl restart docker

docker version 

获取Client.APIversion字段内容
###  Config文件更改

 修改Lancet.config.config.yaml

```
 monitorHosts:
   host10:
     address: tcp://192.168.9.10:2375
     apiVersion: 1.24
   host54:
     address: tcp://192.168.9.54:2375
     apiVersion: 1.24
 intervalTime: 1s
 monitorTime: 30s
 monitorSwitch: true
``` 
 
 monitorHosts 下面可以配置多个host，配置address和apiversion即可。

 hostname是作为区分host用的，上面实例用host10、host54来作为hostname。

 intervalTime 监控间隔时间，即每隔多久去拉取一次容器状态。

 monitorTime： 支持 s/秒、m/分、h/时， 监控总时长，时间到了，监控将停止，将数据写入excel文件，制出的折线图。

 monitorSwitch ：用作在中途紧急停止监控，启动时必须为 true。 例如:monitorTime 配置 2h，但监控了1h后，数据已经足够，
 monitorSwitch设置为false，将紧急停止监控，并将已获取的数据制成图表。