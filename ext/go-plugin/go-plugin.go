package main

import (
	"C"
	"io/ioutil"
	"os"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vagrant/ext/go-plugin/vagrant"
	"github.com/hashicorp/vagrant/ext/go-plugin/vagrant/plugin"
)

var Plugins *plugin.VagrantPlugin

//export Setup
func Setup(enableLogger, timestamps bool, logLevel *C.char) bool {
	lvl := C.GoString(logLevel)
	lopts := &hclog.LoggerOptions{Name: "vagrant"}
	if enableLogger {
		lopts.Output = os.Stderr
	} else {
		lopts.Output = ioutil.Discard
	}
	if !timestamps {
		lopts.TimeFormat = " "
	}
	lopts.Level = hclog.LevelFromString(lvl)
	vagrant.SetDefaultLogger(hclog.New(lopts))

	if Plugins != nil {
		Plugins.Logger.Error("plugins setup failure", "error", "already setup")
		return false
	}

	Plugins = &plugin.VagrantPlugin{
		PluginDirectories: []string{},
		Providers:         map[string]*plugin.RemoteProvider{},
		Logger:            vagrant.DefaultLogger().Named("go-plugin")}
	return true
}

//export LoadPlugins
func LoadPlugins(plgpath *C.char) bool {
	if Plugins == nil {
		vagrant.DefaultLogger().Error("cannot load plugins", "error", "not setup")
		return false
	}

	p := C.GoString(plgpath)
	err := Plugins.LoadPlugins(p)
	if err != nil {
		Plugins.Logger.Error("failed loading plugins",
			"path", p, "error", err)
		return false
	}
	Plugins.Logger.Info("plugins successfully loaded", "path", p)
	return true
}

//export Reset
func Reset() {
	if Plugins != nil {
		Plugins.Logger.Info("resetting loaded plugins")
		Teardown()
		dirs := Plugins.PluginDirectories
		Plugins.PluginDirectories = []string{}
		for _, p := range dirs {
			Plugins.LoadPlugins(p)
		}
	} else {
		Plugins.Logger.Warn("plugin reset failure", "error", "not setup")
	}
}

//export Teardown
func Teardown() {
	// only teardown if setup
	if Plugins == nil {
		vagrant.DefaultLogger().Error("cannot teardown plugins", "error", "not setup")
		return
	}
	Plugins.Logger.Debug("tearing down any active plugins")
	Plugins.Kill()
	Plugins.Logger.Info("plugins have been halted")
}

// stub required for build
func main() {}