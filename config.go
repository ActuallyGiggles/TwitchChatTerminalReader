package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/inancgumus/screen"
	"gopkg.in/yaml.v3"
)

var config T
var configFolderLocation string
var configFileLocation string

type T struct {
	Username          string
	OAuth             string
	LastMonitoredRoom string
}

func getCredentials() {
	appdata, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	configFolderLocation = appdata + "/TwitchChatTerminalReader/"
	configFileLocation = configFolderLocation + "config.yaml"

	readCredentials()
}

func readCredentials() {
	configFile, err := os.ReadFile(configFileLocation)
	if err != nil {
		obtainCredentials()
		return
	}

	err = yaml.Unmarshal([]byte(configFile), &config)
	if err != nil {
		obtainCredentials()
		return
	}

	if config.Username == "" {
		obtainCredentials()
		return
	}
}

func updateCredentials(lastMonitoredRoom string) {
	_, err := os.Stat(configFolderLocation)
	if os.IsNotExist(err) {
		os.Mkdir(configFolderLocation, 0755)
	}

	config.LastMonitoredRoom = lastMonitoredRoom

	data, err := yaml.Marshal(&config)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(configFileLocation, data, 0644)
	if err != nil {
		panic(err)
	}
}

func obtainCredentials() {
	screen.Clear()
	screen.MoveTopLeft()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#be38d5")).Render("What account will you be using to type with?"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#9f2eec")).Render("If you want to remain anonymous, enter an empty value."))
	fmt.Println()
	fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Render("> "))
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	config.Username = strings.ToLower(scanner.Text())

	if config.Username == "" {
		return
	}

	screen.Clear()
	screen.MoveTopLeft()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#f25e92")).Render("Obtaining your OAuth is necessary to connect to Twitch chatrooms as yourself."))
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#e055a8")).Render("Instructions:"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#c850ba")).Render("1. https://twitchapps.com/tmi/"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#be38d5")).Render(`2. Click "Connect"`))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#9f2eec")).Render("3. Authenticate"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7e56f4")).Render("4. Paste the entire line starting with 'oauth' here:"))
	fmt.Println()
	fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Render("> "))
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	config.OAuth = strings.ToLower("oauth:" + strings.Replace(scanner.Text(), "oauth:", "", 1))

	updateCredentials("")
	readCredentials()
}
