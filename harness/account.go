package harness

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Account struct {
	Projects  Projects
	Pipelines Pipelines
}

type Config struct {
	AccountIdentifier string     `yaml:"accountIdentifier"`
	ApiKey            string     `yaml:"apiKey"`
	TargetProjects    []string   `yaml:"targetProjects"`
	ExcludeProjects   []string   `yaml:"excludeProjects"`
	GitDetails        GitDetails `yaml:"gitDetails"`
}

type GitDetails struct {
	BranchName    string `yaml:"branch_name" json:"branch_name"`
	FilePath      string `yaml:"file_path" json:"file_path"`
	CommitMessage string `yaml:"commit_message" json:"commit_message"`
	// BaseBranch    string `yaml:"base_branch,omitempty" json:"-"`
	ConnectorRef string `yaml:"connector_ref" json:"connector_ref"`
	RepoName     string `yaml:"repo_name" json:"repo_name"`
}

func (c *Config) ReadConfig(filepath string) *Config {
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
