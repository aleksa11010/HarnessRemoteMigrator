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
	AccountIdentifier string              `yaml:"accountIdentifier"`
	ApiKey            string              `yaml:"apiKey"`
	TargetProjects    []string            `yaml:"targetProjects"`
	ExcludeProjects   []string            `yaml:"excludeProjects"`
	GitDetails        GitDetails          `yaml:"gitDetails"`
	FileStoreConfig   FileStoreConfig     `yaml:"fileStoreConfig"`
	TargetServices    []map[string]string `yaml:"targetServices"`
	ExcludeServices   []map[string]string `yaml:"excludeServices"`
}

type GitDetails struct {
	BranchName    string `yaml:"branch_name" json:"branch_name"`
	FilePath      string `yaml:"file_path" json:"file_path"`
	CommitMessage string `yaml:"commit_message" json:"commit_message"`
	// BaseBranch    string `yaml:"base_branch,omitempty" json:"-"`
	ConnectorRef string `yaml:"connector_ref" json:"connector_ref"`
	RepoName     string `yaml:"repo_name" json:"repo_name"`
}

type FileStoreConfig struct {
	Organization  string `yaml:"organization"`
	Project       string `yaml:"project"`
	Branch        string `yaml:"branch"`
	RepositoryURL string `yaml:"url"`
	ConnectorRef  string `yaml:"connector_ref" json:"connector_ref"`
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

type ApiResponse struct {
	Status           string            `json:"status"`
	Code             string            `json:"code"`
	Message          string            `json:"message"`
	CorrelationID    string            `json:"correlationId"`
	DetailedMessage  interface{}       `json:"detailedMessage"`
	ResponseMessages []ResponseMessage `json:"responseMessages"`
	Metadata         interface{}       `json:"metadata"`
}

type ResponseMessage struct {
	Code         string        `json:"code"`
	Level        string        `json:"level"`
	Message      string        `json:"message"`
	Exception    interface{}   `json:"exception"`
	FailureTypes []interface{} `json:"failureTypes"`
}
