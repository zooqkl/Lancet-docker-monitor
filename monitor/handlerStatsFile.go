package monitor

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

type HandlerStatsFile struct {
	ContainerName string
	FirstTime     int
	sync.Mutex
}

//var handler *HandlerStatsFile
type Handlers struct {
	HandlersMap map[string]*HandlerStatsFile
	Lock        sync.RWMutex
}

var handlers *Handlers

func NewHandlerStatsFile(hostname string, containName string) *HandlerStatsFile {
	if handlers == nil {
		handlers = &Handlers{
			HandlersMap: make(map[string]*HandlerStatsFile),
		}
	}
	if handler, ok := handlers.getHandler(hostname + containName); ok {
		return handler
	}
	handler := &HandlerStatsFile{ContainerName: containName}
	handlers.setHandler(hostname+containName, handler)
	return handler
}

func (h *Handlers) getHandler(key string) (*HandlerStatsFile, bool) {
	h.Lock.RLock()
	defer h.Lock.RUnlock()
	if handler, ok := h.HandlersMap[key]; ok {
		return handler, ok
	}
	return nil, false
}

func (h *Handlers) setHandler(k string, v *HandlerStatsFile) {
	h.Lock.Lock()
	defer h.Lock.Unlock()
	h.HandlersMap[k] = v
}

func (h *HandlerStatsFile) WriteStatsFile(cs ContainerStatsSpec) (bool, error) {
	h.Lock()
	defer h.Unlock()
	filePath := "./resultData/ExcelFile/" + cs.HostName + "-" + cs.ContainerName + ".xls"
	outputFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return false, fmt.Errorf("Open file [%s] Error: \n", filePath, err)
	}
	outputFile.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	tpsExclWrite := csv.NewWriter(outputFile)
	t, _ := strconv.Atoi(cs.ReadTime)
	if h.FirstTime == 0 {
		h.FirstTime = t
		tpsExclWrite.Write([]string{"Time", "CPU / %", "Memory / %", "NetINTotal / MB", "NetOUTTotal / MB", "BlockRead / MB", "BlockWrite / MB"})
	}
	timesTamps := strconv.Itoa((t - h.FirstTime + 1))
	//当获取不到容器数据时，计算出的timeTamps < 0
	timesTampsInt, err := strconv.Atoi(timesTamps)
	if timesTampsInt < 0 || err != nil {
		return false, fmt.Errorf("Current Container [%s] stats isn't running!,Don't write mointorData! Error :%s", cs.HostName+"-"+cs.ContainerName, err)
	}
	tpsExclWrite.Write([]string{timesTamps, cs.Cpu, cs.Memory, cs.NetIN, cs.NetOUT, cs.BlockRead, cs.BlockWrite})
	tpsExclWrite.Flush()
	return true, nil
}
func (h *HandlerStatsFile) ReadStatsFile(hostname string, containName string) ([]ContainerStatsSpec, error) {
	filePath := "./resultData/ExcelFile/" + hostname + "-" + containName + ".xls"
	fi, err := os.Open(filePath)
	if err != nil {
		logger.Errorf("ReadFile [%s] Error: %s", filePath, err)
		return nil, fmt.Errorf("ReadFile [%s] Error: %s", filePath, err)
	}
	br := bufio.NewReader(fi)
	var conStats []ContainerStatsSpec
	for {
		statsDataBytes, _, c := br.ReadLine()
		if c == io.EOF {
			logger.Debugf("ReadFile %s.xls finish!", containName)
			break
		}
		statsData := string(statsDataBytes)
		//logger.Debugf("data is %s", statsData)
		if strings.Contains(statsData, "Memory") || statsData == "" {
			continue
		}
		//data'format is 9,"  0.04","  0.08","  1.58","  1.08","  0.00","  0.03"
		statsData = strings.Replace(statsData, "\"", "", 12)
		statsData = strings.Replace(statsData, " ", "", 10)
		statsData = strings.Replace(statsData, "\xEF\xBB\xBF", "", -1)
		datas := strings.Split(statsData, ",")
		if len(datas) == 7 {
			conStats = append(conStats, ContainerStatsSpec{hostname, containName, datas[1], datas[2], datas[3], datas[4], datas[5], datas[6], datas[0]})
		}
	}
	return conStats, nil
}
