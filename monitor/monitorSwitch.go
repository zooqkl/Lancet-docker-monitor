package monitor

import (
	"time"
	"strings"
	"io/ioutil"
	"fmt"
)

type MonitorSwitch struct {
	MonitorCliList []*MonitorCli
}

var monitorSwitch *MonitorSwitch

func NewMonitorSwitch(mc []*MonitorCli) *MonitorSwitch {
	if monitorSwitch == nil {
		monitorSwitch = &MonitorSwitch{MonitorCliList: mc}
		return monitorSwitch
	}
	return monitorSwitch
}
func (ms *MonitorSwitch) StartMonitor() {
	mcs := ms.MonitorCliList
	for _, mc := range mcs {
		go startOneMonitor(mc)
	}
}

func (ms *MonitorSwitch) StopMonitor() {
	//mcs := ms.MonitorCliList
	FinishMonitor <- true
	close(FinishMonitor)
	cl, err := getRecordDataList()
	if err != nil {
		logger.Errorf("getRecordDataList Error: %s", err)
		return
	}
	var intervalTime float64
	if len(ms.MonitorCliList) > 0 {
		intervalTime = ms.MonitorCliList[0].intervalTime.Seconds()
	}
	HandleData(cl, intervalTime)
	logger.Debugf("Make the chart completed! please watch in 'Lancet/resultData/ChartFile' Contents !")
	//for _, mc := range mcs {
	//	FinishChart.Add(1)
	//	cl, _ := mc.GetRecordDataList()
	//	go HandleData(cl)
	//}
	//FinishChart.Wait()
}

/*
每次向一个服务器的所有容器获取一次容器状态，每次间隔intervalTime
*/
func startOneMonitor(monCli *MonitorCli) {
	cl, _ := monCli.GetContainList()
	i := 0
	for {
		//每隔1分钟获取一次容器List，防止中途有容器挂掉，还在监控
		if i*(int)(monCli.intervalTime/time.Second)%60 == 0 {
			cl, _ = monCli.GetContainList()
		}
		for _, c := range cl {
			logger.Debugf("start MonitorContain[%s]!", c.ContainerName)
			go monCli.MonitorContain(monCli.Hostname, c.ContainerName)
		}
		select {
		case <-FinishMonitor:
			logger.Debugf("Finish Monitor Work !")
			return
		default:
			logger.Debugf("Monitor Work RuningTime is %d !", i*(int)(monCli.intervalTime))
		}
		i++
		time.Sleep(monCli.intervalTime)
	}
}

func getRecordDataList() ([]*ContainerInfo, error) {
	var containerList = []*ContainerInfo{}
	logger.Debugf("Start get ExcelFileList ！ ")
	excelFilelist, err := ioutil.ReadDir("./resultData/ExcelFile/")
	if err != nil {
		err := fmt.Errorf("get excelFilelist fail! Error: %s", err)
		return nil, err
	}
	for _, excelFile := range excelFilelist {
		fileName := strings.Split(excelFile.Name(), ".xls")[0]
		hostName := strings.Split(fileName, "-")[0]
		continueName := strings.Replace(fileName, hostName+"-", "", 1)
		conInfo := &ContainerInfo{hostName, "", continueName}
		containerList = append(containerList, conInfo)
	}
	return containerList, nil
}
