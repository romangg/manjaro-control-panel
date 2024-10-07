package backend

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Kernel_manager struct {
	Cache []Kernel
	App   *application.App
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

const pacman_kernel_regex = `^linux([0-9][0-9]?[0-9]|[0-9][0-9]?[0-9]-rt)`

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

func get_kernel(name string, version string) Kernel {
	var k Kernel
	k.Name = name
	k.Version = version
	k.RealTime = strings.Contains(name, "rt")
	k.Experimental = strings.Contains(name, "rc")
	k.Recommended = slices.Contains(recommended_kernels, name)
	k.Lts = slices.Contains(lts_kernels, name)
	return k
}

func (mgr *Kernel_manager) Get_kernels() []Kernel {
	var kernels []Kernel

	avail_pkgs := get_available_packages()
	instl_pkgs := get_installed_packages()

	rt_regex := regexp.MustCompile(`^linux[0-9][0-9]?[0-9]-rt$`)
	module_regex := regexp.MustCompile(`^linux[0-9][0-9]?[0-9]-(.*)`)
	module_rt_regex := regexp.MustCompile(`^linux[0-9][0-9]?[0-9]-rt-(.*)`)
	is_module := func(name string) bool {
		return module_rt_regex.MatchString(name) || (module_regex.MatchString(name) && !rt_regex.MatchString(name))
	}

	for name, version := range avail_pkgs {
		if is_module(name) {
			continue
		}

		kernel := get_kernel(name, version)

		if _, is_installed := instl_pkgs[name]; is_installed {
			kernel.Installed = true
		}

		kernels = append(kernels, kernel)
	}

	var instl_modules []string

	for name, version := range instl_pkgs {
		if is_module(name) {
			var mod_name string

			if matches := module_rt_regex.FindStringSubmatch(name); len(matches) > 1 {
				mod_name = matches[1]
			} else {
				mod_name = module_regex.FindStringSubmatch(name)[1]
			}

			if mod_name == "" {
				log.Println("error: could not identify module name for", name)
				continue
			}

			if !slices.Contains(instl_modules, mod_name) {
				instl_modules = append(instl_modules, mod_name)
			}
			continue
		}

		if slices.ContainsFunc(kernels, func(k Kernel) bool {
			return k.Name == name
		}) {
			continue
		}

		kernel := get_kernel(name, version)
		kernel.Installed = true
		kernel.Eol = true

		kernels = append(kernels, kernel)
	}

	running_kernel := get_running_kernel()
	running_kernel_version, _ := get_kernel_version(running_kernel.Version)

	for i := range kernels {
		kernel := &kernels[i]

		if version, _ := get_kernel_version(kernel.Version); version != nil && *running_kernel_version == *version &&
			running_kernel.RealTime == kernel.RealTime {
			kernel.Running = true
		}

		for _, mod := range instl_modules {
			pkg_name := kernel.Name + "-" + mod
			if _, ok := avail_pkgs[pkg_name]; ok {
				kernel.Installed_modules = append(kernel.Installed_modules, pkg_name)
			}
		}
	}

	sort.Slice(kernels, func(i, j int) bool {
		return is_newer(kernels[i].Version, kernels[j].Version)
	})

	Krlmgr.Cache = kernels
	return kernels
}

func (mgr *Kernel_manager) Install_kernel(name string) {
	pacman_install_remove_kernel(name, true)
}

func (mgr *Kernel_manager) Remove_kernel(name string) {
	pacman_install_remove_kernel(name, false)
}

func pacman_install_remove_kernel(name string, install bool) {
	op := "-S"
	op_long := "install"
	if !install {
		op = "-R"
		op_long = "remove"
	}

	find_kernel := func() *Kernel {
		for _, k := range Krlmgr.Cache {
			if k.Name == name {
				return &k
			}
		}
		return nil
	}
	kernel := find_kernel()

	if kernel == nil {
		log.Println("error: failed to identify", name, "kernel")
		Krlmgr.App.EmitEvent("kernelOpFinished", false)
		return
	}

	// Prepare the command
	args := append([]string{"/usr/bin/pacman", "--noconfirm", "--noprogressbar", op, name}, kernel.Installed_modules...)
	cmd := exec.Command("pkexec", args...)
	cmd.Env = append(cmd.Env, "LANG=C", "LC_MESSAGES=C")

	// Capture the output
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("error: failed to", op_long, "kernel:", err)
	}

	Krlmgr.App.EmitEvent("kernelOpFinished", err == nil)
}

func get_available_packages() map[string]string {
	// Prepare the command
	cmd := exec.Command("pacman", "-Ss", pacman_kernel_regex)
	cmd.Env = append(cmd.Env, "LANG=C", "LC_MESSAGES=C")

	// Run the command and capture output
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("error: failed to get available kernels", err)
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
