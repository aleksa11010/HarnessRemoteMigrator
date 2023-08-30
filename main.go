package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aleksa11010/HarnessInlineToRemote/harness"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

func main() {

	log := logrus.New()
	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})

	boldCyan := color.New(color.Bold, color.FgBlue)

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
		log.Errorf(color.RedString("Unable to get projects - %s", err))
		return
	}
	log.Infof(color.BlueString("Found total of %d projects", len(projects.Data.Content)))

	log.Infof("Filtering projects based on configuration...")
	var projectList []harness.ProjectsContent
	if len(accountConfig.TargetProjects) > 0 {
		for _, project := range projects.Data.Content {
			skip := true
			for _, include := range accountConfig.TargetProjects {
				if project.Project.Name == include {
					log.Infof(color.BlueString("Project %s is tageted for migration, adding...", project.Project.Name))
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
					log.Infof(color.BlueString("Project %s is excluded from migration, skipping...", project.Project.Name))
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
	fileTmpl := `{{ blue "Downloading files: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
	var pipelines []harness.PipelineContent
	var templates []harness.Templates
	var failedPipelines, failedTemplates []string
	for _, project := range projectList {
		p := project.Project
		log.Infof(boldCyan.Sprintf("---Processing project %s!---", p.Name))
		// Get all pipelines for the project
		log.Infof("Getting pipelines for project %s", p.Name)
		projectPipelines, err := api.GetAllPipelines(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
		if err != nil {
			log.Errorf(color.RedString("Unable to get pipelines - %s", err))
			return
		}
		log.Infof(color.BlueString("Found total of %d pipelines", len(projectPipelines.Data.Content)))

		// Get all templates for the project
		log.Infof("Getting templates for project %s", project.Project.Name)
		projectTemplates, err := api.GetAllTemplates(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
		if err != nil {
			log.Errorf(color.RedString("Unable to get templates - %s", err))
			return
		}
		log.Infof(color.BlueString("Found total of %d templates", len(projectTemplates)))

		if len(projectPipelines.Data.Content) > 0 {
			log.Infof("Moving found pipelines to remote")
			pipelineBar := pb.ProgressBarTemplate(pipelineTmpl).Start(len(projectPipelines.Data.Content))
			for _, pipeline := range projectPipelines.Data.Content {
				// Set the directory to pipelines and use the identifier as file name
				accountConfig.GitDetails.FilePath = "pipelines/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + pipeline.Identifier + ".yaml"
				_, err := pipeline.MovePipelineToRemote(&api, accountConfig, string(p.OrgIdentifier), p.Identifier)
				if err != nil {
					log.Errorf(color.RedString("Unable to move pipeline - %s", pipeline.Name))
					pipelineBar.Increment()
					failedPipelines = append(failedPipelines, pipeline.Name)
				} else {
					pipelineBar.Increment()
				}
			}
			pipelineBar.Finish()
		}
		pipelines = append(pipelines, projectPipelines.Data.Content...)

		if len(projectTemplates) > 0 {
			log.Infof("Moving found templates to remote")
			templateBar := pb.ProgressBarTemplate(templateTmpl).Start(len(projectPipelines.Data.Content))
			for _, template := range projectTemplates {
				// Set the directory to templates and use the identifier as file name
				accountConfig.GitDetails.FilePath = "templates/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + template.Identifier + ".yaml"
				_, err := template.MoveTemplateToRemote(&api, accountConfig, string(p.OrgIdentifier), p.Identifier)
				if err != nil {
					log.Errorf(color.RedString("Unable to move template - %s", template.Name))
					templateBar.Increment()
					failedTemplates = append(failedTemplates, template.Name)
				} else {
					templateBar.Increment()
				}
			}
			templateBar.Finish()
		}
		templates = append(templates, projectTemplates)
	}
	log.Infof(boldCyan.Sprintf("---Pipelines---"))
	if len(failedPipelines) > 0 {
		log.Warnf(color.HiYellowString("These pipelines (count:%d) failed while moving to remote: \n%s", len(failedPipelines), strings.Join(failedPipelines, ",\n")))
	}
	log.Infof(color.GreenString("Processed total of %d pipelines", len(pipelines)))
	log.Infof(boldCyan.Sprintf("---Templates---"))
	if len(failedTemplates) > 0 {
		log.Warnf(color.HiYellowString("These templates (count:%d) failed while moving to remote: \n%s", len(failedTemplates), strings.Join(failedTemplates, ",\n")))
	}
	log.Infof(color.GreenString("Processed total of %d templates", len(templates)))

	log.Infof(color.GreenString("------"))
	log.Infof(color.GreenString("Moved Pipelines and Templates to remote!"))
	log.Infof(color.GreenString("------"))

	var failedFiles, failedOrgFiles, failedProjectFiles []string
	log.Infof("Getting file store for account %s", accountConfig.AccountIdentifier)
	accountFiles, err := api.GetAllAccountFiles(accountConfig.AccountIdentifier)
	if err != nil {
		log.Errorf(color.RedString("Unable to get file store at account level - %s", err))
		return
	}

	log.Infof("Downloading %d files from Account level", len(accountFiles))
	accountFileBar := pb.ProgressBarTemplate(fileTmpl).Start(len(accountFiles))
	for _, file := range accountFiles {
		err := file.DownloadFile(&api, accountConfig.AccountIdentifier, "", "", "account")
		if err != nil {
			log.Errorf(color.RedString("Unable to download file - %s", err))
			failedFiles = append(failedFiles, file.Name)
		}
		accountFileBar.Increment()
	}
	accountFileBar.Finish()

	log.Info("Getting file store for organizations")
	orgs, err := api.GetAllOrgs(accountConfig.AccountIdentifier)
	if err != nil {
		log.Errorf(color.RedString("Unable to get organizations for account %s - %s", accountConfig.AccountIdentifier, err))
		return
	}

	var allOrgFiles []harness.FileStoreContent
	for _, org := range orgs {
		o := org.Org
		orgFiles, err := api.GetAllOrgFiles(accountConfig.AccountIdentifier, o.Identifier)
		if err != nil {
			log.Errorf(color.RedString("Unable to get file store for org %s - %s", o.Name, err))
		}
		if len(orgFiles) > 0 {
			log.Infof("Downloading %d files from Organization %s", len(orgFiles), o.Name)
			orgFileBar := pb.ProgressBarTemplate(fileTmpl).Start(len(accountFiles))
			for _, file := range orgFiles {
				err := file.DownloadFile(&api, accountConfig.AccountIdentifier, o.Identifier, "", "org/"+o.Identifier)
				if err != nil {
					log.Errorf(color.RedString("Unable to download file - %s", err))
					failedOrgFiles = append(failedOrgFiles, file.Name)
				}
				orgFileBar.Increment()
			}
			orgFileBar.Finish()
		}
		allOrgFiles = append(allOrgFiles, orgFiles...)
	}

	var allProjectFiles []harness.FileStoreContent
	log.Info("Getting file store for projects")
	for _, project := range projectList {
		p := project.Project
		projectFiles, err := api.GetAllProjectFiles(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
		if err != nil {
			log.Errorf(color.RedString("Unable to get file store for project %s - %s", p.Name, err))
		}
		if len(projectFiles) > 0 {
			log.Infof("Downloading %d files from project %s", len(projectFiles), p.Name)
			projectBar := pb.ProgressBarTemplate(fileTmpl).Start(len(accountFiles))
			for _, file := range projectFiles {
				err := file.DownloadFile(&api, accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier, fmt.Sprintf("org/%s/%s", p.OrgIdentifier, p.Identifier))
				if err != nil {
					log.Errorf(color.RedString("Unable to download file - %s", err))
					failedProjectFiles = append(failedProjectFiles, file.Name)
				}
				projectBar.Increment()
			}
			projectBar.Finish()
		}
		allProjectFiles = append(allProjectFiles, projectFiles...)
	}
	log.Infof(boldCyan.Sprintf("---File Store---"))
	log.Infof(color.GreenString("Processed total of %d files at Account level", len(accountFiles)))
	if len(failedFiles) > 0 {
		log.Warnf(color.HiYellowString("These files (count:%d) failed while downloading from account level: \n%s", len(failedFiles), strings.Join(failedFiles, ",\n")))
	}
	log.Infof(color.GreenString("Processed total of %d files at Organization level", len(allOrgFiles)))
	if len(failedOrgFiles) > 0 {
		log.Warnf(color.HiYellowString("These files (count:%d) failed while downloading from account level: \n%s", len(failedOrgFiles), strings.Join(failedOrgFiles, ",\n")))
	}
	log.Infof(color.GreenString("Processed total of %d files at Project level", len(allProjectFiles)))
	if len(failedProjectFiles) > 0 {
		log.Warnf(color.HiYellowString("These files (count:%d) failed while downloading from account level: \n%s", len(failedProjectFiles), strings.Join(failedProjectFiles, ",\n")))
	}

	log.Infof(boldCyan.Sprintf("---Creating Git Repo---"))
	// Init empty repo inside the filestore directory
	cmd := exec.Command("git", "init")
	cmd.Dir = "./filestore"
	err = cmd.Run()
	if err != nil {
		log.Errorf(color.RedString("Unable to init git repo - %s", err))
	}

	log.Infof(color.GreenString("Git repo initialized"))
	// Add files to git repo
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = "./filestore"
	err = cmd.Run()
	if err != nil {
		log.Errorf(color.RedString("Unable to add files to git repo - %s", err))
		return
	}
	log.Info(color.GreenString("Files added to git repo"))

	// Commit files to git repo
	cmd = exec.Command("git", "commit", "-m", "Initial Filestore commit")
	cmd.Dir = "./filestore"
	err = cmd.Run()
	if err != nil {
		log.Errorf(color.RedString("Unable to commit files to git repo - %s", err))
		return
	}
	log.Info(color.GreenString("Files committed to git repo"))

	// Set remote url to git repo
	var url string
	if accountConfig.FileStoreConfig.RepositoryURL != "" {
		url = accountConfig.FileStoreConfig.RepositoryURL
		if !strings.Contains(url, ".git") {
			url += ".git"
		}
	} else {
		var err error
		conn, err := api.GetConnector(
			accountConfig.AccountIdentifier,
			accountConfig.FileStoreConfig.Organization,
			accountConfig.FileStoreConfig.Project,
			accountConfig.GitDetails.ConnectorRef,
		)
		if err != nil {
			log.Errorf(color.RedString("Unable to get connector - %s", err))
			return
		}
		url = conn.Spec.URL + ".git"
	}

	cmd = exec.Command("git", "remote", "add", "origin", url)
	cmd.Dir = "./filestore"
	err = cmd.Run()
	if err != nil {
		log.Errorf(color.RedString("Unable to commit files to git repo - %s", err))
		return
	}
	log.Info(color.GreenString("Remote url set to git repo"))

	// Push files to git repo
	var branch string
	if accountConfig.FileStoreConfig.Branch != "" {
		branch = accountConfig.FileStoreConfig.Branch
	} else {
		log.Errorf(color.RedString("File Store branch is not set"))
		return
	}

	// Check if branch exists
	cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = "./filestore"
	err = cmd.Run()
	if err != nil {
		log.Warnf(color.YellowString("Branch %s does not exist", branch))
		log.Infof("Creating branch %s", branch)

		// Create new branch
		cmd = exec.Command("git", "checkout", "-b", branch)
		cmd.Dir = "./filestore"
		err = cmd.Run()
		if err != nil {
			log.Errorf(color.RedString("Unable to create branch %s - %s", branch, err))
			return
		}
	}
	log.Infof("Branch %s exists", branch)

	// Push files to git repo
	cmd = exec.Command("git", "push", "origin", branch)
	cmd.Dir = "./filestore"
	err = cmd.Run()
	if err != nil {
		log.Errorf(color.RedString("Unable to push files to git repo - %s", err))
		return
	}
	log.Info(color.GreenString("Files pushed to git repo!"))

}
