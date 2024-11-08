package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/nalgeon/codapi/internal/fileio"
)

const (
	configFilename  = "config.json"
	boxesDirname    = "boxes"
	commandsDirname = "commands"
)

// Read reads application config from JSON files.
func Read(path string) (*Config, error) {
	cfg, err := ReadConfig(filepath.Join(path, configFilename))
	if err != nil {
		return nil, err
	}

	cfg, err = ReadBoxes(cfg, filepath.Join(path, boxesDirname))
	if err != nil {
		return nil, err
	}

	cfg, err = ReadCommands(cfg, filepath.Join(path, commandsDirname))
	if err != nil {
		return nil, err
	}

	return cfg, err
}

// ReadConfig reads application config from a JSON file.
func ReadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, err
}

// ReadBoxes reads boxes config from the boxes dir
// or from the boxes.json file if the boxes dir does not exist.
func ReadBoxes(cfg *Config, path string) (*Config, error) {
	var boxes map[string]*Box
	var err error

	if fileio.Exists(path) {
		// prefer the boxes dir
		boxes, err = readBoxesDir(path)
	} else {
		// fallback to boxes.json
		boxes, err = readBoxesFile(path + ".json")
	}
	if err != nil {
		return nil, err
	}

	for _, box := range boxes {
		setBoxDefaults(box, cfg.Box)
	}

	cfg.Boxes = boxes
	return cfg, nil

}

// readBoxesDir reads boxes config from the boxes dir.
func readBoxesDir(path string) (map[string]*Box, error) {
	fnames, err := filepath.Glob(filepath.Join(path, "*.json"))
	if err != nil {
		return nil, err
	}

	boxes := make(map[string]*Box, len(fnames))
	for _, fname := range fnames {
		box, err := fileio.ReadJson[Box](fname)
		if err != nil {
			return nil, err
		}
		if box.Name == "" {
			// use the filename as the box name if it's not set
			box.Name = strings.TrimSuffix(filepath.Base(fname), ".json")
		}
		boxes[box.Name] = &box
	}

	return boxes, err
}

// readBoxesFile reads boxes config from the boxes.json file.
func readBoxesFile(path string) (map[string]*Box, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	boxes := make(map[string]*Box)
	err = json.Unmarshal(data, &boxes)
	if err != nil {
		return nil, err
	}

	return boxes, err
}

// ReadCommands reads commands config from a JSON file.
func ReadCommands(cfg *Config, path string) (*Config, error) {
	fnames, err := filepath.Glob(filepath.Join(path, "*.json"))
	if err != nil {
		return nil, err
	}

	cfg.Commands = make(map[string]SandboxCommands, len(fnames))
	for _, fname := range fnames {
		sandbox := strings.TrimSuffix(filepath.Base(fname), ".json")
		commands, err := fileio.ReadJson[SandboxCommands](fname)
		if err != nil {
			break
		}
		setCommandDefaults(commands, cfg)
		cfg.Commands[sandbox] = commands
	}

	return cfg, err
}

// setCommandDefaults applies global defaults to sandbox commands.
func setCommandDefaults(commands SandboxCommands, cfg *Config) {
	for _, cmd := range commands {
		if cmd.Before != nil {
			setStepDefaults(cmd.Before, cfg.Step)
		}
		for _, step := range cmd.Steps {
			setStepDefaults(step, cfg.Step)
		}
		if cmd.After != nil {
			setStepDefaults(cmd.After, cfg.Step)
		}
	}
}
