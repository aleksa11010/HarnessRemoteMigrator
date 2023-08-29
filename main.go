package main

import (
	"flag"
	"strings"

	"github.com/aleksa11010/HarnessInlineToRemote/harness"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/cheggaaa/pb/v3"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

func main() {

	log := logrus.New()
	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})

	accountArg := flag.String("account", "", "Provide your account ID.")
	apiKeyArg := flag.String("api-key", "", "Provide your API Key.")
	configFile := flag.String("config", "", "Provide a config file.")
	gitConnectorRef := flag.String("git-connector-ref", "", "Provide a git connector ref.")
	gitRepoName := flag.String("git-repo-name", "", "Provide a git repo name.")
	excludeProjects := flag.String("exclude-projects", "", "Provide a list of projects to exclude.")
	targetProjects := flag.String("target-projects", "", "Provide a list of projects to target.")

	flag.Parse()

	accountConfig := harness.Config{}

	if *configFile != "" {
		accountConfig.ReadConfig(*configFile)
	} else {
		accountConfig = harness.Config{
			AccountIdentifier: *accountArg,
			ApiKey:            *apiKeyArg,
			GitDetails: harness.GitDetails{
				BranchName:    "migration",
				CommitMessage: "Migrating piplines from inline to remote",
				// BaseBranch:    "main",
				ConnectorRef: *gitConnectorRef,
				RepoName:     *gitRepoName,
			},
			ExcludeProjects: strings.Split(*excludeProjects, ","),
			TargetProjects:  strings.Split(*targetProjects, ","),
		}
	}

	api := harness.APIRequest{
		BaseURL: harness.BaseURL,
		Client:  resty.New(),
		APIKey:  accountConfig.ApiKey,
	}

	log.Infof("Getting projects for account %s", accountConfig.AccountIdentifier)
	projects, err := api.GetAllProjects(accountConfig.AccountIdentifier)
	if err != nil {
		log.Errorf("Unable to get projects - %s", err)
		return
	}
	log.Infof("Found total of %d projects", len(projects.Data.Content))

	log.Infof("Filtering projects based on configuration...")
	var projectList []harness.ProjectsContent
	if len(accountConfig.TargetProjects) > 0 {
		for _, project := range projects.Data.Content {
			skip := true
			for _, include := range accountConfig.TargetProjects {
				if project.Project.Name == include {
					log.Infof("Project %s is tageted for migration, adding...", project.Project.Name)
					skip = false
					break
				}
			}

			if skip {
				continue
			}
			projectList = append(projectList, project)
		}
	} else if len(accountConfig.ExcludeProjects) > 0 {
		for _, project := range projects.Data.Content {
			skip := false
			for _, exclude := range accountConfig.ExcludeProjects {
				if project.Project.Name == exclude {
					log.Infof("Project %s is excluded from migration, skipping...", project.Project.Name)
					skip = true
					break
				}
			}

			if skip {
				continue
			}
			projectList = append(projectList, project)
		}
	}

	log.Infof("Processing total of %d projects", len(projectList))
	pipelineTmpl := `{{ blue "Processing Pipelines: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
	templateTmpl := `{{ blue "Processing Templates: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
	var pipelines []harness.PipelineContent
	var failedPipelines []string
	for _, project := range projectList {
		log.Infof("---Processing project %s!---", project.Project.Name)
		// Get all pipelines for the project
		log.Infof("Getting pipelines for project %s", project.Project.Name)
		projectPipelines, err := api.GetAllPipelines(accountConfig.AccountIdentifier, string(project.Project.OrgIdentifier), project.Project.Identifier)
		if err != nil {
			log.Errorf("Unable to get pipelines - %s", err)
			return
		}
		log.Infof("Found total of %d pipelines", len(projectPipelines.Data.Content))

		// Get all templates for the project
		log.Infof("Getting templates for project %s", project.Project.Name)
		projectTemplates, err := api.GetAllTemplates(accountConfig.AccountIdentifier, string(project.Project.OrgIdentifier), project.Project.Identifier)
		if err != nil {
			log.Errorf("Unable to get templates - %s", err)
			return
		}
		log.Infof("Found total of %d templates", len(projectTemplates))

		log.Infof("Moving found pipelines to remote")
		pipelineBar := pb.ProgressBarTemplate(pipelineTmpl).Start(len(projectPipelines.Data.Content))
		for _, pipeline := range projectPipelines.Data.Content {
			accountConfig.GitDetails.FilePath = "pipelines/" + pipeline.Identifier + ".yaml"
			resp, err := pipeline.MovePipelineToRemote(&api, accountConfig, string(project.Project.OrgIdentifier), project.Project.Identifier)
			if err != nil {
				log.Errorf("Unable to move pipeline - %s", pipeline.Name)
				pipelineBar.Increment()
				failedPipelines = append(failedPipelines, pipeline.Name)
			} else {
				log.Infof("Moved pipeline %s to remote, response %s", pipeline.Name, string(resp))
				pipelineBar.Increment()
			}
		}
		pipelineBar.Finish()
		pipelines = append(pipelines, projectPipelines.Data.Content...)

		log.Infof("Moving found templates to remote")
		templateBar := pb.ProgressBarTemplate(templateTmpl).Start(len(projectPipelines.Data.Content))
		for _, template := range projectTemplates {
			accountConfig.GitDetails.FilePath = "templates/" + template.Identifier + ".yaml"
			resp, err := template.MoveTemplateToRemote(&api, accountConfig, string(project.Project.OrgIdentifier), project.Project.Identifier)
			if err != nil {
				log.Errorf("Unable to move template - %s", template.Name)
				templateBar.Increment()
				failedPipelines = append(failedPipelines, template.Name)
			} else {
				log.Infof("Moved template %s to remote, response %s", template.Name, string(resp))
				templateBar.Increment()
			}
		}
		templateBar.Finish()
		pipelines = append(pipelines, projectPipelines.Data.Content...)
	}

	log.Warnf("These pipelines (count:%d) failed while moving to remote: \n%s", len(failedPipelines), strings.Join(failedPipelines, ",\n"))
	log.Infof("Processed total of %d pipelines", len(pipelines))
}
