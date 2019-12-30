package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"runtime"

	"github.com/asticode/go-astikit"
	astibundler "github.com/asticode/go-astilectron-bundler"
	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

var ldflags = LDFlags{}

// Flags
var (
	astilectronPath   = flag.String("a", "", "the astilectron path")
	configurationPath = flag.String("c", "", "the configuration path")
	darwin            = flag.Bool("d", false, "if set, will add darwin/amd64 to the environments")
	linux             = flag.Bool("l", false, "if set, will add linux/amd64 to the environments")
	outputPath        = flag.String("o", "", "the output path")
	windows           = flag.Bool("w", false, "if set, will add windows/amd64 to the environments")
)

func init() {
	flag.Var(ldflags, "ldflags", "extra values to concatenate onto -ldflags, eg X:main.Version=1.0.7")
}

func main() {
	// Init
	astilog.SetHandyFlags()
	cmd := astikit.FlagCmd()
	flag.Parse()
	astilog.FlagInit()

	// Get configuration path
	var cp = *configurationPath
	var err error
	if len(cp) == 0 {
		// Get working directory path
		var wd string
		if wd, err = os.Getwd(); err != nil {
			astilog.Fatal(errors.Wrap(err, "os.Getwd failed"))
		}

		// Set configuration path
		cp = filepath.Join(wd, "bundler.json")
	}

	// Open file
	var f *os.File
	if f, err = os.Open(cp); err != nil {
		astilog.Fatal(errors.Wrapf(err, "opening file %s failed", cp))
	}
	defer f.Close()

	// Unmarshal
	var c *astibundler.Configuration
	if err = json.NewDecoder(f).Decode(&c); err != nil {
		astilog.Fatal(errors.Wrap(err, "unmarshaling configuration failed"))
	}

	// Astilectron path
	if len(*astilectronPath) > 0 {
		c.AstilectronPath = *astilectronPath
	}

	// Output path
	if len(*outputPath) > 0 {
		c.OutputPath = *outputPath
	}

	// Environments
	if *darwin {
		c.Environments = append(c.Environments, astibundler.ConfigurationEnvironment{Arch: runtime.GOARCH, OS: "darwin"})
	}
	if *linux {
		c.Environments = append(c.Environments, astibundler.ConfigurationEnvironment{Arch: runtime.GOARCH, OS: "linux"})
	}
	if *windows {
		c.Environments = append(c.Environments, astibundler.ConfigurationEnvironment{Arch: runtime.GOARCH, OS: "windows"})
	}
	if len(c.Environments) == 0 {
		c.Environments = []astibundler.ConfigurationEnvironment{{Arch: runtime.GOARCH, OS: runtime.GOOS}}
	}

	// Flags
	if c.LDFlags == nil {
		c.LDFlags = astibundler.LDFlags(make(map[string][]string))
	}
	c.LDFlags.Merge(astibundler.LDFlags(ldflags))

	// Build bundler
	var b *astibundler.Bundler
	if b, err = astibundler.New(c); err != nil {
		astilog.Fatal(errors.Wrap(err, "building bundler failed"))
	}

	// Handle signals
	b.HandleSignals()

	// Switch on cmd
	switch cmd {
	case "bd":
		// Bind Data
		for _, env := range c.Environments {
			if err = b.BindData(env.OS, env.Arch); err != nil {
				astilog.Fatal(errors.Wrapf(err, "binding data failed for %s/%s", env.OS, env.Arch))
			}
		}
	case "cc":
		// Clear cache
		if err = b.ClearCache(); err != nil {
			astilog.Fatal(errors.Wrap(err, "clearing cache failed"))
		}
	default:
		// Bundle
		if err = b.Bundle(); err != nil {
			astilog.Fatal(errors.Wrap(err, "bundling failed"))
		}
	}
}
