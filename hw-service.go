package main

import (
	"manjaro-control-panel/backend"
)

type HwService struct {
	manager *backend.Hw_manager
}

func (g *HwService) Devices() []backend.Hw_device {
	return backend.Hwmgr.Usb_devices
}
