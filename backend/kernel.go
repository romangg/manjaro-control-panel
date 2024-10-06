package backend

/*
#cgo pkg-config: pamac
extern void op_callback(int result);
#include "pamac-helpers.h"
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Kernel_manager struct {
	impl C.kernel_manager
	App  *application.App
}

type Kernel struct {
	Name    string
	Version string

	Installed    bool
	RealTime     bool
	Experimental bool
	Recommended  bool
	Lts          bool
	Eol          bool
	Running      bool

	Installed_modules []string
}

var Krlmgr Kernel_manager

var recommended_kernels = []string{"linux414", "linux419", "linux54", "linux510", "linux515", "linux61", "linux66"}
var lts_kernels = []string{"linux310", "linux312", "linux314", "linux316", "linux318", "linux41", "linux44", "linux49", "linux414", "linux414-rt", "linux419", "linux419-rt", "linux54", "linux510", "linux515", "linux61", "linux66"}

const pacman_kernel_regex = `^linux([0-9][0-9]?([0-9])|[0-9][0-9]?([0-9])-rt)`

func to_go_string(s *C.gchar) string {
	return C.GoString((*C.char)(s))
}

func to_go_bool(gbool C.gboolean) bool {
	return gbool != 0
}

type kernel_version struct {
	Major int
	Minor int
}

func get_kernel_version(kernel string) (*kernel_version, error) {
	kernel_split := strings.Split(kernel, ".")

	major, err := strconv.Atoi(kernel_split[0])
	if err != nil {
		return nil, errors.New("Error on kernel major version " + kernel + ": " + err.Error())
	}
	minor, err := strconv.Atoi(strings.Split(kernel_split[1], "_")[0])
	if err != nil {
		return nil, errors.New("Error on kernel minor version " + kernel + ": " + err.Error())
	}

	return &kernel_version{Major: major, Minor: minor}, nil
}

func is_newer(check string, cmp string) bool {
	version, err := get_kernel_version(check)
	if err != nil {
		fmt.Println(err)
		return false
	}
	cmp_version, err := get_kernel_version(cmp)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if version.Major == cmp_version.Major {
		if version.Minor == cmp_version.Minor {
			return strings.Contains(cmp, "rt")
		}
		return version.Minor > cmp_version.Minor
	}
	return version.Major > cmp_version.Major
}

func (mgr *Kernel_manager) Create_db() {
	C.create_db(&mgr.impl)
}

func (mgr *Kernel_manager) Get_kernels() []Kernel {
	ckernels := C.get_kernels(mgr.impl.db)
	kernels := make([]Kernel, ckernels.len)

	for i, kernel := range unsafe.Slice(ckernels.data, ckernels.len) {
		var k Kernel
		gname := C.pamac_package_get_name(kernel)

		k.Name = to_go_string(gname)
		k.Version = to_go_string(C.pamac_package_get_version(kernel))
		k.Installed = to_go_bool(C.pamac_database_is_installed_pkg(mgr.impl.db, gname))
		k.RealTime = strings.Contains(k.Name, "rt")
		k.Experimental = strings.Contains(k.Name, "rc")
		k.Recommended = slices.Contains(recommended_kernels, k.Name)
		k.Lts = slices.Contains(lts_kernels, k.Name)

		kernels[i] = k
	}

	var modules []string

	for name, version := range get_installed_packages() {
		if strings.Contains(name, "-") {
			modules = append(modules, name)
			continue
		}

		if slices.ContainsFunc(kernels, func(k Kernel) bool {
			return k.Name == name
		}) {
			continue
		}

		var k Kernel
		k.Name = name
		k.Version = version
		k.Installed = true
		k.RealTime = strings.Contains(name, "rt")
		k.Experimental = strings.Contains(name, "rc")
		k.Recommended = slices.Contains(recommended_kernels, name)
		k.Lts = slices.Contains(lts_kernels, name)
		k.Eol = true

		kernels = append(kernels, k)
	}

	running_kernel := get_running_kernel()
	running_kernel_version, _ := get_kernel_version(running_kernel.Version)

	for i := range kernels {
		k := &kernels[i]

		if version, _ := get_kernel_version(k.Version); version != nil && *running_kernel_version == *version &&
			running_kernel.RealTime == k.RealTime {
			k.Running = true
		}

		if k.Installed {
			for _, mod := range modules {
				if strings.HasPrefix(mod, k.Name) {
					k.Installed_modules = append(k.Installed_modules, mod)
				}
			}
		}
	}

	sort.Slice(kernels, func(i, j int) bool {
		return is_newer(kernels[i].Version, kernels[j].Version)
	})

	C.free_kernels(&ckernels)
	return kernels
}

//export op_callback
func op_callback(result C.int) {
	data := result == 1

	Krlmgr.App.EmitEvent("kernelOpFinished", data)
}

func (mgr *Kernel_manager) Install_kernel(name string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.install_kernel(mgr.impl.db, cName)
}

func (mgr *Kernel_manager) Remove_kernel(name string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.remove_kernel(mgr.impl.db, cName)
}

func get_installed_packages() map[string]string {
	// Prepare the command
	cmd := exec.Command("pacman", "-Qs", pacman_kernel_regex)
	cmd.Env = append(cmd.Env, "LANG=C", "LC_MESSAGES=C")

	// Run the command and capture output
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("error: failed to get installed kernels", err)
		return nil
	}

	// Parse the output
	result := out.String()
	lines := strings.Split(result, "\n")
	packages := make(map[string]string)

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, " ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		repoName := parts[0]
		pkgName := repoName[strings.Index(repoName, "/")+1:]
		pkgVersion := parts[1]

		packages[pkgName] = pkgVersion
	}

	return packages
}

// executes "uname -r" and parses the running kernel version.
func get_running_kernel() Kernel {
	// Prepare the command
	cmd := exec.Command("uname", "-r")
	cmd.Env = append(cmd.Env, "LANG=C", "LC_MESSAGES=C", "LC_ALL=C")

	// Capture the output
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("error: failed to get running kernel", err)
		return Kernel{}
	}

	// Process the output
	result := strings.TrimSpace(out.String())
	aux := strings.Split(result, ".")
	if len(aux) < 2 {
		log.Println("error: unexpected kernel version format")
		return Kernel{}
	}

	// Extract version (major.minor)
	version := aux[0] + "." + aux[1]
	kernel := Kernel{Version: version}

	// Check if the kernel is real-time (contains "-rt")
	if strings.Contains(result, "-rt") {
		kernel.RealTime = true
	}

	return kernel
}
