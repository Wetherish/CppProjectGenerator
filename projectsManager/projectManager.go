package projectsmanager

import (
	"fmt"
	"os"
	"path/filepath"
)

type Project struct {
	Name  string
	Tasks []string
}

type ProjectsManager struct {
	Projects []Project
	RootDir  string
}

func NewProjectsManager(rootDir string) (*ProjectsManager, error) {
	pm := &ProjectsManager{
		RootDir: rootDir,
	}
	err := pm.loadProjects()
	if err != nil {
		return nil, err
	}

	return pm, nil
}

func (pm *ProjectsManager) loadProjects() error {
	dirs, err := os.ReadDir(pm.RootDir)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			project := Project{
				Name: dir.Name(),
			}
			files, err := os.ReadDir(filepath.Join(pm.RootDir, dir.Name()))
			if err != nil {
				return err
			}
			for _, file := range files {
				if !file.IsDir() {
					project.Tasks = append(project.Tasks, file.Name())
				}
			}
			pm.Projects = append(pm.Projects, project)
		}
	}

	return nil
}

func (pm *ProjectsManager) ListProjects() {
	for _, project := range pm.Projects {
		fmt.Printf("Project: %s\n", project.Name)
		fmt.Println("Tasks:")
		for _, task := range project.Tasks {
			fmt.Printf("- %s\n", task)
		}
		fmt.Println()
	}
}

func main() {
	pm, err := NewProjectsManager("/path/to/projects")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	pm.ListProjects()
}
