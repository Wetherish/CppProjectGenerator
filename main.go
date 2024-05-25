package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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
	pathInput := textinput.New()
	pathInput.Placeholder = "Enter project path"
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
					m.statusMessage = "Path and project name cannot be empty"
				} else {
					m.statusMessage = "ENTER CLICkED dziala"
					err := os.MkdirAll(path, os.ModePerm)
					if err != nil {
						m.statusMessage = fmt.Sprintf("Failed to create project directory: %v", err)
					} else {
						err := cppGenerator.GenerateCppProject(path, projectName)
						if err != nil {
							m.statusMessage = fmt.Sprintf("Failed to generate C++ project: %v", err)
						} else {
							m.statusMessage = "C++ project generated successfully!"
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

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var focusStyle = func(focused bool) string {
		if focused {
			return "focus"
		}
		return "blur"
	}

	return fmt.Sprintf(
		"Enter details to generate C++ project:\n\n%s\n\n%s\n\n%s\n\n%s",
		focusStyle(m.focusIndex == 0)+": "+m.pathInput.View(),
		focusStyle(m.focusIndex == 1)+": "+m.projectNameInput.View(),
		"Submit (Press Enter when project name is focused)",
		m.statusMessage,
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
