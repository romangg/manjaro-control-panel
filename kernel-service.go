package main

import (
	"manjaro-control-panel/backend"
)

type KernelService struct {
	manager *backend.Kernel_manager
}

func (g *KernelService) Kernels() []backend.Kernel {
	return g.manager.Get_kernels()
}

func (g *KernelService) Install(name string) {
	g.manager.Install_kernel(name)
}

func (g *KernelService) Remove(name string) {
	g.manager.Remove_kernel(name)
}
