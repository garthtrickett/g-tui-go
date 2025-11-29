package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errmsg           struct{ err error }
	fetchedModelsMsg []list.Item
	apiResponseMsg   string
	ConcatItem       struct {
		Type string `json:"type"`
		Path string `json:"path"`
	}
)

type state int

const (
	showList state = iota
	showChat
)

type sender int

const (
	userMessage sender = iota
	geminiMessage
	systemMessage
)

type chatMessage struct {
	sender  sender
	content string
}

var (
	appStyle = lipgloss.NewStyle().Margin(1, 2)
	userStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	geminiStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	systemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
)

// --- list item ---
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// --- bubble tea model ---
type model struct {
	state         state
	list          list.Model
	textarea      textarea.Model
	spinner       spinner.Model
	client        APIClient
	selectedModel string
	messages      []chatMessage
	loading       bool
	err           error
	loadedFileContent   string
	yoloModeContent string
	yoloModeActive bool
	systemPrompt string
	pastedImage []byte
	pastedImageMimeType string
	width         int
}

func initialModel() model {
	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		tempDir := "/home/garth/.gemini/tmp/3ecdc925ecdb6b43e72d58ab048373ba41b82a51a9e0668f8dbca38696f65085"
		keyFile := filepath.Join(tempDir, "gemini", "api-key")
		content, err := os.ReadFile(keyFile)
		if err == nil {
			apiKey = string(content)
		}
	}

	if apiKey == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			keyFile := filepath.Join(home, ".config", "gemini", "api-key")
			content, err := os.ReadFile(keyFile)
			if err == nil {
				apiKey = string(content)
			}
		}
	}

	// setup list
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a Model"

		// setup text input
		ta := textarea.New()
		ta.Placeholder = ""
		ta.ShowLineNumbers = false
		ta.Focus()
	
		// setup spinner
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
		client := NewLiveAPIClient(apiKey)
	
		m := model{
			state:     showChat,
			list:      l,
			textarea: ta,
			spinner:   s,
			client:    client,
			selectedModel: "gemini-2.5-pro",
			loading:   false,
			loadedFileContent:   "",
			yoloModeContent: "",
			yoloModeActive: false,
			systemPrompt: "",
			pastedImage: nil,
			pastedImageMimeType: "",
		}
home, err := os.UserHomeDir()
	if err == nil {
		geminiMdPath := filepath.Join(home, ".gemini", "GEMINI.md")
		if _, err := os.Stat(geminiMdPath); err == nil {
			content, err := os.ReadFile(geminiMdPath)
			if err == nil {
				m.systemPrompt = string(content)
			}
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, fetchModels(m.client))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.textarea.SetWidth(msg.Width)
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			logToFile("Ctrl-C pressed")
			return m, tea.Quit
		case tea.KeyTab:
			logToFile("Tab pressed")
			userInput := m.textarea.Value()
			if strings.HasPrefix(userInput, "/read ") {
				path := strings.TrimPrefix(userInput, "/read ")
				logToFile("Path: " + path)

				if strings.HasPrefix(path, "~/") {
					home, err := os.UserHomeDir()
					if err != nil {
						logToFile(fmt.Sprintf("Error getting home dir: %v", err))
						return m, nil
					}
					path = filepath.Join(home, path[2:])
				}

				matches, err := filepath.Glob(path + "*")
				if err != nil {
					logToFile(fmt.Sprintf("Error globbing: %v", err))
					return m, nil
				}
				logToFile(fmt.Sprintf("Matches: %v", matches))

				if len(matches) == 1 {
					match := matches[0]
					info, err := os.Stat(match)
					if err != nil {
						logToFile(fmt.Sprintf("Error stating file: %v", err))
						return m, nil
					}
					if info.IsDir() {
						match += "/"
					}
					m.textarea.SetValue("/read " + match)
					m.textarea.SetCursor(len("/read " + match))
				} else if len(matches) > 1 {
					for _, match := range matches {
						info, err := os.Stat(match)
						if err == nil && info.IsDir() {
							m.textarea.SetValue("/read " + match + "/")
							m.textarea.SetCursor(len("/read " + match + "/"))
							return m, nil
						}
					}

					prefix := longestCommonPrefix(matches)
					m.textarea.SetValue("/read " + prefix)
					m.textarea.SetCursor(len("/read " + prefix))
					m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Possible completions: "+strings.Join(matches, ", ")})
				}
			}
		case tea.KeyCtrlY:
			logToFile("Ctrl-Y pressed")
			content, err := os.ReadFile("concat.json")
			if err != nil {
				logToFile(fmt.Sprintf("Error reading concat.json: %v", err))
				m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Error: concat.json not found."} )
				return m, nil
			}

			var items []ConcatItem
			if err := json.Unmarshal(content, &items); err != nil {
				logToFile(fmt.Sprintf("Error unmarshalling concat.json: %v", err))
				m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Error: Invalid concat.json format."} )
				return m, nil
			}

			outFile, err := os.Create("a.txt")
			if err != nil {
				logToFile(fmt.Sprintf("Error creating a.txt: %v", err))
				m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Error: Could not create a.txt."} )
				return m, nil
			}
			defer outFile.Close()

			for _, item := range items {
				switch item.Type {
				case "file":
					appendFile(outFile, item.Path)
				case "folder":
					filepath.Walk(item.Path, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if !info.IsDir() {
							appendFile(outFile, path)
						}
						return nil
					})
				}
			}
			m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Successfully created a.txt"})
			yoloContent, err := os.ReadFile("a.txt")
			if err != nil {
				logToFile(fmt.Sprintf("Error reading a.txt: %v", err))
				m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Error: Could not read a.txt."} )
				return m, nil
			}
			m.yoloModeContent = string(yoloContent)
			m.yoloModeActive = true
			m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "YOLO mode activated. Content of a.txt is loaded and will be sent with your next message."} )

		case tea.KeyCtrlV:
			logToFile("Ctrl-V pressed")
			cmd := exec.Command("wl-paste")
			img, err := cmd.Output()
			if err != nil {
				logToFile(fmt.Sprintf("Error running wl-paste: %v", err))
				m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Error: Could not run wl-paste. Is wl-clipboard installed?"})
				return m, nil
			}

			mimeType := http.DetectContentType(img)
			if !strings.HasPrefix(mimeType, "image/") {
				logToFile(fmt.Sprintf("Pasted content is not an image (MIME: %s)", mimeType))
				m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Error: No image found in clipboard."} )
				return m, nil
			}

			m.pastedImage = img
			m.pastedImageMimeType = mimeType
			m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Image pasted from clipboard. It will be included in your next message."} )
			return m, nil
		case tea.KeyEnter:
			if m.state == showList {
				if i, ok := m.list.SelectedItem().(item); ok {
					m.selectedModel = i.title
					m.state = showChat
				}
				return m, nil
			} else { // In chat mode, send the message
				userInput := m.textarea.Value()
				if strings.HasPrefix(userInput, "/read ") {
					filePath := strings.TrimSpace(strings.TrimPrefix(userInput, "/read "))
					content, err := os.ReadFile(filePath)
					if err != nil {
						m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Error reading file: "+err.Error()})
					} else {
						m.loadedFileContent = string(content)
						m.messages = append(m.messages, chatMessage{sender: systemMessage, content: "Loaded content from "+filePath+". It will be included in your next message."} )
					}
					m.textarea.Reset()
					m.textarea.SetValue("")
					return m, nil
				}

				m.messages = append(m.messages, chatMessage{sender: userMessage, content: "You: "+userInput})
				m.loading = true

				prompt := userInput
				if m.systemPrompt != "" {
					prompt = m.systemPrompt + "\n\n" + prompt
					m.systemPrompt = ""
				}
				if m.yoloModeActive {
					prompt = "File Content:\n" + m.yoloModeContent + "\n\n---\n\nUser Prompt:\n" + prompt
					m.yoloModeActive = false
					m.yoloModeContent = ""
				} else if m.loadedFileContent != "" {
					prompt = "File Content:\n" + m.loadedFileContent + "\n\n---\n\nUser Prompt:\n" + prompt
					m.loadedFileContent = "" // Clear after use
				}

				if m.pastedImage != nil {
					imageBase64 := base64.StdEncoding.EncodeToString(m.pastedImage)
					cmd = tea.Batch(m.spinner.Tick, makeApiCallWithImage(m.client, m.selectedModel, prompt, imageBase64, m.pastedImageMimeType))
					m.pastedImage = nil
					m.pastedImageMimeType = ""
				} else {
					cmd = tea.Batch(m.spinner.Tick, makeApiCall(m.client, m.selectedModel, prompt))
				}
				m.textarea.Reset()
				m.textarea.SetValue("")
			}
		}

	case fetchedModelsMsg:
		m.list.SetItems(msg)
		return m, nil

	case apiResponseMsg:
		m.loading = false
		if msg != "" {
			m.messages = append(m.messages, chatMessage{sender: geminiMessage, content: "Gemini: "+string(msg)})
		}
		m.textarea.Reset()
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}

	var cmds []tea.Cmd
	cmds = append(cmds, cmd)

	// Update components based on current state
	switch m.state {
	case showList:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	case showChat:
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)

		var spinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinnerCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}
	if m.client.(*LiveAPIClient).apiKey == "" {
		return "API key not found. Please set the GEMINI_API_KEY environment variable."
	}

	ss := ""
	switch m.state {
	case showList:
		ss += m.list.View()
	case showChat:
		var chatView strings.Builder
		chatView.WriteString("Chat with " + m.selectedModel + "\n\n")

		// Create a lipgloss style for wrapping
		// Subtracting 4 for margin/padding
		messageStyle := lipgloss.NewStyle().Width(m.width - 4)

		for _, msg := range m.messages {
			var styledMsg string
			switch msg.sender {
			case userMessage:
				styledMsg = userStyle.Render(msg.content)
			case geminiMessage:
				styledMsg = geminiStyle.Render(msg.content)
			case systemMessage:
				styledMsg = systemStyle.Render(msg.content)
			}
			chatView.WriteString(messageStyle.Render(styledMsg) + "\n")
		}
		
		if m.pastedImage != nil {
			chatView.WriteString("\n[Image Pasted] " + m.textarea.View())
		} else if m.yoloModeActive {
			chatView.WriteString("\n[YOLO Mode Active] " + m.textarea.View())
		} else {
			chatView.WriteString("\n" + m.textarea.View())
		}
		// Only show spinner if we're waiting for a response
		if m.loading {
			chatView.WriteString("\n" + m.spinner.View() + " Thinking...")
		}
		ss = chatView.String()
	}

	return appStyle.Render(ss)
}

func fetchModels(client APIClient) tea.Cmd {
	return func() tea.Msg {
		items, err := client.FetchModels()
		if err != nil {
			return errMsg{err}
		}
		return fetchedModelsMsg(items)
	}
}

func makeApiCall(client APIClient, model, prompt string) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GenerateContent(model, prompt)
		if err != nil {
			return errMsg{err}
		}
		return apiResponseMsg(resp)
	}
}

func makeApiCallWithImage(client APIClient, model, prompt, imageBase64, mimeType string) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GenerateContentWithImage(model, prompt, imageBase64, mimeType)
		if err != nil {
			return errMsg{err}
		}
		return apiResponseMsg(resp)
	}
}
