package monitor

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"strconv"
	//"sync"
)

type ContainstatsPlot struct {
	HostName      string
	ContainerName string
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

func NewHandlerPlot(hostName string, containName string) *ContainstatsPlot {
	if plotHandlers == nil {
		plotHandlers = make(map[string]*ContainstatsPlot)
	}
	if handler, ok := plotHandlers[containName]; ok {
		return handler
	}
	p, _ := plot.New()
	plotHandlers[containName] = &ContainstatsPlot{HostName: hostName, ContainerName: containName, Plot: p}
	return plotHandlers[containName]
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
	fileName := cp.HostName + "-" + cp.ContainerName + "-" + p.Title.Text + ".png"
	filePath := "./resultData/ChartFile/" + fileName
	err := p.Save(4*vg.Inch, 4*vg.Inch, filePath)
	if err != nil {
		logger.Errorf("MakeChart %s  Error: %s", fileName, err)
		panic(err)
	}
	// clear old data
	p.Clear()
	logger.Debugf("MakeChart %s  Success!", fileName)
}

func (cp *ContainstatsPlot) FormatChartData(cs []ContainerStatsSpec) []ChartInfo {
	//chartInfos := make([]ChartInfo)
	var chartInfos []ChartInfo
	chartInfos = append(chartInfos, ChartInfo{Title: "CPU资源占用", XLabel: "Time", YLabel: "CPU   %"})
	chartInfos = append(chartInfos, ChartInfo{Title: "Memory资源占用", XLabel: "Time", YLabel: "MEMORY   %"})
	chartInfos = append(chartInfos, ChartInfo{Title: "NET带宽Speed占用", XLabel: "Time", YLabel: "NetSpeed  mb/s"})
	chartInfos = append(chartInfos, ChartInfo{Title: "BLOCK资源占用", XLabel: "Time", YLabel: "BLOCK  MB"})

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
		previousNetIN, _ := strconv.ParseFloat(cs[i-1].NetIN, 64)
		previousNetout, _ := strconv.ParseFloat(cs[i-1].NetOUT, 64)
		netINSpeed := netIn - previousNetIN
		netOUTSpeed := netOut - previousNetout

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

func HandleData(cl []*ContainerInfo) {
	for index, c := range cl {
		handler := NewHandlerStatsFile(c.ContainerName)
		conStatss, err := handler.ReadStatsFile(c.HostName, c.ContainerName)
		if err != nil {
			logger.Errorf("ReadStatsFile Error: %s", err)
		}
		p := NewHandlerPlot(c.HostName, c.ContainerName)
		charInfos := p.FormatChartData(conStatss)
		logger.Debugf("hostname is %s ,The hostname contain  container'length is %d,Current index is %d ;ContainerName is %s ;charInfos' length is %d", c.HostName, len(cl), index, c.ContainerName, len(charInfos))
		for _, d := range charInfos {
			p.MakeChart(d)
		}
	}
	//FinishChart.Done()
}
