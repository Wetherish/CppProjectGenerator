package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle          = lipgloss.NewStyle().Margin(1, 2)
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6"))
	itemStyle         = lipgloss.NewStyle().Padding(0, 1)
	selectedItemStyle = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#50FA7B"))
)

type item struct {
	title, desc, path, icon, language string
}

const (
	goIcon     = "\uea2b"  // 🎧 - Go
	cppIcon    = "\uedb8"  // 📁 - C++
	cIcon      = "\uedb1"  // 💻 - C
	pythonIcon = "\uee10"  // 🐍 - Python
	javaIcon   = "\uede3"  // 🖥️ - Java
	htmlIcon   = "\uede0"  // 🎨 - CSS
	jsIcon     = "\uede6"  // 💾 - JavaScript
	tsIcon     = "\ue73f"  // 💵 - TypeScript
	cssIcon    = "\uedbb"  // 📈 - C#
	csIcon     = "\uedb9"  // 📈 - C#
	folderIcon = "\uea04"  // 📁 - Foldear
	fileIcon   = "\ue0f6 " // 📄 - F
	mdIcon     = "\ue8e3 " // 📄 -
)

var languageIcons = map[string]string{
	".go":   goIcon,
	".cpp":  cppIcon,
	".c":    cIcon,
	".py":   pythonIcon,
	".java": javaIcon,
	".html": htmlIcon,
	".css":  cssIcon,
	".js":   jsIcon,
	".ts":   tsIcon,
	".cs":   csIcon,
	".md":   mdIcon,
}

func (i item) Title() string       { return fmt.Sprintf("%s %s", i.icon, i.title) }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type ideItem struct {
	title, command string
}

func (i ideItem) Title() string       { return i.title }
func (i ideItem) Description() string { return i.title }
func (i ideItem) FilterValue() string { return i.title }

type model struct {
	list         list.Model
	viewport     viewport.Model
	currentDir   string
	previousDir  string
	selectingIDE bool
	selectedPath string
	contentView  bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.selectingIDE {
				return m.openSelectedIDE()
			}
			return m.navigateToSelectedItem()
		case "ctrl+b":
			if !m.selectingIDE && !m.contentView {
				return m.navigateToPreviousDir()
			}
		case "ctrl+o":
			m.displayIDEOptions()
		case "ctrl+g":
			m.displayIDEOptions()
		case "esc":
			if m.contentView {
				m.contentView = false
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.viewport.Width = msg.Width - h
		m.viewport.Height = msg.Height - v
	}

	var cmd tea.Cmd
	if m.contentView {
		m.viewport, cmd = m.viewport.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	if m.contentView {
		return docStyle.Render(m.viewport.View())
	}
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true).MarginTop(1)
	hintStyle.Render("Use Tab to switch, Enter to submit, Ctrl+B to go back or q to quit")
	return docStyle.Render(m.list.View())

}

func getItems(path string) ([]item, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var items []item
	for _, entry := range entries {
		itemType := "File"
		icon := fileIcon
		language := ""
		if entry.IsDir() {
			itemType = "Directory"
			icon = folderIcon
		} else {
			ext := filepath.Ext(entry.Name())
			if val, ok := languageIcons[ext]; ok {
				icon = val
				language = ext
			}
		}
		items = append(items, item{title: entry.Name(), desc: itemType, path: filepath.Join(path, entry.Name()), icon: icon, language: language})
	}
	return items, nil
}

func execCommand(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Error executing command:", err)
		os.Exit(1)
	}
}

func (m *model) navigateToSelectedItem() (tea.Model, tea.Cmd) {
	selectedItem := m.list.SelectedItem().(item)
	fileInfo, _ := os.Stat(selectedItem.path)
	if fileInfo.IsDir() {
		newItems, err := getItems(selectedItem.path)
		if err != nil {
			fmt.Println("Error reading directory contents:", err)
			os.Exit(1)
		}
		m.selectedPath = selectedItem.path
		m.updateDir(selectedItem.path, newItems)
	} else {
		in, _ := os.ReadFile(selectedItem.path)
		content := string(in)
		var out string

		// Check the file extension and apply syntax highlighting if applicable
		ext := strings.ToLower(filepath.Ext(selectedItem.path))
		switch ext {
		case ".go", ".cpp", ".c", ".h", ".py", ".js", ".java", ".html", ".css", ".ts", ".tsx", ".cs":
			out = highlightSyntax(content, ext)
		default:
			out, _ = glamour.Render(content, "dark")
		}

		m.viewport.SetContent(out)
		m.contentView = true
	}
	return m, nil
}

func (m *model) navigateToPreviousDir() (tea.Model, tea.Cmd) {
	parentDir := filepath.Dir(m.currentDir)
	newItems, err := getItems(parentDir)
	if err != nil {
		fmt.Println("Error reading directory contents:", err)
		os.Exit(1)
	}
	m.updateDir(parentDir, newItems)
	return m, nil
}

func (m *model) updateDir(newDir string, newItems []item) {
	m.previousDir = m.currentDir
	m.currentDir = newDir
	m.list.SetItems(convertToListItems(newItems))
}

func (m *model) displayIDEOptions() {
	ides := []ideItem{
		{title: "VSCode", command: "code"},
		{title: "Neo vim", command: "nvim"},
		{title: "Rider", command: "rider"},
	}

	var listItems []list.Item
	for _, ide := range ides {
		listItems = append(listItems, ide)
	}

	m.list.SetItems(listItems)
	m.list.Title = "Choose Your IDE"
	m.selectingIDE = true
}

func (m *model) openSelectedIDE() (tea.Model, tea.Cmd) {
	selectedIDE := m.list.SelectedItem().(ideItem)
	execCommand(selectedIDE.command, m.selectedPath)
	return m, tea.Quit
}

func convertToListItems(items []item) []list.Item {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	return listItems
}

// highlightSyntax highlights the syntax of the given code content based on its file extension
func highlightSyntax(content, ext string) string {

	var lexer chroma.Lexer
	switch ext {
	case ".go":
		lexer = lexers.Get("go")
	case ".cpp", ".c", ".h":
		lexer = lexers.Get("cpp")
	case ".py":
		lexer = lexers.Get("python")
	case ".cs":
		lexer = lexers.Get("c#")
	case ".js":
		lexer = lexers.Get("javascript")
	case ".ts", ".tsx":
		lexer = lexers.Get("tsx")
	case ".java":
		lexer = lexers.Get("java")
	case ".html":
		lexer = lexers.Get("html")
	case ".css":
		lexer = lexers.Get("css")
	default:
		lexer = lexers.Analyse(content)
	}

	style := styles.Get("dracula")
	formatter := formatters.Get("terminal256")

	iterator, _ := lexer.Tokenise(nil, content)
	var sb strings.Builder
	formatter.Format(&sb, style, iterator)

	return sb.String()
}

func main() {
	projectDir := "/home/bartek/Projects"
	initialItems, err := getItems(projectDir)
	if err != nil {
		fmt.Println("Error reading directories:", err)
		os.Exit(1)
	}
	listItems := convertToListItems(initialItems)
	delegate := list.NewDefaultDelegate()

	// Styling the list items
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = selectedItemStyle
	delegate.Styles.NormalTitle = itemStyle
	delegate.Styles.NormalDesc = itemStyle

	vp := viewport.New(0, 0) // Width and height will be set later

	m := model{
		list:       list.New(listItems, delegate, 0, 0),
		viewport:   vp,
		currentDir: projectDir,
	}
	m.list.Title = titleStyle.Render("Project Directories")
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
