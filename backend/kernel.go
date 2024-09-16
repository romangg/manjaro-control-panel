package backend

/*
#cgo pkg-config: pamac
extern void op_callback(int result);
#include "pamac-helpers.h"
*/
import "C"

import (
	"errors"
	"fmt"
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
	Name         string
	Version      string
	Installed    bool
	RealTime     bool
	Experimental bool
	Recommended  bool
	Lts          bool
}

var Krlmgr Kernel_manager

var recommended_kernels = []string{"linux414", "linux419", "linux54", "linux510", "linux515", "linux61", "linux66"}
var lts_kernels = []string{"linux310", "linux312", "linux314", "linux316", "linux318", "linux41", "linux44", "linux49", "linux414", "linux414-rt", "linux419", "linux419-rt", "linux54", "linux510", "linux515", "linux61", "linux66"}

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
	minor, err := strconv.Atoi(kernel_split[1])
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

	for i, ckernel := range unsafe.Slice(ckernels.data, ckernels.len) {
		var k Kernel
		gname := C.pamac_package_get_name(ckernel)

		k.Name = to_go_string(gname)
		k.Version = to_go_string(C.pamac_package_get_version(ckernel))
		k.Installed = to_go_bool(C.pamac_database_is_installed_pkg(mgr.impl.db, gname))
		k.RealTime = strings.Contains(k.Name, "rt")
		k.Experimental = strings.Contains(k.Name, "rc")
		k.Recommended = slices.Contains(recommended_kernels, k.Name)
		k.Lts = slices.Contains(lts_kernels, k.Name)

		kernels[i] = k
	}

	sort.Slice(kernels, func(i, j int) bool {
		return is_newer(kernels[i].Version, kernels[j].Version)
	})

	C.free_kernels(&ckernels)
	return kernels
}

//export op_callback
func op_callback(result C.int) {
	fmt.Println("XXX op_callback", Krlmgr, result)
	data := result == 1

	Krlmgr.App.EmitEvent("kernelOpFinished", data)
}

func (mgr *Kernel_manager) Install_kernel(name string) {
	fmt.Println("XXX INSTALL", name)

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.install_kernel(mgr.impl.db, cName)
}

func (mgr *Kernel_manager) Remove_kernel(name string) {
	fmt.Println("XXX REMOVE", name)

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.remove_kernel(mgr.impl.db, cName)
}
