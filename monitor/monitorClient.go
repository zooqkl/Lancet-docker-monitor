package monitor

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"github.com/docker/go-connections/tlsconfig"
)

type MonitorCli struct {
	Hostname       string
	host           string
	apiVersion     string
	intervalTime   time.Duration
	dockerClient   *client.Client
	tlsSwitch      bool
	clientCertPath []string
}

var FinishMonitor chan bool

func NewMonitorCliFromConf(hostname string, host string, apiVersion string, intervalTime time.Duration, tlsSwitch bool, tlsCertPath []string) (*MonitorCli, error) {
	moncli := &MonitorCli{hostname, host, apiVersion, intervalTime, nil, tlsSwitch, tlsCertPath}
	dkcli, err := moncli.newClient(tlsSwitch, tlsCertPath)
	if err != nil {
		return nil, fmt.Errorf("New monitorCli Error:%s", err)
	}
	moncli.dockerClient = dkcli
	return moncli, nil
}

func (mc *MonitorCli) newClient(tlsSwitch bool, tlsCertPath []string) (*client.Client, error) {

	if mc.dockerClient != nil {
		return mc.dockerClient, nil
	}
	var httpClient *http.Client

	if tlsSwitch {
		if len(tlsCertPath) != 3 {
			return nil, fmt.Errorf("ClientCertPath's length error，expect 3 !")
		}
		options := tlsconfig.Options{
			CAFile:             tlsCertPath[0],
			CertFile:           tlsCertPath[1],
			KeyFile:            tlsCertPath[2],
			InsecureSkipVerify: tlsSwitch,
		}
		tlsc, err := tlsconfig.Client(options)
		if err != nil {
			return nil, err
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsc,
			},
		}
	}

	cli, err := client.NewClient(mc.host, mc.apiVersion, httpClient, nil)
	if err != nil {
		return nil, fmt.Errorf("NewClient connection error: %s", err)
	}
	return cli, nil
}

func (mc *MonitorCli) GetContainList() ([]*ContainerInfo, error) {
	var containerList = []*ContainerInfo{}
	if mc.dockerClient == nil {
		logger.Errorf("Pleace init Client ! dockerClient is nil ! MonitorCli :%v", mc)
		return nil, fmt.Errorf("Pleace init Client ! dockerClient is nil ! MonitorCli :%v", mc)
	}
	logger.Debugf("Start get ContainerList ！ ")
	containers, err := mc.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		conInfo := &ContainerInfo{mc.Hostname, container.ID[:10], strings.Split(container.Names[0], "/")[1]}
		containerList = append(containerList, conInfo)
	}
	return containerList, nil
}

func (mc *MonitorCli) GetContainStats(hostname string, containName string) (*ContainerStatsSpec, error) {
	conStats := new(ContainerStatsSpec)
	resp, err := mc.dockerClient.ContainerStats(context.Background(), containName, false)
	if err != nil {
		logger.Errorf("GetContainStats fail error : %s", err)
		return nil, fmt.Errorf("GetContainStats fail error : %s", err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	cpu, err := conStats.calculateCPUPercentUnix(content)
	if err != nil {
		logger.Errorf("[calculateCPUPercentUnix] Error:%s", err)
		return nil, fmt.Errorf("[calculateCPUPercentUnix] Error:%s", err)
	}
	memory, err := conStats.calculateMemory(content)
	if err != nil {
		logger.Errorf("[calculateMemory] Error:%s", err)
		return nil, fmt.Errorf("[calculateMemory] Error:%s", err)
	}
	netIO, err := conStats.calculateNetIO(content)
	if err != nil {
		return nil, fmt.Errorf("[calculateNetIO] Error:%s", err)
	}
	blockIO, err := conStats.calculateBlockIO(content)
	if err != nil {
		return nil, fmt.Errorf("[calculateBlockIO] Error:%s", err)
	}
	readTime, err := conStats.calculateReadTime(content)
	if err != nil {
		return nil, fmt.Errorf("[calculateReadTime] Error:%s", err)
	}
	conStats.ContainerName = containName
	conStats.Cpu = fmt.Sprintf("%6.2f", cpu)
	conStats.Memory = fmt.Sprintf("%6.2f", memory)
	conStats.NetIN = fmt.Sprintf("%6.2f", netIO[0])
	conStats.NetOUT = fmt.Sprintf("%6.2f", netIO[1])
	conStats.BlockRead = fmt.Sprintf("%6.2f", blockIO[0])
	conStats.BlockWrite = fmt.Sprintf("%6.2f", blockIO[1])
	conStats.ReadTime = readTime
	conStats.HostName = hostname
	return conStats, nil
}

func (mc *MonitorCli) MonitorContain(hostname string, cname string) {
	cstats, err := mc.GetContainStats(hostname, cname)
	if err != nil || nil == cstats {
		logger.Errorf("GetContainStats Error: %s", err)
		return
	}
	handler := NewHandlerStatsFile(hostname, cname)
	result, err := handler.WriteStatsFile(*cstats)
	if err != nil {
		logger.Errorf("WriteStatsFile Error: %s", err)
	}
	logger.Debugf("Container [%s] WriteStatsFile 's result is   %t", hostname+"-"+cname, result)
}
