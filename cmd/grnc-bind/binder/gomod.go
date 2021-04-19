package binder

import (
	"encoding/json"
	"fmt"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/types"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const facComponentFolder = "comp-def"
const facConfigFolder = "config"
const manifestPrefix = "manifest."

// FindExternalFacilities parses the first level of modules imported by this application's go.mod file and
// tries to find properly defined Granitic external facilities
func FindExternalFacilities(is types.StringSet, l logging.Logger, loader DefinitionLoader) (*ExternalFacilities, error) {

	cwd, _ := os.Getwd()

	if m, err := ParseModFile(cwd, l); err != nil {
		return nil, err
	} else {

		return modulesToFacilities(is, m, l, loader)
	}
}

func modulesToFacilities(is types.StringSet, mf *modFile, l logging.Logger, loader DefinitionLoader) (*ExternalFacilities, error) {

	cp, err := cachePath(l)

	if err != nil {
		return nil, err
	}

	l.LogDebugf("Expecting downloaded packages to be in %s\n", cp)

	if !folderExists(cp) {
		return nil, fmt.Errorf("expected to find downloaded modules in %s but that folder does not exist. Check your GOPATH and ensure you have run 'go mod download'", cp)
	}

	ef := new(ExternalFacilities)
	ef.Info = make([]*ExternalFacility, 0)

ModLoop:
	for _, mod := range mf.Require {

		modName := mod.Path

		var p string
		l.LogDebugf("Checking module %s %s", mod.Path, mod.Version)

		if moduleIsGranitic(mod.Path) || is.Contains(mod.Path) {
			l.LogDebugf("Ignoring module %s", mod.Path)
			continue ModLoop
		}

		local := replacePath(&mod, mf.Replace)

		if local {
			p = mod.Path
			l.LogDebugf("Using replaced path %s", p)
		} else {
			p = constructPath(mod.Path, cp, mod.Version)
		}

		l.LogDebugf("Module filesystem path is: %s", p)

		if !folderExists(p) {

			return nil, fmt.Errorf("could not find module %s %s at the expected location on the filesystem: %s check you have run 'go mod download'", mod.Path, mod.Version, p)

		}

		valid, err := validExternalFacility(p, l, loader)

		if err != nil {
			return nil, err
		}

		if valid != nil {
			l.LogDebugf("External facility found in module %s", mod.Path)
			valid.ModuleName = modName
			valid.ModulePath = p
			valid.ModuleVersion = mod.Version
			ef.Info = append(ef.Info, valid)

		} else {
			l.LogDebugf("Module %s does not contain an external facility", mod.Path)
		}

	}

	l.LogDebugf("Found %d external facility definition(s)", len(ef.Info))

	return ef, nil
}

func validExternalFacility(p string, l logging.Logger, loader DefinitionLoader) (*ExternalFacility, error) {
	fp := filepath.Join(p, "facility")

	if !folderExists(fp) {
		l.LogDebugf("No 'facility' folder found")

		return nil, nil
	}

	// Find and parse the manifest file for the facility
	mf, err := locateManifest(fp)

	if err != nil {
		return nil, fmt.Errorf("problem reading folder %s: %s", fp, err.Error())
	}

	if mf == "" {
		l.LogDebugf("No manifest file in %s assuming this is not a Granitic module", fp)
		return nil, nil
	}

	mani, err := loader.FacilityManifest(mf)

	if err != nil {
		return nil, fmt.Errorf("unable to parse manifest file at %s: %s", mf, err.Error())
	}

	ex := new(ExternalFacility)
	ex.Manifest = mani

	cfPath := filepath.Join(fp, facConfigFolder)

	populated, err := directoryAndNotEmpty(cfPath)

	if err != nil {
		return nil, fmt.Errorf("problem reading folder %s: %s", cfPath, err.Error())
	}

	if populated {
		ex.Config = cfPath
	}

	cmpPath := filepath.Join(fp, facComponentFolder)

	populated, err = directoryAndNotEmpty(cmpPath)

	if err != nil {
		return nil, fmt.Errorf("problem reading folder %s: %s", cfPath, err.Error())
	}

	if populated {
		ex.Components = cmpPath
	}

	if ex.Config == "" && ex.Components == "" {
		return nil, fmt.Errorf("%s appears to be a facility but it is malformed (must have one or both non-empty folders faility/%s and facility/%s", p, facComponentFolder, facConfigFolder)
	}

	return ex, nil

}

func moduleIsGranitic(p string) bool {

	if strings.HasPrefix(p, "github.com/graniticio/granitic/v") ||
		strings.HasPrefix(p, "github.com/graniticio/granitic-yaml/v") {

		return true
	}

	return false
}

func constructPath(modulePath string, cachePath string, version string) string {

	f := fmt.Sprintf("%s@%s", modulePath, version)

	return filepath.Join(cachePath, f)
}

func cachePath(l logging.Logger) (string, error) {
	gmc := os.Getenv("GOMODCACHE")

	if gmc != "" {
		l.LogDebugf("GOMODCACHE environment variable set. Using as location for downloaded modules")
		return gmc, nil
	}

	gmc = os.Getenv("GOPATH")

	if gmc != "" {
		l.LogDebugf("GOPATH environment variable set. Using $GOPATH/pkg/mod for downloaded modules")

		return filepath.Join(gmc, "pkg", "mod"), nil
	}

	l.LogWarnf("Neither GOPATH nor GOMODCACHE environment variable set. Assuming user home directory contains go artifacts")

	hd, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(hd, "go", "pkg", "mod"), nil
}

//ParseModFile tries to parse the mod file in the supplied directory and returns an error if parsing failed
func ParseModFile(d string, l logging.Logger) (*modFile, error) {

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)

	if err := os.Chdir(d); err != nil {
		return nil, fmt.Errorf("unable to change directory to %s: %s", d, err.Error())
	}

	goExec, err := exec.LookPath("go")

	if err != nil {
		return nil, fmt.Errorf("could not find the 'go' executable on your path. Make sure it is available in your OS PATH environment variable")
	}

	cmd := exec.Command(goExec, "mod", "edit", "--json")

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	m := modFile{}

	if err := json.NewDecoder(stdout).Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil

}

// Module represents a mod file on disk
type Module struct {
	Name         string
	Version      string
	ExpectedPath string
}

// CheckModFileExists makes sure that a go.mod file exists in the supplied directory
func CheckModFileExists(d string) bool {

	f := filepath.Join(d, "go.mod")

	info, err := os.Stat(f)

	if os.IsNotExist(err) || info.IsDir() {
		return false
	}

	return true
}

type modFile struct {
	Require []requirement
	Replace []replacement
}

// Honour any replace statements and indicate if the path has been replaced with a local filesystem path
func replacePath(req *requirement, rep []replacement) bool {

	for _, r := range rep {

		origPath := req.Path

		if r.Old.Path == origPath {

			req.Path = r.New.Path

			if r.New.Version != "" {
				req.Version = r.New.Version
			}

		}

	}

	return localPath(req.Path)
}

func localPath(p string) bool {

	if len(p) == 0 {
		return false
	}

	c := p[0]

	return c == '\\' || c == '/' || c == '.' || (len(p) > 1 && p[1] == ':')

}

type requirement struct {
	Path    string
	Version string
}

type replacement struct {
	Old modPath
	New modPath
}

type modPath struct {
	Path    string
	Version string
}

func directoryAndNotEmpty(p string) (bool, error) {

	if !folderExists(p) {

		return false, nil
	}

	d, err := ioutil.ReadDir(p)

	if err != nil {
		return false, err
	}

	return len(d) > 0, nil
}

func locateManifest(p string) (string, error) {

	d, err := ioutil.ReadDir(p)

	if err != nil {
		return "", err
	}

	for _, f := range d {
		if strings.HasPrefix(strings.ToLower(f.Name()), manifestPrefix) {
			return filepath.Join(p, f.Name()), nil
		}
	}

	return "", nil
}
