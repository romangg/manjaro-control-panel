package main

import (
	"manjaro-control-panel/backend"
)

type HwService struct{}

func (g *HwService) Devices() []backend.Hw_device {
	return backend.Hwmgr.Pci_devices
}

func (g *HwService) InstallConfig(name string) {
	backend.Hwmgr.Install_pci_config(name)
}

func (g *HwService) RemoveConfig(name string) {
	backend.Hwmgr.Remove_pci_config(name)
}

func (g *HwService) InstallFreeGpuConfig() bool {
	return backend.Hwmgr.Install_free_gpu_config()
}

func (g *HwService) InstallProprietaryGpuConfig() bool {
	return backend.Hwmgr.Install_proprietary_gpu_config()
}
