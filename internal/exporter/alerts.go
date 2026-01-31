package exporter

import (
	"reflect"

	"gitea.libretechconsulting.com/50W/starlink-exporter/pkg/device"
)

// AlertNames is the list of all alert types by field name
var AlertNames = []string{
	// Original alerts
	"MotorsStuck",
	"ThermalThrottle",
	"ThermalShutdown",
	"MastNotNearVertical",
	"UnexpectedLocation",
	"SlowEthernetSpeeds",
	"Roaming",
	// New alert types
	"SlowEthernetSpeeds_100",
	"InstallPending",
	"IsHeating",
	"PowerSupplyThermalThrottle",
	"IsPowerSaveIdle",
	"DbfTelemStale",
	"LowMotorCurrent",
	"LowerSignalThanPredicted",
	"ObstructionMapReset",
	"DishWaterDetected",
	"RouterWaterDetected",
	"UpsuRouterPortSlow",
	"NoEthernetLink",
}

// IsAlerting returns 1.0 if the named alert is active, 0.0 otherwise
func IsAlerting(a *device.DishAlerts, alert string) float64 {
	if a == nil {
		return 0
	}
	r := reflect.ValueOf(a)
	field := reflect.Indirect(r).FieldByName(alert)
	if !field.IsValid() {
		return 0
	}
	if field.Bool() {
		return 1
	}
	return 0
}

// CountAlerts counts the number of alerts currently activated
func CountAlerts(a *device.DishAlerts) float64 {
	if a == nil {
		return 0
	}
	var firing float64
	r := reflect.ValueOf(a)
	for _, alert := range AlertNames {
		field := reflect.Indirect(r).FieldByName(alert)
		if field.IsValid() && field.Bool() {
			firing++
		}
	}
	return firing
}
