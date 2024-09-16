package backend

/*
#cgo pkg-config: hwinfo
#include "hwinfo-helpers.h"
*/
import "C"
import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const hw_mhwd_cfg_name = "MHWDCONFIG"
const hw_mhwd_usb_cfg_dir = "/var/lib/mhwd/db/usb"
const hw_mhwd_pci_cfg_dir = "/var/lib/mhwd/db/pci"
const hw_mhwd_usb_db_dir = "/var/lib/mhwd/local/usb"
const hw_mhwd_pci_db_dir = "/var/lib/mhwd/local/pci"
const hw_mhwd_script_path = "/var/lib/mhwd/scripts/mhwd"

const hw_pm_cache_dir = "/var/cache/pacman/pkg"
const hw_pm_config = "/etc/pacman.conf"
const hw_pm_root = "/"

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

type Hw_kind int

const (
	Usb_kind Hw_kind = iota
	Pci_kind
)

type Hw_device struct {
	Kind Hw_kind

	Class_name, Class_id   string
	Device_name, Device_id string
	Vendor_name, Vendor_id string
	Sysfs_bus_id, Sysfs_id string

	Available_configs, Installed_configs []*Hw_config
}

type Hw_ids struct {
	Class_ids, Vendor_ids, Device_ids []string
}

type Hw_config_ids struct {
	Hw        Hw_ids
	Blacklist Hw_ids
}

type Hw_config struct {
	Kind Hw_kind

	Hw []Hw_config_ids

	Base_path, Config_path  string
	Name, Info, Version     string
	Freedriver              bool
	Priority                int
	Conflicts, Dependencies []string
}

func fill_devices() {
	Hwmgr.Pci_devices = get_devices(Pci_kind)
	Hwmgr.Usb_devices = get_devices(Usb_kind)
}

func update_configs() {
	// Clear config vectors in each device element
	for _, dev := range Hwmgr.Pci_devices {
		dev.Available_configs = nil
	}
	for _, dev := range Hwmgr.Usb_devices {
		dev.Available_configs = nil
	}

	Hwmgr.All_pci_configs = nil
	Hwmgr.All_usb_configs = nil

	// Refill data
	fill_all_configs(Pci_kind)
	fill_all_configs(Usb_kind)

	set_matching_configs(&Hwmgr.Pci_devices, &Hwmgr.All_pci_configs, false)
	set_matching_configs(&Hwmgr.Usb_devices, &Hwmgr.All_usb_configs, false)

	// // Update also installed config data
	// updateInstalledConfigData(data)
}

func fill_all_configs(kind Hw_kind) {
	var config_paths []string
	var configs *[]Hw_config

	if kind == Usb_kind {
		configs = &Hwmgr.All_usb_configs
		config_paths = get_recursive_directory_file_list(hw_mhwd_usb_cfg_dir, hw_mhwd_cfg_name)
	} else {
		configs = &Hwmgr.All_pci_configs
		config_paths = get_recursive_directory_file_list(hw_mhwd_pci_cfg_dir, hw_mhwd_cfg_name)
	}

	for _, path := range config_paths {
		var cfg Hw_config
		if fill_config(&cfg, path, kind) {
			*configs = append(*configs, cfg)
		} else {
			Hwmgr.Invalid_configs = append(Hwmgr.Invalid_configs, cfg)
		}
	}
}

func new_hw_config_ids() Hw_config_ids {
	return Hw_config_ids{Hw: Hw_ids{}, Blacklist: Hw_ids{}}
}

func fill_config(cfg *Hw_config, config_path string, kind Hw_kind) bool {
	cfg.Kind = kind
	cfg.Priority = 0
	cfg.Freedriver = true
	cfg.Base_path = config_path[:strings.LastIndex(config_path, "/")]
	cfg.Config_path = config_path

	cfg.Hw = append(cfg.Hw, new_hw_config_ids())

	return read_config_file(cfg, config_path)
}

func read_config_file(config *Hw_config, config_path string) bool {
	if len(config.Hw) == 0 {
		panic("Config Hw is empty")
	}

	file, err := os.Open(config_path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Remove comments
		if pos := strings.Index(line, "#"); pos != -1 {
			line = line[:pos]
		}

		// Trim spaces and check for empty line
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split line by `=`
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(strings.Trim(parts[1], "\""))

		// Handle external file with `>`
		if len(value) > 1 && value[0] == '>' {
			ext_file_path := get_right_config_path(value[1:], config.Base_path)
			ext_file, err := os.Open(ext_file_path)
			if err != nil {
				return false
			}
			defer ext_file.Close()

			ext_scanner := bufio.NewScanner(ext_file)
			var ext_value strings.Builder

			for ext_scanner.Scan() {
				ext_line := ext_scanner.Text()

				if pos := strings.Index(ext_line, "#"); pos != -1 {
					ext_line = ext_line[:pos]
				}

				ext_line = strings.TrimSpace(ext_line)
				if ext_line != "" {
					ext_value.WriteString(" ")
					ext_value.WriteString(ext_line)
				}
			}

			value = strings.TrimSpace(ext_value.String())

			// Remove multiple spaces
			for strings.Contains(value, "  ") {
				value = strings.Replace(value, "  ", " ", -1)
			}
		}

		// Process each key
		switch key {
		case "include":
			read_config_file(config, get_right_config_path(value, config.Base_path))
		case "name":
			config.Name = strings.ToLower(value)
		case "version":
			config.Version = value
		case "info":
			config.Info = value
		case "priority":
			priority, err := strconv.Atoi(value)
			if err == nil {
				config.Priority = priority
			}
		case "freedriver":
			if strings.ToLower(value) == "true" {
				config.Freedriver = true
			} else if strings.ToLower(value) == "false" {
				config.Freedriver = false
			}
		case "classids":
			// Add new HardwareIDs group to slice if not empty
			if len(config.Hw[len(config.Hw)-1].Hw.Class_ids) != 0 {
				config.Hw = append(config.Hw, new_hw_config_ids())
			}
			config.Hw[len(config.Hw)-1].Hw.Class_ids = split_value(value, "")
		case "vendorids":
			// Add new HardwareIDs group to slice if not empty
			if len(config.Hw[len(config.Hw)-1].Hw.Vendor_ids) != 0 {
				config.Hw = append(config.Hw, new_hw_config_ids())
			}
			config.Hw[len(config.Hw)-1].Hw.Vendor_ids = split_value(value, "")
		case "deviceids":
			// Add new HardwareIDs group to slice if not empty
			if len(config.Hw[len(config.Hw)-1].Hw.Device_ids) != 0 {
				config.Hw = append(config.Hw, new_hw_config_ids())
			}
			config.Hw[len(config.Hw)-1].Hw.Device_ids = split_value(value, "")
		case "blacklistedclassids":
			config.Hw[len(config.Hw)-1].Blacklist.Class_ids = split_value(value, "")
		case "blacklistedvendorids":
			config.Hw[len(config.Hw)-1].Blacklist.Vendor_ids = split_value(value, "")
		case "blacklisteddeviceids":
			config.Hw[len(config.Hw)-1].Blacklist.Device_ids = split_value(value, "")
		case "mhwddepends":
			config.Dependencies = split_value(value, "")
		case "mhwdconflicts":
			config.Conflicts = split_value(value, "")
		}
	}

	// Append "*" to all empty vectors
	for i := range config.Hw {
		if len(config.Hw[i].Hw.Class_ids) == 0 {
			config.Hw[i].Hw.Class_ids = append(config.Hw[i].Hw.Class_ids, "*")
		}
		if len(config.Hw[i].Hw.Vendor_ids) == 0 {
			config.Hw[i].Hw.Vendor_ids = append(config.Hw[i].Hw.Vendor_ids, "*")
		}
		if len(config.Hw[i].Hw.Device_ids) == 0 {
			config.Hw[i].Hw.Device_ids = append(config.Hw[i].Hw.Device_ids, "*")
		}
	}

	// If the name is empty, return false
	return config.Name != ""
}

// returns the correct path by prepending base if necessary
func get_right_config_path(str, base string) string {
	str = strings.TrimSpace(str)

	if len(str) == 0 || strings.HasPrefix(str, "/") {
		return str
	}

	return filepath.Join(base, str)
}

func set_matching_configs(devices *[]Hw_device, configs *[]Hw_config, set_as_installed bool) {
	for _, cfg := range *configs {
		set_matching_config(devices, &cfg, set_as_installed)
	}
}

func set_matching_config(devices *[]Hw_device, config *Hw_config, set_as_installed bool) {
	found_devices := get_devices_of_config(devices, config)

	// Set config to all matching devices
	for _, dev := range found_devices {
		if set_as_installed {
			add_config_sorted(&dev.Installed_configs, config)
		} else {
			add_config_sorted(&dev.Available_configs, config)
		}
	}
}

// Splits a string by spaces and processes it based on the 'only_ending' suffix
func split_value(str, only_ending string) []string {
	// Convert the input string to lowercase and split it by spaces
	work := strings.Fields(strings.ToLower(str))
	var final []string

	for _, item := range work {
		if item != "" && only_ending == "" {
			final = append(final, item)
		} else if item != "" && strings.HasSuffix(item, "."+only_ending) && len(item) > 5 {
			// Remove the last 5 characters from the item
			final = append(final, item[:len(item)-5])
		}
	}

	return final
}

// Adds a config to the slice of configs, ensuring the slice remains sorted by priority
func add_config_sorted(configs *[]*Hw_config, config *Hw_config) {
	for _, exist := range *configs {
		if config.Name == exist.Name {
			// Config with the same name already exists
			return
		}
	}

	for i, exist := range *configs {
		if config.Priority > exist.Priority {
			// Insert config into the slice while maintaining priority order
			*configs = append((*configs)[:i], append([]*Hw_config{config}, (*configs)[i:]...)...)
			return
		}
	}

	// If no higher priority was found, append the config to the end
	*configs = append(*configs, config)
}

func get_devices_of_config(devices *[]Hw_device, config *Hw_config) []*Hw_device {
	var found_devices []*Hw_device

	// Loop through the hardware ids in the config
	for _, ids := range config.Hw {
		found_device := false

		// Loop through each device
		for _, device := range *devices {
			found := false

			// Check class IDs
			for _, id := range ids.Hw.Class_ids {
				if id == "*" || id == device.Class_id {
					found = true
					break
				}
			}
			if !found {
				continue
			}

			// Check blacklisted class IDs
			found = false
			for _, id := range ids.Blacklist.Class_ids {
				if id == device.Class_id {
					found = true
					break
				}
			}
			if found {
				continue
			}

			// Check vendor IDs
			found = false
			for _, id := range ids.Hw.Vendor_ids {
				if id == "*" || id == device.Vendor_id {
					found = true
					break
				}
			}
			if !found {
				continue
			}

			// Check blacklisted vendor IDs
			found = false
			for _, id := range ids.Blacklist.Vendor_ids {
				if id == device.Vendor_id {
					found = true
					break
				}
			}
			if found {
				continue
			}

			// Check device IDs
			found = false
			for _, id := range ids.Hw.Device_ids {
				if id == "*" || id == device.Device_id {
					found = true
					break
				}
			}
			if !found {
				continue
			}

			// Check blacklisted device IDs
			found = false
			for _, id := range ids.Blacklist.Device_ids {
				if id == device.Device_id {
					found = true
					break
				}
			}
			if found {
				continue
			}

			found_device = true
			found_devices = append(found_devices, &device)
		}

		// If no device found for the current HardwareIDs, clear the foundDevices and return
		if !found_device {
			return nil
		}
	}

	return found_devices
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

// returns a list of file paths in the directory and its subdirectories that match the given
// filename (if provided).
func get_recursive_directory_file_list(dir_path string, only_filename string) []string {
	var list []string

	// Open the directory
	dir, err := os.Open(dir_path)
	if err != nil {
		return list
	}
	defer dir.Close()

	// Read directory contents
	files, err := dir.Readdir(-1)
	if err != nil {
		return list
	}

	for _, file := range files {
		filename := file.Name()
		filepath := filepath.Join(dir_path, filename)

		// Skip "." and ".."
		if filename == "." || filename == ".." {
			continue
		}

		// If the file is a regular file, check if it matches the filename filter
		if file.Mode().IsRegular() && (only_filename == "" || only_filename == filename) {
			list = append(list, filepath)

			// If the file is a directory, recurse into it
		} else if file.IsDir() {
			sublist := get_recursive_directory_file_list(filepath, only_filename)
			list = append(list, sublist...)
		}
	}

	return list
}
