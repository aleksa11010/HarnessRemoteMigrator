package harness

import (
	"encoding/json"
	"fmt"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/google/martian/log"
)

type APIRequest struct {
	BaseURL string
	Client  *resty.Client
	APIKey  string
}

type Entities struct {
	EntityType   string
	EntityResult interface{}
}

func GetAccountIDFromAPIKey(apiKey string) string {
	accountId := strings.Split(apiKey, ".")[1]
	if accountId == "" {
		log.Errorf("Failed to get account ID from API key - %s", apiKey)
	}

	return accountId
}

func (api *APIRequest) GetAllProjects(account string) (Projects, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetQueryParams(map[string]string{
			"accountIdentifier": account,
			"hasModule":         "true",
			"pageSize":          "500",
		}).
		Get(api.BaseURL + "/ng/api/projects")
	if err != nil {
		return Projects{}, err
	}
	projects := Projects{}
	err = json.Unmarshal(resp.Body(), &projects)
	if err != nil {
		return Projects{}, err
	}

	return projects, nil
}

func (api *APIRequest) GetAllPipelines(account, org, project string) (Pipelines, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetBody(`{"filterType": "PipelineSetup"}`).
		SetQueryParams(map[string]string{
			"accountIdentifier": account,
			"orgIdentifier":     org,
			"projectIdentifier": project,
			"size":              "1000",
		}).
		Post(api.BaseURL + "/pipeline/api/pipelines/list")
	if err != nil {
		return Pipelines{}, err
	}
	pipelines := Pipelines{}
	err = json.Unmarshal(resp.Body(), &pipelines)
	if err != nil {
		return Pipelines{}, err
	}

	return pipelines, nil
}

func (api *APIRequest) GetAllTemplates(account, org, project string) (Templates, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetHeader("Harness-Account", account).
		SetQueryParams(map[string]string{
			"orgIdentifier":     org,
			"projectIdentifier": project,
			"limit":             "1000",
		}).
		Get(api.BaseURL + fmt.Sprintf("/v1/orgs/%s/projects/%s/templates", org, project))
	if err != nil {
		return Templates{}, err
	}
	templates := Templates{}
	err = json.Unmarshal(resp.Body(), &templates)
	if err != nil {
		return Templates{}, err
	}

	return templates, nil
}

func (p *PipelineContent) MovePipelineToRemote(api *APIRequest, c Config, org, project string) (string, error) {
	type RequestBody struct {
		GitDetails              GitDetails `json:"git_details"`
		PipelineIdentifier      string     `json:"pipeline_identifier"`
		MoveConfigOperationType string     `json:"move_config_operation_type"`
	}

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Harness-Account", c.AccountIdentifier).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"org":      org,
			"project":  project,
			"pipeline": p.Identifier,
		}).
		SetBody(RequestBody{
			GitDetails:              c.GitDetails,
			PipelineIdentifier:      p.Identifier,
			MoveConfigOperationType: "INLINE_TO_REMOTE",
		}).
		Post(api.BaseURL + fmt.Sprintf("/v1/orgs/%s/projects/%s/pipelines/%s/move-config", org, project, p.Identifier))

	if resp.StatusCode() != 200 {
		err = fmt.Errorf(string(resp.Body()))
		return "", err
	}

	return string(resp.Body()), err
}

func (t *Template) MoveTemplateToRemote(api *APIRequest, c Config, org, project string) (string, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Harness-Account", c.AccountIdentifier).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"templateIdenfier":  t.Identifier,
			"accountIdentifier": c.AccountIdentifier,
			"projectIdentifier": project,
			"orgIdentifier":     t.Org,
			"versionlabel":      t.VersionLabel,
			"connectorRef":      c.GitDetails.ConnectorRef,
			"repoName":          c.GitDetails.RepoName,
			"branch":            c.GitDetails.BranchName,
			"filePath":          c.GitDetails.FilePath,
			"commitMsg":         c.GitDetails.CommitMessage,
			"moveConfigType":    "INLINE_TO_REMOTE",
		}).
		Post(api.BaseURL + fmt.Sprintf("/template/api/templates/move-config/%s", t.Identifier))

	if resp.StatusCode() != 200 {
		err = fmt.Errorf(string(resp.Body()))
		return "", err
	}

	return string(resp.Body()), err
}
