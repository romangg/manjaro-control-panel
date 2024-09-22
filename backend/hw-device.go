package backend

type Hw_kind int

const (
	Usb_kind Hw_kind = iota
	Pci_kind
)

type Hw_device struct {
	Kind Hw_kind

	Model                        string
	Class_name, Class_id         string
	Device_name, Device_id       string
	Vendor_name, Vendor_id       string
	Subvendor_name, Subvendor_id string
	Sysfs_bus_id, Sysfs_id       string

	Available_configs, Installed_configs []*Hw_config
}

func get_devices_of_config(devices *[]Hw_device, config *Hw_config) []*Hw_device {
	var found_devices []*Hw_device

	// Loop through the hardware ids in the config
	for _, ids := range config.Hw {
		found_device := false

		// Loop through each device
		for index := range *devices {
			device := &(*devices)[index]
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
			found_devices = append(found_devices, device)
		}

		// If no device found for the current HardwareIDs, clear the foundDevices and return
		if !found_device {
			return nil
		}
	}

	return found_devices
}
