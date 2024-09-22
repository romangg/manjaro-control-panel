package backend

/*
#cgo pkg-config: hwinfo
#include "hwinfo-helpers.h"
*/
import "C"
import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var hw_sync_pm_db = true

type Hw_manager struct {
	impl C.hw_manager
	App  *application.App

	Usb_devices, Pci_devices []Hw_device

	Installed_usb_configs, Installed_pci_configs []Hw_config
	All_usb_configs, All_pci_configs             []Hw_config
	Invalid_configs                              []Hw_config
}

var Hwmgr Hw_manager

func Fill_devices() {
	Hwmgr.Pci_devices = get_devices(Pci_kind)
	Hwmgr.Usb_devices = get_devices(Usb_kind)
}

func from_hex(hexnum uint16, fill int) string {
	return fmt.Sprintf("%0*x", fill, hexnum)
}

func from_char_array(c *C.char) string {
	if c == nil {
		return ""
	}
	return C.GoString(c)
}

func get_devices(kind Hw_kind) []Hw_device {
	var ckind C.hwtype
	if kind == Usb_kind {
		ckind = C.HWUSB
	} else {
		ckind = C.HWPCI
	}

	hw_list := C.hw_get_devices(ckind)

	var devices []Hw_device

	for hd := hw_list.first; hd != nil; hd = hd.next {
		var dev Hw_device
		dev.Kind = kind

		dev.Class_id = from_hex(uint16(hd.base_class.id), 2) +
			strings.ToLower(from_hex(uint16(hd.sub_class.id), 2))
		dev.Vendor_id = strings.ToLower(from_hex(uint16(hd.vendor.id), 4))
		dev.Subvendor_id = strings.ToLower(from_hex(uint16(hd.sub_vendor.id), 4))
		dev.Device_id = strings.ToLower(from_hex(uint16(hd.device.id), 4))

		dev.Model = from_char_array(hd.model)

		dev.Class_name = from_char_array(hd.base_class.name)
		dev.Vendor_name = from_char_array(hd.vendor.name)
		dev.Subvendor_name = from_char_array(hd.sub_vendor.name)
		dev.Device_name = from_char_array(hd.device.name)
		dev.Sysfs_bus_id = from_char_array(hd.sysfs_bus_id)
		dev.Sysfs_id = from_char_array(hd.sysfs_id)

		devices = append(devices, dev)
	}

	C.hd_free_hd_list(hw_list.first)
	C.hd_free_hd_data(hw_list.data)
	C.free(unsafe.Pointer(hw_list.data))

	return devices
}

func (mgr *Hw_manager) Install_free_gpu_config() bool {
	fmt.Println("MHWD will autodetect your open-source graphic drivers and install it.")
	return install_gpu_config("free")
}

func (mgr *Hw_manager) Install_proprietary_gpu_config() bool {
	fmt.Println("MHWD will autodetect your proprietary graphic drivers and install it.")
	return install_gpu_config("nonfree")
}

func install_gpu_config(sel string) bool {
	// Arguments for the installation process
	args := []string{"-a", "pci", sel, "0300"}

	if err := exec_mhwd(args); err != nil {
		fmt.Println("Install failed: ", err)
		return false
	}

	fmt.Println("Installation completed successfully.")
	return true
}

func (mgr *Hw_manager) Install_pci_config(name string) bool {
	fmt.Println("MHWD will install the '", name, "' configuration.")
	return exec_pci_config_op(name, "-i")
}

func (mgr *Hw_manager) Remove_pci_config(name string) bool {
	fmt.Println("MHWD will remove the '", name, "' configuration.")
	return exec_pci_config_op(name, "-r")
}

func exec_pci_config_op(name string, sel string) bool {
	// Arguments for the installation process
	args := []string{sel, "pci", name}

	if err := exec_mhwd(args); err != nil {
		fmt.Println("Operation failed: ", err)
		return false
	}

	fmt.Println("MHWD " + sel + "-operation completed successfully.")
	return true
}

func exec_mhwd(args []string) error {
	cmd := exec.Command("mhwd", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
