package plugin

import (
	"encoding/json"
	"os"

	"github.com/IBM-Bluemix/bluemix-cli-sdk/bluemix"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/bluemix/configuration/config_helpers"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/bluemix/configuration/core_config"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/i18n"
)

// Run plugin with os.Args
func Start(plugin Plugin) {
	Run(plugin, os.Args[1:])
}

// Run plugin with args
func Run(plugin Plugin, args []string) {
	if isMetadataRequest(args) {
		metadata := fillMetadata(plugin.GetMetadata())
		json, err := json.Marshal(metadata)
		if err != nil {
			panic(err)
		}
		os.Stdout.Write(json)
		return
	}

	context := GetPluginContext(plugin.GetMetadata().Name)

	// initialization
	i18n.T = i18n.Tfunc(context.Locale())

	plugin.Run(context, args)
}

func fillMetadata(metadata PluginMetadata) PluginMetadata {
	sdkVersion := bluemix.Version
	metadata.SDKVersion = VersionType{
		Major: sdkVersion.Major,
		Minor: sdkVersion.Minor,
		Build: sdkVersion.Build,
	}
	return metadata
}

func GetPluginContext(pluginName string) PluginContext {
	coreConfig := core_config.NewCoreConfig(
		func(err error) {
			panic("configuration error: " + err.Error())
		})
	pluginPath := config_helpers.PluginDir(pluginName)
	return createPluginContext(pluginPath, coreConfig)
}

func isMetadataRequest(args []string) bool {
	return len(args) == 1 && args[0] == "SendMetadata"
}
