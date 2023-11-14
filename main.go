package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"golang.org/x/term"
)

var incomingMessages chan twitchMsg

func main() {
	incomingMessages = make(chan twitchMsg)
	getCredentials()
	go twitchClientLoggedIn()
	go twitchAnonymousClient()
	go handleMessages()

	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg    error
	twitchMsg struct {
		User      string
		UserColor string
		Message   string
		IsWhisper bool
		Badge     string
		IsNotice  bool
	}
	model struct {
		viewport      viewport.Model
		messages      []string
		textarea      textarea.Model
		width, height int
		initDone      bool
	}
)

func initialModel() model {
	m := model{}
	m.viewport = viewport.New(1, 1)

	m.textarea = textarea.New()
	if config.Username != "" {
		m.textarea.Placeholder = "Send a message..."
	}
	m.textarea.Focus()
	m.textarea.Prompt = "> "
	m.textarea.ShowLineNumbers = false
	m.textarea.FocusedStyle.CursorLine = lipgloss.NewStyle()
	m.textarea.CharLimit = 500
	m.textarea.SetHeight(1)
	m.textarea.KeyMap.InsertNewline.SetEnabled(false)

	if config.LastMonitoredRoom != "" {
		m.messages = append(m.messages, lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#df56a9")).Render("Currently connected to "+config.LastMonitoredRoom+"'s chat."))
	}
	m.messages = append(m.messages, lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#d93fc0")).Render("Type '.channel [channel_name]' to join a Twitch chatroom."))
	m.messages = append(m.messages, lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#be38d5")).Render("You can only join one chatroom at a time."))
	m.messages = append(m.messages, lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#9f2eec")).Render("Be cordial and have fun.")+"\n")

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	twitchCmd := handleMessages()
	tickCmd := tickWindowSize()

	switch msg := msg.(type) {
	case tickMsg:
		w, h, _ := term.GetSize(int(os.Stdout.Fd()))

		if !m.initDone {
			m.viewport = viewport.New(w, h-4)
			var messagesFixed []string
			for _, message := range m.messages {
				messagesFixed = append(messagesFixed, wordwrap.String(message, m.width))
			}

			m.viewport.SetContent(strings.Join(messagesFixed, "\n"))

			m.textarea.SetWidth(w)

			m.width = w
			m.height = h
			m.initDone = true
		}

		if m.width != w || m.height != h {
			m.viewport = viewport.New(w, h-4)
			var messagesFixed []string
			for _, message := range m.messages {
				messagesFixed = append(messagesFixed, wordwrap.String(message, m.width))
			}

			m.viewport.SetContent(strings.Join(messagesFixed, "\n"))

			m.textarea.SetWidth(w)

			m.textarea.Reset()
			m.viewport.GotoBottom()

			m.width = w
			m.height = h

			tea.ClearScreen()
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if channel, isCommand := strings.CutPrefix(m.textarea.Value(), ".channel "); isCommand {
				Depart(config.LastMonitoredRoom)
				Join(channel)
				updateCredentials(channel)

				m.messages = []string{}
				m.messages = append(m.messages, lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#df56a9")).Render("Currently connected to "+config.LastMonitoredRoom+"'s chat."))
				m.messages = append(m.messages, lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#d93fc0")).Render("Type '.channel [channel_name]' to join a Twitch chatroom."))
				m.messages = append(m.messages, lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#be38d5")).Render("You can only join one chatroom at a time."))
				m.messages = append(m.messages, lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#9f2eec")).Render("Be cordial and have fun.")+"\n")
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
			} else {
				if config.Username != "" {
					Say(config.LastMonitoredRoom, m.textarea.Value())
				}
			}

			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	case twitchMsg:
		if msg.IsWhisper {
			var style = lipgloss.NewStyle().Foreground(lipgloss.Color(msg.UserColor)).Background(lipgloss.Color("#3C3C3C"))

			m.messages = append(m.messages, style.Render(msg.User+": "+msg.Message))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
		} else {
			var style = lipgloss.NewStyle().Foreground(lipgloss.Color(msg.UserColor))

			var badge string
			switch msg.Badge {
			case "broadcaster":
				badge = "🔴"
			case "moderator":
				badge = "⚔️"
			case "vip":
				badge = "💎"
			case "subscriber":
				badge = "➕"
			}

			m.messages = append(m.messages, badge+" "+style.Render(msg.User)+": "+msg.Message)

			if len(m.messages) > 100 {
				m.messages = m.messages[len(m.messages)-100:]
			}

			var messagesFixed []string
			for _, message := range m.messages {
				messagesFixed = append(messagesFixed, wordwrap.String(message, m.width))
			}

			m.viewport.SetContent(strings.Join(messagesFixed, "\n"))
			m.viewport.GotoBottom()
		}
	}

	return m, tea.Batch(tiCmd, vpCmd, twitchCmd, tickCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func handleMessages() tea.Cmd {
	return func() tea.Msg {
		message := <-incomingMessages
		return twitchMsg(message)
	}
}

type tickMsg bool

func tickWindowSize() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Second / 4)
		return tickMsg(true)
	}
}

func prettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}
