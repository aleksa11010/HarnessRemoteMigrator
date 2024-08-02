package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/google/martian/log"
)

type APIRequest struct {
	BaseURL string
	Client  *resty.Client
	APIKey  string
}

func GetAccountIDFromAPIKey(apiKey string) string {
	accountId := strings.Split(apiKey, ".")[1]
	if accountId == "" {
		log.Errorf("Failed to get account ID from API key - %s", apiKey)
	}

	return accountId
}

func GetServiceManifestStoreType(connectorType string) string {
	if connectorType == "Gitlab" {
		return "GitLab"
	}
	return connectorType
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

func (api *APIRequest) GetInputsets(account, org, project, pipeline string) ([]*InputsetContent, error) {

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"routingId":          account,
			"accountIdentifier":  account,
			"orgIdentifier":      org,
			"projectIdentifier":  project,
			"pipelineIdentifier": pipeline,
			"size":               "1000",
		}).
		Get(api.BaseURL + "/gateway/pipeline/api/inputSets")
	if err != nil {
		return nil, err
	}

	result := ListInputsetResponse{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}

	return result.Data.Content, nil
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
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return "", err
		}
		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		return "", fmt.Errorf(errMsg)
	}

	return string(resp.Body()), err
}

func (t *Template) MoveTemplateToRemote(api *APIRequest, c Config) (string, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Harness-Account", c.AccountIdentifier).
		SetHeader("Content-Type", "application/json").
		SetPathParam("templateIdentifier", t.Identifier).
		SetQueryParams(map[string]string{
			"accountIdentifier": c.AccountIdentifier,
			"projectIdentifier": t.Project,
			"orgIdentifier":     t.Org,
			"versionLabel":      t.VersionLabel,
			"connectorRef":      c.GitDetails.ConnectorRef,
			"repoName":          c.GitDetails.RepoName,
			"branch":            c.GitDetails.BranchName,
			"isNewBranch":       "false",
			"filePath":          c.GitDetails.FilePath,
			"commitMsg":         c.GitDetails.CommitMessage,
			"moveConfigType":    "INLINE_TO_REMOTE",
		}).
		Post(api.BaseURL + "/template/api/templates/move-config/{templateIdentifier}")

	if resp.StatusCode() != 200 {
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return "", err
		}
		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		return "", fmt.Errorf(errMsg)
	}

	return string(resp.Body()), err
}

func (s *ServiceClass) MoveServiceToRemote(api *APIRequest, c Config) (string, bool, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetPathParam("serviceIdentifier", s.Identifier).
		SetQueryParams(map[string]string{
			"accountIdentifier": c.AccountIdentifier,
			"projectIdentifier": s.Project,
			"orgIdentifier":     s.Org,
			"connectorRef":      c.GitDetails.ConnectorRef,
			"repoName":          c.GitDetails.RepoName,
			"branch":            c.GitDetails.BranchName,
			"isNewBranch":       "false",
			"filePath":          c.GitDetails.FilePath,
			"commitMsg":         c.GitDetails.CommitMessage,
			"moveConfigType":    "INLINE_TO_REMOTE",
		}).
		Post(api.BaseURL + "/gateway/ng/api/servicesV2/move-config/{serviceIdentifier}")

	if resp.StatusCode() != 200 {
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return "", false, err
		}

		// WHEN A SERVICE IS ALREADY REMOTE WE DON'T REPORT IT AS ERROR
		if len(ar.ResponseMessages) == 1 && strings.Contains(ar.ResponseMessages[0].Message, "is already remote") {
			return "", true, nil
		}

		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		return "", false, fmt.Errorf(errMsg)
	}

	return string(resp.Body()), false, err
}

func (e *EnvironmentClass) MoveEnvironmentToRemote(api *APIRequest, c Config) error {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetPathParam("environmentIdentifier", e.Identifier).
		SetQueryParams(map[string]string{
			"accountIdentifier": c.AccountIdentifier,
			"projectIdentifier": e.ProjectIdentifier,
			"orgIdentifier":     e.OrgIdentifier,
			"connectorRef":      c.GitDetails.ConnectorRef,
			"repoName":          c.GitDetails.RepoName,
			"branch":            c.GitDetails.BranchName,
			"isNewBranch":       "false",
			"isHarnessCodeRepo": "false",
			"filePath":          c.GitDetails.FilePath,
			"commitMsg":         c.GitDetails.CommitMessage,
			"moveConfigType":    "INLINE_TO_REMOTE",
		}).
		Post(api.BaseURL + "/gateway/ng/api/environmentsV2/move-config/{environmentIdentifier}")

	if resp.StatusCode() != 200 {
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return err
		}
		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		return fmt.Errorf(errMsg)
	}

	return err
}

func (is *InputsetContent) MoveInputsetToRemote(api *APIRequest, c Config, project, org string) error {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetPathParam("identifier", is.Identifier).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"accountIdentifier":  c.AccountIdentifier,
			"projectIdentifier":  project,
			"orgIdentifier":      org,
			"pipelineIdentifier": is.PipelineIdentifier,
			"inputSetIdentifier": is.Identifier,
			"connectorRef":       c.GitDetails.ConnectorRef,
			"repoName":           c.GitDetails.RepoName,
			"branch":             c.GitDetails.BranchName,
			"isNewBranch":        "false",
			"isHarnessCodeRepo":  "false",
			"filePath":           c.GitDetails.FilePath,
			"commitMsg":          c.GitDetails.CommitMessage,
			"moveConfigType":     "INLINE_TO_REMOTE",
		}).
		Post(api.BaseURL + "/gateway/pipeline/api/inputSets/move-config/{identifier}")

	if resp.StatusCode() != 200 {
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return err
		}
		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		return fmt.Errorf(errMsg)
	}

	return err
}

func (i *Infrastructure) MoveInfrastructureToRemote(api *APIRequest, c Config, envId string) error {

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetPathParam("infraIdentifier", i.Identifier).
		SetQueryParams(map[string]string{
			"accountIdentifier":     c.AccountIdentifier,
			"projectIdentifier":     i.ProjectIdentifier,
			"orgIdentifier":         i.OrgIdentifier,
			"environmentIdentifier": envId,
			"connectorRef":          c.GitDetails.ConnectorRef,
			"repoName":              c.GitDetails.RepoName,
			"branch":                c.GitDetails.BranchName,
			"isNewBranch":           "false",
			"isHarnessCodeRepo":     "false",
			"filePath":              c.GitDetails.FilePath,
			"commitMsg":             c.GitDetails.CommitMessage,
			"moveConfigType":        "INLINE_TO_REMOTE",
		}).
		Post(api.BaseURL + "/gateway/ng/api/infrastructures/move-config/{infraIdentifier}")

	if resp.StatusCode() != 200 {
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return err
		}
		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		return fmt.Errorf(errMsg)
	}

	return err
}

func (ov *OverridesV2Content) MoveToRemote(api *APIRequest, c Config) error {

	params := map[string]string{
		"accountIdentifier":    c.AccountIdentifier,
		"projectIdentifier":    ov.ProjectIdentifier,
		"orgIdentifier":        ov.OrgIdentifier,
		"connectorRef":         c.GitDetails.ConnectorRef,
		"repoName":             c.GitDetails.RepoName,
		"branch":               c.GitDetails.BranchName,
		"isNewBranch":          "false",
		"isHarnessCodeRepo":    "false",
		"filePath":             c.GitDetails.FilePath,
		"commitMsg":            c.GitDetails.CommitMessage,
		"moveConfigType":       "INLINE_TO_REMOTE",
		"serviceOverridesType": string(ov.Type),
		"identifier":           ov.Identifier,
	}

	if len(ov.EnvironmentRef) != 0 {
		params["environmentRef"] = ov.EnvironmentRef
	}
	if len(ov.ServiceRef) > 0 {
		params["serviceRef"] = ov.ServiceRef
	}
	if len(ov.InfraIdentifier) > 0 {
		params["infraIdentifier"] = ov.InfraIdentifier
	}

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetQueryParams(params).
		Post(api.BaseURL + "/gateway/ng/api/serviceOverrides/move-config")

	if resp.StatusCode() != 200 {
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return err
		}
		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		return fmt.Errorf(errMsg)
	}

	return err
}

func (api *APIRequest) GetAllOrgs(account string) (Organizations, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Harness-Account", account).
		SetQueryParams(map[string]string{
			"limit": "1000",
		}).
		Get(api.BaseURL + "/v1/orgs")
	if err != nil {
		return Organizations{}, err
	}

	organizations := Organizations{}
	err = json.Unmarshal(resp.Body(), &organizations)
	if err != nil {
		return Organizations{}, err
	}

	return organizations, nil
}

func (api *APIRequest) GetAllAccountFiles(account string) ([]FileStoreContent, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"accountIdentifier": account,
			"pageSize":          "2000",
		}).
		Get(api.BaseURL + "/ng/api/file-store")
	if err != nil {
		return []FileStoreContent{}, err
	}

	fileStore := FileStore{}
	err = json.Unmarshal(resp.Body(), &fileStore)
	if err != nil {
		return []FileStoreContent{}, err
	}

	return fileStore.Data.Content, nil
}

func (api *APIRequest) GetAllOrgFiles(account, org string) ([]FileStoreContent, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"accountIdentifier": account,
			"orgIdentifier":     org,
			"pageSize":          "2000",
		}).
		Get(api.BaseURL + "/ng/api/file-store")
	if err != nil {
		return []FileStoreContent{}, err
	}

	fileStore := FileStore{}
	err = json.Unmarshal(resp.Body(), &fileStore)
	if err != nil {
		return []FileStoreContent{}, err
	}

	return fileStore.Data.Content, nil
}

func (api *APIRequest) GetAllProjectFiles(account, org, project string) ([]FileStoreContent, error) {
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"accountIdentifier": account,
			"orgIdentifier":     org,
			"projectIdentifier": project,
			"pageSize":          "2000",
		}).
		Get(api.BaseURL + "/ng/api/file-store")
	if err != nil {
		return []FileStoreContent{}, err
	}

	fileStore := FileStore{}
	err = json.Unmarshal(resp.Body(), &fileStore)
	if err != nil {
		return []FileStoreContent{}, err
	}

	return fileStore.Data.Content, nil
}

func (f *FileStoreContent) DownloadFile(api *APIRequest, account, org, project, folder string) error {
	var params map[string]string
	if project == "" && org == "" {
		params = map[string]string{
			"accountIdentifier": account,
		}
	} else if org != "" && project == "" {
		params = map[string]string{
			"accountIdentifier": account,
			"orgIdentifier":     org,
		}
	} else {
		params = map[string]string{
			"accountIdentifier": account,
			"orgIdentifier":     org,
			"projectIdentifier": project,
		}
	}

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(params).
		SetPathParam("id", f.Identifier).
		Get(api.BaseURL + "/ng/api/file-store/files/{id}/download")
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return err
		}
		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		if !strings.Contains(errMsg, "Downloading folder not supported") {
			return fmt.Errorf(errMsg)
		}
	}

	if !strings.Contains(f.Path, ".") {
		return nil
	}

	err = os.MkdirAll(filepath.Dir("./filestore/filestore/"+folder+f.Path), 0755)
	if err != nil {
		return err
	}

	out, err := os.Create("./filestore/filestore/" + folder + f.Path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(resp.Body())
	if err != nil {
		return err
	}

	return nil
}

func (api *APIRequest) GetConnector(account, org, project, identifier string) (ConnectorClass, error) {
	params := map[string]string{
		"accountIdentifier": account,
		"orgIdentifier":     org,
		"projectIdentifier": project,
	}

	if strings.Contains(identifier, ".") {
		identifier = strings.Split(identifier, ".")[1]
	}
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(params).
		SetPathParam("identifier", identifier).
		Get(api.BaseURL + "/ng/api/connectors/{identifier}")
	if err != nil {
		return ConnectorClass{}, err
	}

	connector := Connector{}
	err = json.Unmarshal(resp.Body(), &connector)
	if err != nil {
		return ConnectorClass{}, err
	}
	if len(connector.Data.Connector.Identifier) == 0 {
		return ConnectorClass{}, fmt.Errorf("invalid connector")
	}

	return connector.Data.Connector, nil
}

func (api *APIRequest) GetServices(account, org, project string) ([]*ServiceClass, error) {
	params := map[string]string{
		"accountIdentifier": account,
		"orgIdentifier":     org,
		"projectIdentifier": project,
		"limit":             "1000",
	}

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetHeader("Harness-Account", account).
		SetQueryParams(params).
		SetPathParam("org", org).
		SetPathParam("project", project).
		Get(api.BaseURL + "/v1/orgs/{org}/projects/{project}/services")
	if err != nil {
		return []*ServiceClass{}, err
	}

	service := []*Service{}
	err = json.Unmarshal(resp.Body(), &service)
	if err != nil {
		return []*ServiceClass{}, err
	}

	serviceList := []*ServiceClass{}
	for _, s := range service {
		serviceList = append(serviceList, &s.Service)
	}

	return serviceList, nil
}

func (api *APIRequest) UpdateService(service ServiceRequest, account string) error {
	params := map[string]string{
		"accountIdentifier": account,
	}
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(params).
		SetBody(service).
		Put(api.BaseURL + "/ng/api/servicesV2")
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return err
		}
		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		return fmt.Errorf(errMsg)
	}

	return nil
}

func (api *APIRequest) GetEnvironments(account, org, project string) ([]*EnvironmentClass, error) {
	params := map[string]string{
		"accountIdentifier": account,
	}
	if org != "" {
		params["orgIdentifier"] = org
	}
	if project != "" {
		params["projectIdentifier"] = project
	}

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(params).
		Get(api.BaseURL + "/ng/api/environmentsV2")
	if err != nil {
		return []*EnvironmentClass{}, err
	}

	env := Environment{}
	err = json.Unmarshal(resp.Body(), &env)
	if err != nil {
		return []*EnvironmentClass{}, err
	}

	envList := []*EnvironmentClass{}
	for _, e := range env.Data.Content {
		envList = append(envList, &e.Environment)
	}

	return envList, nil
}

func (api *APIRequest) GetInfrastructures(account, org, project, envId string) ([]*Infrastructure, error) {

	params := map[string]string{
		"accountIdentifier":     account,
		"orgIdentifier":         org,
		"projectIdentifier":     project,
		"environmentIdentifier": envId,
		"limit":                 "1000",
	}

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(params).
		Get(api.BaseURL + "/ng/api/infrastructures")
	if err != nil {
		return []*Infrastructure{}, err
	}

	result := InfraDefResult{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return []*Infrastructure{}, err
	}

	infraList := []*Infrastructure{}
	for _, content := range result.Data.Content {
		infraList = append(infraList, &content.Infrastructure)
	}

	return infraList, nil
}

func (api *APIRequest) UpdateEnvironment(env EnvironmentRequest, account string) error {
	params := map[string]string{
		"accountIdentifier": account,
	}
	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(params).
		SetBody(env).
		Put(api.BaseURL + "/ng/api/environmentsV2/serviceOverrides")
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		ar := ApiResponse{}
		err = json.Unmarshal(resp.Body(), &ar)
		if err != nil {
			return err
		}
		errMsg := fmt.Sprintf("CorrelationId: %s, ResponseMessages: %+v", ar.CorrelationID, ar.ResponseMessages)
		return fmt.Errorf(errMsg)
	}

	return nil
}

func (api *APIRequest) GetServiceOverrides(environment, account, org, project string) ([]*ServiceOverrideContent, error) {
	params := map[string]string{
		"environmentIdentifier": environment,
		"accountIdentifier":     account,
		"orgIdentifier":         org,
		"projectIdentifier":     project,
	}

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(params).
		SetPathParam("org", org).
		SetPathParam("project", project).
		Get(api.BaseURL + "/ng/api/environmentsV2/serviceOverrides")
	if err != nil {
		return []*ServiceOverrideContent{}, err
	}

	overrides := ServiceOverride{}
	err = json.Unmarshal(resp.Body(), &overrides)
	if err != nil {
		return []*ServiceOverrideContent{}, err
	}

	var overrideList []*ServiceOverrideContent
	for _, o := range overrides.Data.Content {
		overrideList = append(overrideList, &o)
	}

	return overrideList, nil
}

func (api *APIRequest) GetOverridesV2(account, org, project string, ovType OverridesV2Type) ([]OverridesV2Content, error) {

	params := map[string]string{
		"accountIdentifier": account,
		"orgIdentifier":     org,
		"projectIdentifier": project,
		"page":              "0",
		"size":              "1000",
		"type":              string(ovType),
	}

	resp, err := api.Client.R().
		SetHeader("x-api-key", api.APIKey).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(params).
		Post(api.BaseURL + "/ng/api/serviceOverrides/v2/list")
	if err != nil {
		return nil, err
	}

	result := OverridesV2Response{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}

	var list []OverridesV2Content
	list = append(list, result.Data.Content...)

	return list, nil
}
