package backend

import (
	_ "embed"
	"encoding/json"
	"log"
	"os/exec"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed language-packages.json
var language_packages_json []byte

type Language_manager struct {
	App *application.App
}

type Language_package struct {
	Name                  string
	Pkg                   string
	Parent_pkgs           []string
	Parent_pkgs_installed []string
	Installed             []string
	Available             []string
}

var Lngmgr Language_manager

// Reads installed packages using `pacman -Qq`
func installed_packages() []string {
	cmd := exec.Command("pacman", "-Qq")
	output, err := cmd.Output()
	if err != nil {
		log.Println("error: failed to get installed packages (pacman)!")
		return nil
	}
	lines := strings.Split(string(output), "\n")
	var instl_pkgs []string
	for _, line := range lines {
		if len(line) > 0 {
			instl_pkgs = append(instl_pkgs, line)
		}
	}
	return instl_pkgs
}

// Reads available packages using `pacman -Sl`
func available_packages() []string {
	cmd := exec.Command("pacman", "-Sl")
	output, err := cmd.Output()
	if err != nil {
		log.Println("error: failed to get informations about available packages (pacman)!")
		return nil
	}
	lines := strings.Split(string(output), "\n")
	var available_packages []string
	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				available_packages = append(available_packages, parts[1])
			}
		}
	}
	return available_packages
}

// Intersects two lists of strings
func intersect(packages1, packages2 []string) []string {
	set := make(map[string]bool)
	for _, pkg := range packages1 {
		set[pkg] = true
	}

	var result []string
	for _, pkg := range packages2 {
		if set[pkg] {
			result = append(result, pkg)
		}
	}
	return result
}

// Filters language packages by prefix match
func filter_pkg(pkg string, packages []string) []string {
	pkg_name := pkg
	if strings.Contains(pkg, "%") {
		pkg_name = pkg[:strings.Index(pkg, "%")]
	}

	var result []string
	for _, p := range packages {
		if strings.HasPrefix(p, pkg_name) {
			result = append(result, p)
		}
	}
	return result
}

// Main logic to process language packages
func Get_language_packs() []Language_package {
	installed_pkg := installed_packages()
	available_pkg := available_packages()

	log.Println("XXX Get_language_packs")

	var json_object map[string]interface{}
	if err := json.Unmarshal(language_packages_json, &json_object); err != nil {
		log.Println("Failed to unmarshal JSON")
		return nil
	}

	pkgs_val, ok := json_object["Packages"].([]interface{})
	if !ok {
		log.Println("Invalid JSON structure")
		return nil
	}

	var lp_list []Language_package
	for _, pkg_iface := range pkgs_val {
		pkg_map := pkg_iface.(map[string]interface{})

		name := pkg_map["name"].(string)
		pkg := pkg_map["l10n_package"].(string)

		var parent_pkgs []string
		for _, parent := range pkg_map["parent_packages"].([]interface{}) {
			parent_pkgs = append(parent_pkgs, parent.(string))
		}

		parent_pkgs_installed := intersect(parent_pkgs, installed_pkg)
		pkg_installed := filter_pkg(pkg, installed_pkg)
		pkg_available := filter_pkg(pkg, available_pkg)

		lp := Language_package{
			Name:                  name,
			Pkg:                   pkg,
			Parent_pkgs:           parent_pkgs,
			Parent_pkgs_installed: parent_pkgs_installed,
			Installed:             pkg_installed,
			Available:             pkg_available,
		}
		lp_list = append(lp_list, lp)
	}

	return lp_list
}
