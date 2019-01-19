package monitor

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"strconv"
	"strings"
	"os"
)

type ContainstatsPlot struct {
	HostName      string
	ContainerName string
	IntervalTime  float64
	Plot          *plot.Plot
}

//var FinishChart sync.WaitGroup

type ChartInfo struct {
	Title      string
	XLabel     string
	YLabel     string
	ChartDatas [][]XYData
}
type XYData struct {
	X float64
	Y float64
}

var plotHandlers map[string]*ContainstatsPlot

func NewHandlerPlot(hostName string, containName string, intervalTime float64) *ContainstatsPlot {
	if plotHandlers == nil {
		plotHandlers = make(map[string]*ContainstatsPlot)
	}
	if handler, ok := plotHandlers[hostName+containName]; ok {
		return handler
	}
	p, _ := plot.New()
	plotHandlers[hostName+containName] = &ContainstatsPlot{HostName: hostName, ContainerName: containName, IntervalTime: intervalTime, Plot: p}
	return plotHandlers[hostName+containName]
}
func (cp *ContainstatsPlot) MakeChart(ci ChartInfo) {
	p := cp.Plot

	p.Title.Text = ci.Title
	p.X.Label.Text = ci.XLabel
	p.Y.Label.Text = ci.YLabel

	//p.BackgroundColor = color.RGBA{39, 40, 34, 255}
	points := make([]plotter.XYs, 0)

	for _, chartData := range ci.ChartDatas {
		var xydatas plotter.XYs
		for _, xyData := range chartData {
			xydatas = append(xydatas, XYData{xyData.X, xyData.Y})
		}
		points = append(points, xydatas)
	}

	if len(points) == 2 {
		plotutil.AddLinePoints(p, points[0], points[1])
	} else {
		for _, point := range points {
			plotutil.AddLinePoints(p, point)
		}
	}
	//NET Speed \n green-netIn  red-netOut ==> NET Speed

	fileName := cp.HostName + "-" + cp.ContainerName + "-" + strings.Split(ci.Title, "\n")[0] + ".png"
	if err := CreatePath("./resultData/ChartFile/"); err != nil {
		logger.Fatal(err)
	}
	filePath := "./resultData/ChartFile/" + fileName

	err := p.Save(4*vg.Inch, 4*vg.Inch, filePath)
	if err != nil {
		logger.Fatalf("MakeChart %s  Error: %s", fileName, err)
	}
	// clear old data
	p.Clear()
	logger.Debugf("MakeChart %s  Success!", fileName)
}

func (cp *ContainstatsPlot) FormatChartData(cs []ContainerStatsSpec) []ChartInfo {
	//chartInfos := make([]ChartInfo)
	var chartInfos []ChartInfo
	chartInfos = append(chartInfos, ChartInfo{Title: "CPU", XLabel: "Time", YLabel: "CPU   %"})
	chartInfos = append(chartInfos, ChartInfo{Title: "Memory", XLabel: "Time", YLabel: "MEMORY   %"})
	chartInfos = append(chartInfos, ChartInfo{Title: "NET Speed \ngreen-netIn  red-netOut ", XLabel: "Time", YLabel: "NetSpeed  mb/s"})
	chartInfos = append(chartInfos, ChartInfo{Title: "BLOCKIO \ngreen-blockWrite  red-blockRead   ", XLabel: "Time", YLabel: "BLOCK  MB"})

	//var  cpu_charts, mem_charts, net_charts, blk_charts [][] XYData
	var cpu_chart, mem_chart, netin_chart, netout_chart, blkRead_chart, blkWrite_chart []XYData
	for i, conStats := range cs {
		readTime, err := strconv.ParseFloat(conStats.ReadTime, 64)
		if err != nil {
			logger.Errorf("ParseFloat Error: %s", err)
		}
		cpu, _ := strconv.ParseFloat(conStats.Cpu, 64)
		mem, _ := strconv.ParseFloat(conStats.Memory, 64)
		netIn, _ := strconv.ParseFloat(conStats.NetIN, 64)
		netOut, _ := strconv.ParseFloat(conStats.NetOUT, 64)
		blockRead, _ := strconv.ParseFloat(conStats.BlockRead, 64)
		blockWrite, _ := strconv.ParseFloat(conStats.BlockWrite, 64)

		cpu_chart = append(cpu_chart, XYData{readTime, cpu})
		mem_chart = append(mem_chart, XYData{readTime, mem})
		blkRead_chart = append(blkRead_chart, XYData{readTime, blockRead})
		blkWrite_chart = append(blkWrite_chart, XYData{readTime, blockWrite})
		if i == 0 {
			netin_chart = append(netin_chart, XYData{readTime, netIn})
			netout_chart = append(netout_chart, XYData{readTime, netOut})
			continue
		}

		intervalTime := cp.IntervalTime
		previousNetIN, _ := strconv.ParseFloat(cs[i-1].NetIN, 64)
		previousNetout, _ := strconv.ParseFloat(cs[i-1].NetOUT, 64)
		netINSpeed := (netIn - previousNetIN) / intervalTime
		netOUTSpeed := (netOut - previousNetout) / intervalTime
		if netINSpeed < 0 {
			netINSpeed = netIn / intervalTime
		}
		if netOUTSpeed < 0 {
			netOUTSpeed = netOut / intervalTime
		}

		netin_chart = append(netin_chart, XYData{readTime, netINSpeed})
		netout_chart = append(netout_chart, XYData{readTime, netOUTSpeed})

	}
	chartInfos[0].ChartDatas = append(chartInfos[0].ChartDatas, cpu_chart)
	chartInfos[1].ChartDatas = append(chartInfos[1].ChartDatas, mem_chart)
	chartInfos[2].ChartDatas = append(chartInfos[2].ChartDatas, netin_chart)
	chartInfos[2].ChartDatas = append(chartInfos[2].ChartDatas, netout_chart)
	chartInfos[3].ChartDatas = append(chartInfos[3].ChartDatas, blkRead_chart)
	chartInfos[3].ChartDatas = append(chartInfos[3].ChartDatas, blkWrite_chart)
	//logger.Debugf("chartInfo  is %v", chartInfos)
	return chartInfos
}

func HandleData(cl []*ContainerInfo, intervalTime float64) {
	for index, c := range cl {
		handler := NewHandlerStatsFile(c.HostName,c.ContainerName)
		conStatss, err := handler.ReadStatsFile(c.HostName, c.ContainerName)
		if err != nil {
			logger.Errorf("ReadStatsFile Error: %s", err)
		}
		p := NewHandlerPlot(c.HostName, c.ContainerName, intervalTime)
		charInfos := p.FormatChartData(conStatss)
		logger.Debugf("hostname is %s ,The hostname contain  container'length is %d,Current index is %d ;ContainerName is %s ;charInfos' length is %d", c.HostName, len(cl), index, c.ContainerName, len(charInfos))
		for _, d := range charInfos {
			p.MakeChart(d)
		}
	}
	//FinishChart.Done()
}

// 判断文件夹是否存在,若不存在则创建
func CreatePath(path string) (error) {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(path, os.ModePerm)
		return err
	}
	return nil
}
