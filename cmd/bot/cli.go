package main

import (
	"errors"
	"flag"
	"os"
	"strings"
)

type CliOption struct {
	Name         string
	Short        string
	DefaultValue string
	Value        string
	Description  string
}

func (cmd *CliOption) StringVar() {
	flag.StringVar(&cmd.Value, cmd.Short, os.Getenv(strings.ToUpper(cmd.Name)), cmd.Description)
}

func (cmd CliOption) OrEnv() (value string, err error) {
	if cmd.Value != "" {
		value = cmd.Value
	} else {
		if cmd.DefaultValue != "" {
			value = cmd.DefaultValue
		} else {
			err = errors.New("Unable to find value")
		}
	}
	return
}

var (
	Token   CliOption
	Buttons CliOption
)

func init() {
	MissingToken = "No token provided. Please run: anibot -t <bot token>"
}

func SetupSharedOptions() {
	Token = CliOption{Name: "token", Short: "t", Description: "Bot Token"}
	Buttons = CliOption{Name: "buttons", Short: "b", Description: "Buttons path"}

	Token.StringVar()
	Buttons.StringVar()
}
