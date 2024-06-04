package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	cppGenerator "github.com/Wetherish/cppProjectGen/cppProjectGen"
)

type model struct {
	pathInput        textinput.Model
	projectNameInput textinput.Model
	statusMessage    string
	focusIndex       int
	quitting         bool
}

func initialModel() model {
	// Define the default path
	defaultPath := filepath.Join(os.Getenv("HOME"), "projects/")

	pathInput := textinput.New()
	pathInput.Placeholder = "Enter project path"
	pathInput.SetValue(defaultPath)
	pathInput.Focus()

	projectNameInput := textinput.New()
	projectNameInput.Placeholder = "Enter project name"

	return model{
		pathInput:        pathInput,
		projectNameInput: projectNameInput,
		focusIndex:       0,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "tab", "shift+tab", "up", "down":
			if msg.String() == "tab" || msg.String() == "down" {
				m.focusIndex++
				if m.focusIndex > 1 {
					m.focusIndex = 0
				}
			} else if msg.String() == "shift+tab" || msg.String() == "up" {
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = 1
				}
			}

			cmds := []tea.Cmd{}
			switch m.focusIndex {
			case 0:
				cmds = append(cmds, m.pathInput.Focus())
				if m.pathInput.Focused() {
					m.projectNameInput.Blur()
				}

			case 1:
				cmds = append(cmds, m.projectNameInput.Focus())
				if m.projectNameInput.Focused() {
					m.pathInput.Blur()
				}

			}

			return m, tea.Batch(cmds...)

		case "enter":
			if m.focusIndex == 1 {
				path := m.pathInput.Value()
				projectName := m.projectNameInput.Value()

				if path == "" || projectName == "" {
					m.statusMessage = errorStyle.Render("Path and project name cannot be empty")
				} else {
					m.statusMessage = successStyle.Render("Processing...")
					err := os.MkdirAll(path, os.ModePerm)
					if err != nil {
						m.statusMessage = errorStyle.Render(fmt.Sprintf("Failed to create project directory: %v", err))
					} else {
						err := cppGenerator.GenerateCppProject(path, projectName)
						if err != nil {
							m.statusMessage = errorStyle.Render(fmt.Sprintf("Failed to generate C++ project: %v", err))
						} else {
							m.statusMessage = successStyle.Render("C++ project generated successfully!")
							exec.Command("code", path+"/"+projectName).Run()
							m.quit()
							return m, tea.Quit
						}
					}
				}
				return m, nil
			}
		}
	}
	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(msg)
	m.projectNameInput, _ = m.projectNameInput.Update(msg)

	return m, cmd
}
func (m model) quit() model {
	m.quitting = true
	return m
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true).Underline(true).MarginBottom(1)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Background(lipgloss.Color("235")).Padding(0, 1).Border(lipgloss.RoundedBorder())
	focusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Background(lipgloss.Color("235")).Padding(0, 1).Border(lipgloss.RoundedBorder()).Bold(true)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Background(lipgloss.Color("235")).Padding(0, 1).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true).MarginTop(1)

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s\n\n%s\n",
		titleStyle.Render("Enter details to generate C++ project"),
		focusStyle.Render(m.pathInput.View()),
		inputStyle.Render(m.projectNameInput.View()),
		statusStyle.Render(m.statusMessage),
		hintStyle.Render("Use Tab to switch, Enter to submit, Ctrl+C or q to quit"),
	)
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Background(lipgloss.Color("235")).Bold(true)
		fmt.Printf("%s\n", errorStyle.Render(fmt.Sprintf("Error running program: %v", err)))
		os.Exit(1)
	}
}
