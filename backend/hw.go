package backend

/*
#cgo pkg-config: hwinfo
#include "hwinfo-helpers.h"
*/
import "C"
import (
	"fmt"
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
		dev.Device_id = strings.ToLower(from_hex(uint16(hd.device.id), 4))

		dev.Class_name = from_char_array(hd.base_class.name)
		dev.Vendor_name = from_char_array(hd.vendor.name)
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
