// +build !client

package request

import (
	"context"
	"github.com/inexio/go-monitoringplugin"
)

func (r *CheckHardwareHealthRequest) process(ctx context.Context) (Response, error) {
	r.init()

	hhRequest := ReadHardwareHealthRequest{ReadRequest{r.BaseRequest}}
	response, err := hhRequest.process(ctx)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while processing read hardware health request", true) {
		return &CheckResponse{r.mon.GetInfo()}, nil
	}
	res := response.(*ReadHardwareHealthResponse)

	if res.EnvironmentMonitorState != nil {
		stateInt, err := (*res.EnvironmentMonitorState).GetInt()
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "read out invalid environment monitor state", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("environment_monitor_state", stateInt))
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		r.mon.UpdateStatusIf((*res.EnvironmentMonitorState) != "normal", monitoringplugin.CRITICAL, "environment monitor state is critical")
	}

	for _, fan := range res.Fans {
		if r.mon.UpdateStatusIf(fan.State == nil || fan.Description == nil, monitoringplugin.UNKNOWN, "description or state is missing for fan") {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		stateInt, err := (*fan.State).GetInt()
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "read out invalid hardware health component state for fan", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		p := monitoringplugin.NewPerformanceDataPoint("fan_state", stateInt).SetLabel(*fan.Description)
		err = r.mon.AddPerformanceDataPoint(p)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	for _, powerSupply := range res.PowerSupply {
		if r.mon.UpdateStatusIf(powerSupply.State == nil, monitoringplugin.UNKNOWN, "state is missing for power supply") {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		stateInt, err := (*powerSupply.State).GetInt()
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "read out invalid hardware health component state for power supply", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}

		p := monitoringplugin.NewPerformanceDataPoint("power_supply_state", stateInt)
		if powerSupply.Description != nil {
			p.SetLabel(*powerSupply.Description)
		} else if r.mon.UpdateStatusIf(len(res.PowerSupply) != 1, monitoringplugin.UNKNOWN, "description is missing for power supply") {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
		err = r.mon.AddPerformanceDataPoint(p)
		if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while adding performance data point", true) {
			r.mon.PrintPerformanceData(false)
			return &CheckResponse{r.mon.GetInfo()}, nil
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}
