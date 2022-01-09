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

func main() {

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	parser := argparse.NewParser("gocrawler helper", "a helper script for gocrawler")

	addCommand := parser.NewCommand("add", "adds a plugin from a local file")

	file := addCommand.File("f", "file", os.O_RDONLY, 0444, &argparse.Options{
		Required: true,
		Help:     "the plugin file to add",
	})

	name := addCommand.String("t", "tag", &argparse.Options{
		Required: true,
		Help:     "the name of the plugin",
	})

	moveFile := addCommand.Flag("", "mv", &argparse.Options{
		Help: "if specified, the specified file will be copied to ROOT_FOLDER/plugins",
	})

	activateCommand := parser.NewCommand("activate", "activates/deactivates a loaded plugin")

	activateDisable := activateCommand.Flag("d", "disable", &argparse.Options{
		Help: "disables a plugin instead of enabling it",
	})

	activateCommand.String("t", "tag", &argparse.Options{
		Required: true,
		Help:     "the name of the plugin",
	})

	if err := parser.Parse(os.Args); err != nil {
		log.Fatal(err)
	}

	if addCommand.Happened() {
		if argparse.IsNilFile(file) {
			log.Fatal("could not find specified file")
		}

		absolutePath, err := filepath.Abs(file.Name())

		if err != nil {
			log.Fatal(err)
		}

		if *moveFile {
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
			Name:   *name,
			Path:   absolutePath,
			Active: true,
		}

		cfg.Plugins = append(cfg.Plugins, pluginConfig)

	} else if activateCommand.Happened() {
		var tag string
		for _, arg := range activateCommand.GetArgs() {
			if arg.GetLname() == "tag" {
				tag = *arg.GetResult().(*string)
			}
		}

		found := false
		for i, plg := range cfg.Plugins {
			if plg.Name == tag {
				if *activateDisable {
					cfg.Plugins[i].Active = false
				} else {
					cfg.Plugins[i].Active = true

				}
				found = true
			}
		}

		if !found {
			log.Fatal("could not fing plugin ", tag)
		}

	}
	if config.SaveConfig(cfg) {
		fmt.Println("successfully saved config")
	} else {
		log.Fatal("could not save config")
	}

}
