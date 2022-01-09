package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/akamensky/argparse"
	"github.com/m1dugh/crawler/internal/config"
)

func AddConfigCommand(parser *argparse.Parser) *argparse.Command {

	configCommand := parser.NewCommand("config", "manages config")

	addCommand := configCommand.NewCommand("add", "adds a plugin from a local file")

	addCommand.File("f", "file", os.O_RDONLY, 0444, &argparse.Options{
		Required: true,
		Help:     "the plugin file to add",
	})

	addCommand.String("t", "tag", &argparse.Options{
		Required: true,
		Help:     "the name of the plugin",
	})

	addCommand.Flag("", "mv", &argparse.Options{
		Help: "if specified, the specified file will be copied to ROOT_FOLDER/plugins",
	})

	activateCommand := configCommand.NewCommand("activate", "activates/deactivates a loaded plugin")

	activateCommand.Flag("d", "disable", &argparse.Options{
		Help: "disables a plugin instead of enabling it",
	})

	activateCommand.String("t", "tag", &argparse.Options{
		Required: true,
		Help:     "the name of the plugin",
	})

	deleteCommand := configCommand.NewCommand("remove", "removes a plugin configuration")

	deleteCommand.String("t", "tag", &argparse.Options{
		Required: true,
		Help:     "the tag of the config to delete",
	})

	listCommand := configCommand.NewCommand("list", "lists plugins registered in the config file")

	listCommand.Flag("a", "all", &argparse.Options{
		Help: "print all plugins including disabled",
	})

	listCommand.Flag("p", "path", &argparse.Options{
		Help: "print path of the plugins",
	})

	return configCommand

}

func HandleConfigCommand(configCommand *argparse.Command) {

	cfg, err := config.GetConfig()

	if err != nil {
		log.Fatal(err)
	}

	for _, command := range configCommand.GetCommands() {
		if command.Happened() {
			if command.GetName() == "add" {
				var file *os.File
				var tag string
				var mvFlag bool
				for _, arg := range command.GetArgs() {
					switch arg.GetLname() {
					case "file":
						file = arg.GetResult().(*os.File)
					case "tag":
						tag = *arg.GetResult().(*string)
					case "mv":
						mvFlag = *arg.GetResult().(*bool)
					}
				}
				handleAddCommand(&cfg, file, tag, mvFlag)
			} else if command.GetName() == "activate" {
				var disable bool
				var tag string

				for _, arg := range command.GetArgs() {
					switch arg.GetLname() {
					case "disable":
						disable = *arg.GetResult().(*bool)
					case "tag":
						tag = *arg.GetResult().(*string)
					}
				}

				handleActivateCommand(&cfg, tag, disable)
			} else if command.GetName() == "remove" {
				var tag string

				for _, arg := range command.GetArgs() {
					switch arg.GetLname() {
					case "tag":
						tag = *arg.GetResult().(*string)
					}
				}

				handleRemoveCommand(&cfg, tag)
			} else if command.GetName() == "list" {
				var all bool
				var path bool

				for _, arg := range command.GetArgs() {
					switch arg.GetLname() {
					case "all":
						all = *arg.GetResult().(*bool)
					case "path":
						path = *arg.GetResult().(*bool)
					}
				}

				handleListCommand(cfg, all, path)
			}
		}
	}
}

func handleListCommand(cfg config.Config, all, path bool) {
	for _, p := range cfg.Plugins {
		if all || p.Active {
			var message string
			if p.Active {
				message += "o "
			} else {
				message += "x "
			}

			message += p.Name
			if path {
				message += "\t=>\t" + p.Path
			}

			fmt.Println(message)

		}
	}
}

func handleRemoveCommand(cfg *config.Config, tag string) {
	found := false
	for i, p := range cfg.Plugins {
		if p.Name == tag {
			cfg.Plugins = append(cfg.Plugins[:i], cfg.Plugins[i+1:]...)
			found = true
		}
	}

	if !found {
		log.Fatal("could not find plugin ", tag)

	}
	if config.SaveConfig(*cfg) {
		fmt.Println("successfully removed", tag, "from config")
	} else {
		log.Fatal("could not save config")
	}
}

func handleActivateCommand(cfg *config.Config, tag string, disable bool) {
	found := false
	for i, plg := range cfg.Plugins {
		if plg.Name == tag {
			if disable {
				cfg.Plugins[i].Active = false
			} else {
				cfg.Plugins[i].Active = true

			}
			found = true
		}
	}

	if !found {
		log.Fatal("could not find plugin ", tag)
	}
	if config.SaveConfig(*cfg) {
		var action string
		if disable {
			action = "disabled"
		} else {
			action = "enabled"
		}

		fmt.Println("successfully", action, tag)
	} else {
		log.Fatal("could not save config")
	}
}

func handleAddCommand(cfg *config.Config, file *os.File, tag string, mvFlag bool) {
	if argparse.IsNilFile(file) {
		log.Fatal("could not find specified file")
	}

	absolutePath, err := filepath.Abs(file.Name())

	if err != nil {
		log.Fatal(err)
	}

	if mvFlag {
		defer file.Close()
		pathParts := strings.Split(absolutePath, "/")

		name := pathParts[len(pathParts)-1]

		absolutePath = path.Join(config.PLUGIN_PATH, name)
		destination, err := os.Create(absolutePath)

		if err != nil {
			log.Fatal(err)
		}
		defer destination.Close()

		_, err = io.Copy(destination, file)
		if err != nil {
			log.Fatal(err)
		}

	}
	pluginConfig := &config.PluginConfig{
		Name:   tag,
		Path:   absolutePath,
		Active: true,
	}

	cfg.Plugins = append(cfg.Plugins, pluginConfig)

	if config.SaveConfig(*cfg) {
		fmt.Println("successfully added", tag, "to config")
	} else {
		log.Fatal("could not save config")
	}
}
