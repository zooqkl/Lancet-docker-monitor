package monitor

import (
	"Lancet/logging"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"strconv"
	"strings"
	"time"
)

type ContainerInfo struct {
	HostName      string
	ContainerID   string
	ContainerName string
	//ContainerStats ContainerStatsSpec
}

type ContainerStatsSpec struct {
	HostName      string
	ContainerName string
	Cpu           string
	Memory        string
	NetIN         string
	NetOUT        string
	BlockRead     string
	BlockWrite    string
	ReadTime      string
}

var logger = lancetlogging.GetLogger()

/*
docker daemon会记录这次读取/sys/fs/cgroup/cpuacct/docker/[containerId]/cpuacct.usage的值，
作为cpu_total_usage；并记录了上一次读取的该值为pre_cpu_total_usage；
读取/proc/stat中cpu field value，并进行累加，得到system_usage;并记录上一次的值为pre_system_usage；
读取/sys/fs/cgroup/cpuacct/docker/[containerId]/cpuacct.usage_percpu中的记录，组成数组per_cpu_usage_array；

docker stats计算Cpu Percent的算法：

cpu_delta = cpu_total_usage - pre_cpu_total_usage;
system_delta = system_usage - pre_system_usage;
CPU % = ((cpu_delta / system_delta) * length(per_cpu_usage_array) ) * 100.0

*/
func (cs *ContainerStatsSpec) calculateCPUPercentUnix(containStats []byte) (float64, error) {

	j, err := simplejson.NewJson(containStats)
	if err != nil {
		return 0.0, err
	}
	cpu_total_usage, _ := j.Get("cpu_stats").Get("cpu_usage").Get("total_usage").Uint64()
	pre_cpu_total_usage, _ := j.Get("precpu_stats").Get("cpu_usage").Get("total_usage").Uint64()
	system_usage, _ := j.Get("cpu_stats").Get("system_cpu_usage").Uint64()
	pre_system_usage, _ := j.Get("precpu_stats").Get("system_cpu_usage").Uint64()
	percpuUsage := j.Get("precpu_stats").Get("cpu_usage").Get("percpu_usage").MustArray()
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(cpu_total_usage) - float64(pre_cpu_total_usage)
		// calculate the change for the entire system between readings
		systemDelta = float64(system_usage) - float64(pre_system_usage)
	)
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(percpuUsage)) * 100.0
	}
	return cpuPercent, nil
}

/*
读取/sys/fs/cgroup/memory/docker/[containerId]/memory.usage_in_bytes的值，作为mem_usage；
如果容器限制了内存，则读取/sys/fs/cgroup/memory/docker/[id]/memory.limit_in_bytes作为mem_limit，否则mem_limit = machine_mem；

MEM USAGE = mem_usage
MEM LIMIT = mem_limit
MEM % = (mem_usage / mem_limit) * 100.0
*/
func (cs *ContainerStatsSpec) calculateMemory(containStats []byte) (float64, error) {
	var memPercent = 0.0
	j, err := simplejson.NewJson(containStats)
	if err != nil {
		return 0.0, err
	}

	//memUsage, _ := j.Get("memory_stats").Get("usage").Float64()
	//memLimit, _ := j.Get("memory_stats").Get("limit").Float64()
	//memCache, _ := j.Get("memory_stats").Get("stats").Get("cache").Float64()
	//
	//if memUsage > 0.0 && memLimit > 0.0 && memCache >= 0.0 {
	//	memPercent = ((float64(memUsage) - float64(memCache)) / float64(memLimit)) * 100.0
	//}

	memUsage, _ := j.Get("memory_stats").Get("usage").Float64()
	memLimit, _ := j.Get("memory_stats").Get("limit").Float64()

	if memUsage > 0.0 && memLimit > 0.0 {
		memPercent = (float64(memUsage) / float64(memLimit)) * 100.0
	}

	return memPercent, nil
}

/*
获取属于该容器network namespace veth pairs在主机中对应的veth*虚拟网卡EthInterface数组，
然后循环数组中每个网卡设备，读取/sys/class/net//statistics/rx_bytes得到rx_bytes,
读取/sys/class/net//statistics/tx_bytes得到对应的tx_bytes。
将所有这些虚拟网卡对应的rx_bytes累加得到该容器的rx_bytes。
将所有这些虚拟网卡对应的tx_bytes累加得到该容器的tx_bytes。
NET I = rx_bytes
NET O = tx_bytes
*/
func (cs *ContainerStatsSpec) calculateNetIO(containStats []byte) ([]float64, error) {
	var netIn = 0.0
	var netOut = 0.0
	j, err := simplejson.NewJson(containStats)
	if err != nil {
		return nil, err
	}
	netInBytes, _ := j.Get("networks").Get("eth0").Get("rx_bytes").Float64()
	netOutBytes, _ := j.Get("networks").Get("eth0").Get("tx_bytes").Float64()

	if netInBytes > 0.0 && netOutBytes > 0.0 {
		netIn = float64(netInBytes) / 1024 / 1024
		netOut = float64(netOutBytes) / 1024 / 1024
	}
	return []float64{netIn, netOut}, nil
}

/*
获取每个块设备的IoServiceBytesRecursive数据：先去读取/sys/fs/cgroup/blkio/docker/[containerId]/blkio.io_serviced_recursive中是否有有效值，
如果有，则读取/sys/fs/cgroup/blkio/docker/[containerId]/blkio.io_service_bytes_recursive的值返回；
如果没有，就去读取/sys/fs/cgroup/blkio/docker/[containerId]/blkio.throttle.io_service_bytes中的值返回；
将每个块设备的IoServiceBytesRecursive数据中所有read field对应value进行累加，得到该容器的blk_read值；
将每个块设备的IoServiceBytesRecursive数据中所有write field对应value进行累加，得到该容器的blk_write值；
*/
func (cs *ContainerStatsSpec) calculateBlockIO(containStats []byte) ([]float64, error) {
	type BlkJoin struct {
		Major uint64 `json:"major"`
		Minor uint64 `json:"minor"`
		Op    string `json:"op"`
		Value uint64 `json:"value"`
	}
	var blk_read = 0.0
	var blk_write = 0.0
	var blockBytes []byte
	blk_statss := []*BlkJoin{}

	j, err := simplejson.NewJson(containStats)
	if err != nil {
		return nil, err
	}
	blkArray, _ := j.Get("blkio_stats").Get("io_serviced_recursive").Array()

	var blk_service []interface{}
	if len(blkArray) == 0 {
		blk_service, _ = j.Get("blkio_stats").Get("io_service_bytes").Array()
	} else {
		blk_service, _ = j.Get("blkio_stats").Get("io_service_bytes_recursive").Array()
	}
	blockBytes, _ = json.Marshal(blk_service)
	err = json.Unmarshal(blockBytes, &blk_statss)
	if err != nil {
		return nil, fmt.Errorf("[calculateBlockIO] json Unmarshal blk_statss Error: %s", err)
	}
	var blk_writeBytes, blk_readBytes uint64
	for _, blk := range blk_statss {
		if strings.EqualFold(blk.Op, "Write") {
			blk_writeBytes += blk.Value
		} else if strings.EqualFold(blk.Op, "Read") {
			blk_readBytes += blk.Value
		}
	}
	blk_write = float64(blk_writeBytes) / 1024 / 1024
	blk_read = float64(blk_readBytes) / 1024 / 1024
	return []float64{blk_read, blk_write}, nil
}

/*
   将获取容器状态的时间戳记录下来，并转化为Unix time，单位为秒
*/
func (cs *ContainerStatsSpec) calculateReadTime(containStats []byte) (string, error) {
	var readTime string
	j, err := simplejson.NewJson(containStats)
	if err != nil {
		logger.Errorf("Use simplejson NewJon Error: %s", err)
		return "", err
	}
	readTime, _ = j.Get("read").String()

	t, err := time.Parse(time.RFC3339Nano, readTime)
	if err != nil {
		logger.Errorf("Time parse Error :%s", err)
		return "", fmt.Errorf("Time parse Error :%s", err)
	}
	unix_time := t.Unix()
	readTime = strconv.Itoa(int(unix_time))
	return readTime, nil
}
