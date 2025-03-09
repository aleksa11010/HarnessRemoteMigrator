package main

import (
	"bytes"
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
	"gopkg.in/yaml.v2"
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
	allFlag := flag.Bool("all", false, "Migrate all entities.")
	pipelinesFlag := flag.Bool("pipelines", false, "Migrate pipelines.")
	inputsetsFlag := flag.Bool("inputsets", false, "Migrate inputsets.")
	templatesFlag := flag.Bool("templates", false, "Migrate templates.")
	servicesFlag := flag.Bool("services", false, "Migrate services")
	envFlag := flag.Bool("environments", false, "Migrate environments")
	infraDefFlag := flag.Bool("infraDef", false, "Migrate infrastructure definition")
	filestoreFlag := flag.Bool("filestore", false, "Migrate filestore.")
	serviceManifests := flag.Bool("service", false, "Migrate service manifests.")
	forceServiceUpdate := flag.Bool("update-service", false, "Force update remote service manifests")
	overridesFlag := flag.Bool("overrides", false, "Migrate service overrides")
	overridesV2Flag := flag.Bool("overrides-v2", false, "Migrate service overrides V2")
	urlEncoding := flag.Bool("url-encode-string", false, "Encode Paths as URL friendly strings")
	cgFolderStructure := flag.Bool("alt-path", false, "CG-like folder structure for Git")
	prod3 := flag.Bool("prod3", false, "User Prod3 base URL for API calls")
	customGitDetailsFilePath := flag.String("custom-remote-path", "", "A custom file path where to save remote manifests.")
	gitX := flag.Bool("gitx", false, "Migrate entity following the Git Experience definitions")

	flag.Parse()

	type MigrationScope struct {
		Pipelines            bool
		Inputsets            bool
		Templates            bool
		Services             bool
		Environments         bool
		InfraDef             bool
		FileStore            bool
		ServiceManifests     bool
		ForceUpdateManifests bool
		Overrides            bool
		OverridesV2          bool
		UrlEncoding          bool
		CGFolderStructure    bool
		Prod3                bool
		GitX                 bool
	}
	scope := MigrationScope{}

	if *allFlag {
		scope = MigrationScope{
			Pipelines:            true,
			Inputsets:            true,
			Templates:            true,
			Services:             true,
			Environments:         true,
			InfraDef:             true,
			FileStore:            true,
			ServiceManifests:     true,
			ForceUpdateManifests: *forceServiceUpdate,
			Overrides:            true,
			OverridesV2:          true,
			UrlEncoding:          *urlEncoding,
			CGFolderStructure:    false,
			Prod3:                false,
			GitX:                 *gitX,
		}
	} else {
		scope = MigrationScope{
			Pipelines:            *pipelinesFlag,
			Inputsets:            *inputsetsFlag,
			Templates:            *templatesFlag,
			Services:             *servicesFlag,
			Environments:         *envFlag,
			InfraDef:             *infraDefFlag,
			FileStore:            *filestoreFlag,
			ServiceManifests:     *serviceManifests,
			ForceUpdateManifests: *forceServiceUpdate,
			Overrides:            *overridesFlag,
			OverridesV2:          *overridesV2Flag,
			UrlEncoding:          *urlEncoding,
			CGFolderStructure:    *cgFolderStructure,
			Prod3:                *prod3,
			GitX:                 *gitX,
		}
	}

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

	var baseUrl string
	if scope.Prod3 {
		baseUrl = harness.BaseURLProd3
	} else {
		baseUrl = harness.BaseURL
	}
	api := harness.APIRequest{
		BaseURL: baseUrl,
		Client:  resty.New(),
		APIKey:  accountConfig.ApiKey,
	}

	if !scope.Pipelines && !scope.Templates && !scope.FileStore && !scope.Overrides && !scope.Services && !scope.Environments && !scope.InfraDef && !scope.Inputsets && !scope.OverridesV2 {
		log.Errorf(color.RedString("You need to specify at least one type of entity to migrate!"))
		log.Errorf(color.RedString("Please use -pipelines, -templates, -services, -environments, -overrides-v2, -filestore or -overrides flags"))
		log.Errorf(color.RedString("If you want to migrate all entities, use -all flag"))
		return
	}

	if (scope.ServiceManifests || scope.Overrides) && !scope.FileStore && !*allFlag {
		log.Errorf(color.RedString("In order to migrate Service Manifests and/or Service Overrides you need to use FileStore flag!"))
		log.Errorf(color.RedString("Please use -filestore flag followed by entities you want to migrate, -service or -overrides"))
		log.Errorf(color.RedString("If you want to migrate all entities, use -all flag"))
		return
	}

	log.Infof("Getting projects for account %s", accountConfig.AccountIdentifier)
	projects, err := api.GetAllProjects(accountConfig.AccountIdentifier)
	if err != nil {
		log.Errorf(color.RedString("Unable to get projects - %s", err))
		return
	}
	log.Infof(color.BlueString("Found total of %d projects", len(projects.Data.Content)))
	if len(projects.Data.Content) == 0 {
		log.Errorf(color.RedString("Did not find any projects!"))
		log.Errorf(color.RedString("Please check your token and/or configuration file."))
		return
	}

	log.Infof("Filtering projects based on configuration...")
	var projectList []harness.ProjectsContent
	if len(accountConfig.TargetProjects) > 0 {
		for _, project := range projects.Data.Content {
			skip := true
			for _, include := range accountConfig.TargetProjects {
				if project.Project.Name == include || project.Project.Identifier == include {
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
				if project.Project.Name == exclude || project.Project.Identifier == exclude {
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
	} else {
		projectList = append(projectList, projects.Data.Content...)
	}

	log.Infof("Processing total of %d projects", len(projectList))
	pipelineTmpl := `{{ blue "Processing Pipelines: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
	serviceTmpl := `{{ blue "Processing Services: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
	templateTmpl := `{{ blue "Processing Templates: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
	servicesTmpl := `{{ blue "Processing Services: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
	envTmpl := `{{ blue "Processing Environments: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
	fileTmpl := `{{ blue "Downloading files: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
	overridesTmpl := `{{ blue "Downloading files: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `

	var pipelines []harness.PipelineContent
	var templates []harness.Template
	var services []*harness.ServiceClass
	var environments []*harness.EnvironmentClass
	var failedPipelines, failedTemplates, failedServices, failedEnvs, alreadyRemoteServices []string
	for _, project := range projectList {
		p := project.Project
		log.Infof(boldCyan.Sprintf("---Processing project %s!---", p.Name))
		// Get all pipelines for the project
		if scope.Pipelines {
			log.Infof("Getting pipelines for project %s", p.Name)
			projectPipelines, err := api.GetAllPipelines(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
			if err != nil {
				log.Errorf(color.RedString("Unable to get pipelines - %s", err))
				return
			}
			log.Infof(color.BlueString("Found total of %d pipelines", len(projectPipelines.Data.Content)))

			if len(projectPipelines.Data.Content) > 0 {
				log.Infof("Moving found pipelines to remote")
				pipelineBar := pb.ProgressBarTemplate(pipelineTmpl).Start(len(projectPipelines.Data.Content))
				for _, pipeline := range projectPipelines.Data.Content {
					// Set the directory to pipelines and use the identifier as file name
					if scope.UrlEncoding {
						accountConfig.GitDetails.FilePath = "pipelines%2F" + string(p.OrgIdentifier) + "%2F" + p.Identifier + "%2F" + pipeline.Identifier + ".yaml"
					} else {
						if scope.CGFolderStructure {
							accountConfig.GitDetails.FilePath = "account/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/pipelines/" + pipeline.Identifier + ".yaml"
						} else {
							accountConfig.GitDetails.FilePath = harness.GetPipelineFilePath(scope.GitX, *customGitDetailsFilePath, p, pipeline)
						}
					}
					_, err := pipeline.MovePipelineToRemote(&api, accountConfig, string(p.OrgIdentifier), p.Identifier)
					if err != nil {
						log.Errorf(color.RedString("Unable to move pipeline - %s", pipeline.Name))
						log.Errorf(color.RedString(err.Error()))
						failedPipelines = append(failedPipelines, pipeline.Name)
					}
					pipelineBar.Increment()
				}
				pipelineBar.Finish()
			}
			pipelines = append(pipelines, projectPipelines.Data.Content...)
		}
		if scope.Inputsets {
			log.Infof("Getting inputsets for project %s", p.Name)
			projectPipelines, err := api.GetAllPipelines(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
			if err != nil {
				log.Errorf(color.RedString("Unable to get pipelines for inputsets - %s", err))
				return
			}

			if len(projectPipelines.Data.Content) > 0 {
				for _, pipeline := range projectPipelines.Data.Content {
					if pipeline.StoreType != "REMOTE" {
						continue
					}

					inputsets, err := api.GetInputsets(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier, pipeline.Identifier)
					if err != nil {
						log.Errorf(color.RedString("Unable to list inputsets from pipeline - %s", pipeline.Name))
						continue
					}

					for _, is := range inputsets {
						accountConfig.GitDetails.FilePath = harness.GetInputsetFilePath(scope.GitX, *customGitDetailsFilePath, p, is)

						err := is.MoveInputsetToRemote(&api, accountConfig, p.Identifier, string(p.OrgIdentifier))
						if err != nil {
							log.Errorf(color.RedString("Unable to move inputsets [%s] for pipeline - %s", is.Name, pipeline.Name))
							log.Errorf(color.RedString(err.Error()))

						}
					}
				}
			}
		}

		if scope.Templates {
			// Get all templates for the project
			log.Infof("Getting templates for project %s", project.Project.Name)
			projectTemplates, err := api.GetAllTemplates(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
			if err != nil {
				log.Errorf(color.RedString("Unable to get templates - %s", err))
				return
			}
			log.Infof(color.BlueString("Found total of %d templates", len(projectTemplates)))
			if len(projectTemplates) > 0 {
				log.Infof("Moving found templates to remote")
				templateBar := pb.ProgressBarTemplate(templateTmpl).Start(len(projectTemplates))
				for _, template := range projectTemplates {
					// Set the directory to templates and use the identifier as file name
					if scope.UrlEncoding {
						accountConfig.GitDetails.FilePath = "templates%2f" + string(p.OrgIdentifier) + "%2F" + p.Identifier + "%2F" + template.Identifier + "-" + template.VersionLabel + ".yaml"
						template.GitDetails = accountConfig.GitDetails
					} else {
						if scope.CGFolderStructure {
							accountConfig.GitDetails.FilePath = "account/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/templates/" + template.Identifier + "-" + template.VersionLabel + ".yaml"
							template.GitDetails = accountConfig.GitDetails
						} else {
							accountConfig.GitDetails.FilePath = harness.GetTemplateFilePath(scope.GitX, *customGitDetailsFilePath, p, template)
							template.GitDetails = accountConfig.GitDetails
						}
					}
					if template.StoreType == "REMOTE" {
						log.Infof("Template [%s] Version [%s] is already remote!", template.Identifier, template.VersionLabel)
					} else {
						_, err := template.MoveTemplateToRemote(&api, accountConfig)
						if err != nil {
							log.Errorf(color.RedString("Unable to move template - %s", template.Name))
							log.Errorf(color.RedString(err.Error()))
							failedTemplates = append(failedTemplates, template.Name)
						}
					}
					templateBar.Increment()
				}
				templateBar.Finish()
			}
			templates = append(templates, projectTemplates...)
		}

		// SERVICES MOVE TO REMOTE
		if scope.Services {
			log.Infof("Getting services for project %s", project.Project.Name)
			projectServices, err := api.GetServices(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
			if err != nil {
				log.Errorf(color.RedString("Unable to get services - %s", err))
				return
			}

			log.Infof(color.BlueString("Found total of %d services", len(projectServices)))
			if len(projectServices) > 0 {
				log.Infof("Moving found services to remote")
				servicesBar := pb.ProgressBarTemplate(servicesTmpl).Start(len(projectServices))

				for _, service := range projectServices {
					accountConfig.GitDetails.FilePath = harness.GetServiceFilePath(scope.GitX, *customGitDetailsFilePath, p, *service)

					if service.StoreType == "REMOTE" {
						log.Infof("Service [%s] is already remote", service.Identifier)
					} else {
						_, alreadyRemote, err := service.MoveServiceToRemote(&api, accountConfig)
						if err != nil {
							log.Errorf(color.RedString("Unable to move service - %s", service.Name))
							log.Errorf(color.RedString(err.Error()))
							failedServices = append(failedServices, service.Name)
						}
						if alreadyRemote {
							alreadyRemoteServices = append(alreadyRemoteServices, service.Name)
						}
					}
					servicesBar.Increment()
				}
				servicesBar.Finish()
			}
			services = append(services, projectServices...)
		}

		// ENVIRONMENTS MOVE TO REMOTE
		if scope.Environments {
			log.Infof("Getting environments for project %s", project.Project.Name)
			projectEnvironments, err := api.GetEnvironments(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
			if err != nil {
				log.Errorf(color.RedString("Unable to get environments - %s", err))
				return
			}

			log.Infof(color.BlueString("Found total of %d environments", len(projectEnvironments)))
			if len(projectEnvironments) > 0 {
				log.Infof("Moving found environments to remote")
				envBar := pb.ProgressBarTemplate(envTmpl).Start(len(projectEnvironments))

				for _, environment := range projectEnvironments {
					accountConfig.GitDetails.FilePath = harness.GetEnvironmentFilePath(scope.GitX, *customGitDetailsFilePath, p, *environment)

					if environment.StoreType == "REMOTE" {
						log.Infof("Environment [%s] is already remote", environment.Identifier)
					} else {
						err = environment.MoveEnvironmentToRemote(&api, accountConfig)
						if err != nil {
							log.Errorf(color.RedString("Unable to move environment - %s", environment.Name))
							log.Errorf(color.RedString(err.Error()))
							failedEnvs = append(failedEnvs, environment.Name)
						}
					}
					envBar.Increment()
				}
				envBar.Finish()
			}
			environments = append(environments, projectEnvironments...)
		}

		// INFRA-DEF MOVE TO REMOTE
		if scope.InfraDef {
			err := processInfraDefScope(log, api, *customGitDetailsFilePath, accountConfig, p, scope.GitX)
			if err != nil {
				log.Errorf(color.RedString("Unable to inline-to-remote infrastructure - %s", err))
			}
		}

		// OVERRIDES-V2 MOVE TO REMOTE
		if scope.OverridesV2 {
			err := processOverridesV2(log, api, *customGitDetailsFilePath, accountConfig, p, scope.GitX)
			if err != nil {
				log.Errorf(color.RedString("Unable to inline-to-remote overrides v2 - %s", err))
			}
		}
	}
	if scope.Pipelines {
		pipelinesSummary(log, boldCyan, failedPipelines, pipelines)
	}
	if scope.Templates {
		templatesSummary(log, boldCyan, failedTemplates, templates)
	}
	if scope.Services {
		servicesSummary(log, boldCyan, failedTemplates, failedServices, alreadyRemoteServices, services)
	}
	if scope.Environments {
		environmentsSummary(log, boldCyan, failedEnvs, environments)
	}
	if scope.FileStore {
		var failedFiles, failedOrgFiles, failedProjectFiles, failedServices []string
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
					err := file.DownloadFile(&api, accountConfig.AccountIdentifier, o.Identifier, "", "/"+o.Identifier)
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
				projectBar := pb.ProgressBarTemplate(fileTmpl).Start(len(projectFiles))
				for _, file := range projectFiles {
					err := file.DownloadFile(&api, accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier, fmt.Sprintf("/%s/%s", p.OrgIdentifier, p.Identifier))
					if err != nil {
						log.Errorf(color.RedString("Unable to download file [%s] with identifier [%s] - %s", file.Name, file.Identifier, err))
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
			log.Warnf(color.HiYellowString("These files (count:%d) failed while downloading: \n%s", len(failedProjectFiles), strings.Join(failedProjectFiles, ",\n")))
		}

		log.Infof(boldCyan.Sprintf("---Creating Git Repo---"))
		var stderr bytes.Buffer
		// Init empty repo inside the filestore directory
		cmd := exec.Command("git", "init")
		cmd.Dir = "./filestore"
		cmd.Stderr, cmd.Stdout = &stderr, &stderr
		err = cmd.Run()
		if err != nil {
			log.Errorf(color.RedString("Unable to init git repo - %s", err))
		}

		// Set pull default to merge
		cmd = exec.Command("git", "config", "pull.rebase", "false")
		cmd.Dir = "./filestore"
		cmd.Stderr, cmd.Stdout = &stderr, &stderr
		err = cmd.Run()
		if err != nil {
			errorMessage := stderr.String()
			log.Errorf(color.RedString("Unable to set git pull.rebase to false - Git Operations log:\n %s", errorMessage))
		}

		log.Infof(color.GreenString("Git repo initialized"))
		// Add files to git repo
		cmd = exec.Command("git", "add", ".")
		cmd.Dir = "./filestore"
		cmd.Stderr, cmd.Stdout = &stderr, &stderr
		err = cmd.Run()
		if err != nil {
			log.Errorf(color.RedString("Unable to add files to git repo - Git Operations log:\n %s", err))
			return
		}
		log.Info(color.GreenString("Files added to git repo"))

		// Commit files to git repo
		cmd = exec.Command("git", "commit", "-m", "Initial Filestore commit")
		cmd.Dir = "./filestore"
		cmd.Stderr, cmd.Stdout = &stderr, &stderr
		err = cmd.Run()
		if err != nil {
			errorMessage := stderr.String()
			if !strings.Contains(errorMessage, "nothing to commit") {
				log.Errorf("Unable to commit files to git repo - Git Operations log:\n %s", errorMessage)
				return
			}
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
		cmd.Stderr, cmd.Stdout = &stderr, &stderr
		err = cmd.Run()
		if err != nil {
			errorMessage := stderr.String()
			if !strings.Contains(errorMessage, "remote origin already exists.") {
				log.Errorf("Unable to add remote origin to git repo - Git Operations log:\n %s", errorMessage)
				return
			}
		}
		log.Info(color.GreenString("Remote url set to git repo"))

		// Push files to git repo
		var branch string
		if accountConfig.FileStoreConfig.Branch != "" {
			branch = accountConfig.FileStoreConfig.Branch
		} else {
			log.Error(color.RedString("File Store branch is not set"))
			return
		}

		// Check if branch exists
		cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
		cmd.Dir = "./filestore"
		cmd.Stderr, cmd.Stdout = &stderr, &stderr
		err = cmd.Run()
		if err != nil {
			log.Warnf(color.YellowString("Branch %s does not exist", branch))
			log.Infof("Creating branch %s", branch)

			// Create new branch
			cmd = exec.Command("git", "checkout", "-b", branch)
			cmd.Dir = "./filestore"
			cmd.Stderr, cmd.Stdout = &stderr, &stderr

			err = cmd.Run()
			if err != nil {
				log.Errorf(color.RedString("Unable to create branch %s - %s", branch, err))
				return
			}
		}
		log.Infof("Branch %s exists", branch)

		cmd = exec.Command("git", "pull", "origin", branch, "--allow-unrelated-histories", "--no-ff")
		cmd.Dir = "./filestore"
		cmd.Stderr, cmd.Stdout = &stderr, &stderr
		err = cmd.Run()
		if err != nil {
			errorMessage := stderr.String()
			if !strings.Contains(errorMessage, "couldn't find remote ref") {
				log.Errorf("Unable to pull from remote repo - Git Operations log:\n %s", errorMessage)
				return
			}
		}

		// Push files to git repo
		cmd = exec.Command("git", "push", "origin", branch)
		cmd.Dir = "./filestore"
		cmd.Stderr, cmd.Stdout = &stderr, &stderr
		err = cmd.Run()
		if err != nil {
			log.Errorf(color.RedString("Unable to push files to git repo - Git Operations log:\n %s", err))
			return
		}
		log.Info(color.GreenString("Files pushed to git repo!"))

		if scope.ServiceManifests {
			var targetServices, excludeServices []map[string]string
			targetServices = accountConfig.TargetServices
			excludeServices = accountConfig.ExcludeServices

			log.Infof(boldCyan.Sprintf("---Getting Connector Info---"))
			conn, err := api.GetConnector(
				accountConfig.AccountIdentifier,
				accountConfig.FileStoreConfig.Organization,
				accountConfig.FileStoreConfig.Project,
				accountConfig.GitDetails.ConnectorRef,
			)
			if err != nil {
				log.Errorf("Unable to get Connector info - %s", err)
				return
			}
			log.Infof(boldCyan.Sprintf("---Getting Service Info---"))
			var serviceList []*harness.ServiceClass
			for _, project := range projectList {
				p := project.Project
				log.Infof(boldCyan.Sprintf("---Processing project %s!---", p.Name))
				service, err := api.GetServices(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
				if err != nil {
					log.Errorf(color.RedString("Unable to get service - %s", err))
				}
				if len(targetServices) > 0 {
					var targeted []*harness.ServiceClass
					for _, s := range service {
						for _, t := range targetServices {
							if value, exists := t[s.Identifier]; exists && value == s.Project {
								log.Infof("Service [%s] in project [%s] is targeted for migration!", s.Name, s.Project)
								targeted = append(targeted, s)
							}
						}
					}
					serviceList = append(serviceList, targeted...)
				} else if len(excludeServices) > 0 {
					var targeted []*harness.ServiceClass
					for _, s := range service {
						for _, t := range targetServices {
							if value, exists := t[s.Identifier]; exists && value == s.Project {
								log.Infof("Service [%s] in project [%s] is targeted for exclusion, skipping!", s.Name, s.Project)
								continue
							} else {
								targeted = append(targeted, s)
							}
						}
					}
					serviceList = append(serviceList, targeted...)
				} else {
					serviceList = append(serviceList, service...)
				}
			}
			log.Infof(color.BlueString("Found total of %d services", len(serviceList)))
			log.Infof(boldCyan.Sprintf("---Processing Services---"))
			serviceBar := pb.ProgressBarTemplate(serviceTmpl).Start(len(serviceList))
			for _, service := range serviceList {
				serviceYaml, err := service.ParseYAML()
				if err != nil {
					log.Errorf(color.RedString("Unable to parse service YAML - %s", err))
				}
				var update = false
				for i := range serviceYaml.Service.ServiceDefinition.Spec.Manifests {
					m := &serviceYaml.Service.ServiceDefinition.Spec.Manifests[i]
					if m.Manifest.Spec.Store.Type == "Harness" {
						m.Manifest.Spec.Store.Type = harness.GetServiceManifestStoreType(conn.Type)
						var files []string
						for _, file := range m.Manifest.Spec.Store.Spec.Files {
							files = append(files, fmt.Sprintf("filestore/%s/%s%s", service.Org, service.Project, file))
						}
						var valueFiles []string
						if len(m.Manifest.Spec.ValuesPaths) > 0 {
							log.Infof("Setting values file paths")
							for _, v := range m.Manifest.Spec.ValuesPaths {
								valueFiles = append(valueFiles, fmt.Sprintf("filestore/%s/%s%s", service.Org, service.Project, v))
							}
							log.Infof("Setting following value file paths : %+v", valueFiles)
						}
						log.Infof("Setting following file paths : %+v", files)
						if m.Manifest.Spec.Store.Type == "GitLab" || m.Manifest.Spec.Store.Type == "Github" {
							m.Manifest.Spec.Store.Spec.Paths = files
						} else {
							m.Manifest.Spec.Store.Spec.Files = files
						}
						m.Manifest.Spec.Store.Spec.Branch = accountConfig.GitDetails.BranchName
						m.Manifest.Spec.Store.Spec.ConnectorRef = accountConfig.GitDetails.ConnectorRef
						m.Manifest.Spec.Store.Spec.GitFetchType = "Branch"
						m.Manifest.Spec.ValuesPaths = valueFiles

						update = true
					} else if scope.ForceUpdateManifests {
						m.Manifest.Spec.Store.Type = harness.GetServiceManifestStoreType(conn.Type)
						var files []string
						for _, file := range m.Manifest.Spec.Store.Spec.Files {
							files = append(files, fmt.Sprintf("filestore/%s/%s%s", service.Org, service.Project, file))
						}
						var valueFiles []string
						if len(m.Manifest.Spec.Store.ValuesPaths) > 0 {
							for _, v := range m.Manifest.Spec.Store.ValuesPaths {
								valueFiles = append(valueFiles, fmt.Sprintf("filestore/%s/%s%s", service.Org, service.Project, v))
							}
						}
						log.Infof("Setting following file paths : %+v", files)
						m.Manifest.Spec.Store.Spec.Paths = files
						m.Manifest.Spec.Store.Spec.Branch = accountConfig.GitDetails.BranchName
						m.Manifest.Spec.Store.Spec.ConnectorRef = accountConfig.GitDetails.ConnectorRef
						m.Manifest.Spec.Store.Spec.GitFetchType = "Branch"
						m.Manifest.Spec.ValuesPaths = valueFiles

						update = true
					} else {
						log.Infof("Manifest [%s] for Service [%s] is already remote!", m.Manifest.Identifier, service.Name)
					}
				}

				if update {
					// Marshal the modified ServiceYaml back to a YAML string
					modifiedYAML, err := yaml.Marshal(serviceYaml)
					if err != nil {
						log.Errorf(color.RedString("Unable to marshal modified service YAML - %s", err))
						failedServices = append(failedServices, service.Name)
					} else {
						service.YAML = string(modifiedYAML)
					}

					err = service.UpdateService(&api)
					if err != nil {
						log.Errorf(color.RedString("Unable to move service manifests - %s <%s>", service.Name, err, conn))
						failedServices = append(failedServices, service.Name)
					}
				}
				serviceBar.Increment()
			}
			serviceBar.Finish()

			if len(failedServices) > 0 {
				log.Warnf(color.HiYellowString("These Service Manifests (count:%d) failed while moving to remote: \n%s", len(failedServices), strings.Join(failedServices, ",\n")))
			}
		}
		if scope.Overrides {
			log.Info(boldCyan.Sprintf("Processing Service overrides"))
			log.Infof(boldCyan.Sprintf("---Getting Connector Info---"))
			conn, err := api.GetConnector(
				accountConfig.AccountIdentifier,
				accountConfig.FileStoreConfig.Organization,
				accountConfig.FileStoreConfig.Project,
				accountConfig.GitDetails.ConnectorRef,
			)
			if err != nil {
				log.Errorf("Unable to get Connector info - %s", err)
				return
			}
			for _, project := range projectList {
				p := project.Project
				overrideTypes := []harness.OverridesV2Type{harness.OV2_Global, harness.OV2_Service, harness.OV2_Infra, harness.OV2_ServiceInfra}
				var overrides []harness.OverridesV2Content
				log.Infof(boldCyan.Sprintf("---Fetching Overrides V2 ---"))
				for _, ovType := range overrideTypes {
					ov, err := api.GetOverridesV2(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier, ovType)

					if err != nil {
						log.Errorf("Failed to get service overrides V2 type %s - %s", ovType, err)
					} else {
						overrides = append(overrides, ov...)
					}
				}

				if len(overrides) > 0 {
					log.Infof("Updating %d service overrides V2 from FileStore to Git", len(overrides))

					pbTemplate := `{{ blue "Processing: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
					pbBar := pb.ProgressBarTemplate(pbTemplate).Start(len(overrides))

					for _, override := range overrides {
						update := false
						for i := range override.Spec.Manifests {
							m := &override.Spec.Manifests[i]
							if m.Manifest.Spec.Store.Type == "Harness" {

								m.Manifest.Spec.Store.Type = conn.Type
								var files []string
								for _, file := range m.Manifest.Spec.Store.Spec.Files {
									files = append(files, fmt.Sprintf("filestore/%s/%s%s", override.OrgIdentifier, override.ProjectIdentifier, file))
								}
								var valueFiles []string
								if len(m.Manifest.Spec.ValuesPaths) > 0 {
									for _, v := range m.Manifest.Spec.ValuesPaths {
										valueFiles = append(valueFiles, fmt.Sprintf("filestore/%s/%s%s", override.OrgIdentifier, override.ProjectIdentifier, v))
									}
								}
								log.Infof("Setting following file paths : %+v", files)
								m.Manifest.Spec.Store.Spec.Paths = files
								m.Manifest.Spec.Store.Spec.Branch = accountConfig.GitDetails.BranchName
								m.Manifest.Spec.Store.Spec.ConnectorRef = accountConfig.GitDetails.ConnectorRef
								m.Manifest.Spec.Store.Spec.GitFetchType = "Branch"
								m.Manifest.Spec.ValuesPaths = valueFiles

								update = true
							} else if scope.ForceUpdateManifests {
								m.Manifest.Spec.Store.Type = conn.Type
								var files []string
								for _, file := range m.Manifest.Spec.Store.Spec.Files {
									files = append(files, fmt.Sprintf("filestore/%s/%s%s", override.OrgIdentifier, override.ProjectIdentifier, file))
								}
								var valueFiles []string
								if len(m.Manifest.Spec.ValuesPaths) > 0 {
									for _, v := range m.Manifest.Spec.ValuesPaths {
										valueFiles = append(valueFiles, fmt.Sprintf("filestore/%s/%s%s", override.OrgIdentifier, override.ProjectIdentifier, v))
									}
								}
								log.Infof("Setting following file paths : %+v", files)
								m.Manifest.Spec.Store.Spec.Paths = files
								m.Manifest.Spec.Store.Spec.Branch = accountConfig.GitDetails.BranchName
								m.Manifest.Spec.Store.Spec.ConnectorRef = accountConfig.GitDetails.ConnectorRef
								m.Manifest.Spec.Store.Spec.GitFetchType = "Branch"
								m.Manifest.Spec.ValuesPaths = valueFiles

								update = true
							} else {
								log.Infof("Override Manifest [%s] for Environment [%s] is already remote!", m.Manifest.Identifier, override.EnvironmentRef)
							}
							if update {
								// Marshal the modified ServiceYaml back to a YAML string
								log.Infof("Updating Override [%s]", override.Identifier)
								override.YAML = ""
								err := override.UpdateOverrideV2(&api, accountConfig.AccountIdentifier)

								if err != nil {
									log.Errorf(color.RedString("Unable to move service override manifests for environment [%s]", override.EnvironmentRef))
									failedServices = append(failedServices, override.EnvironmentRef)
								}
							}
						}
						pbBar.Increment()
					}
					pbBar.Finish()
				}

			}

			var environmentList []*harness.EnvironmentClass

			log.Info("Getting environments for Account level")
			envs, err := api.GetEnvironments(accountConfig.AccountIdentifier, "", "")
			if err != nil {
				log.Errorf(color.RedString("Unable to get environments for account level. - %s", err))
			}
			environmentList = append(environmentList, envs...)

			orgs, err := api.GetAllOrgs(accountConfig.AccountIdentifier)
			if err != nil {
				log.Errorf(color.RedString("Unable to get organizations for account %s - %s", accountConfig.AccountIdentifier, err))
				return
			}

			for _, o := range orgs {
				org := o.Org
				log.Infof("Getting environements for organization [%s]", org.Identifier)
				envs, err := api.GetEnvironments(accountConfig.AccountIdentifier, org.Identifier, "")
				if err != nil {
					log.Errorf(color.RedString("Unable to get environment for [%s] organization - %s", org.Name, err))
					continue
				}

				environmentList = append(environmentList, envs...)
			}

			for _, project := range projectList {
				p := project.Project
				envs, err := api.GetEnvironments(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
				if err != nil {
					log.Errorf(color.RedString("Unable to get environments for [%s] project. - %s", p.Name, err))
				}
				environmentList = append(environmentList, envs...)
			}

			overridesBar := pb.ProgressBarTemplate(overridesTmpl).Start(len(environmentList))
			var overrideList []*harness.ServiceOverrideContent
			for _, env := range environmentList {
				//Get All environment overrides
				overrides, err := api.GetServiceOverrides(env.Identifier, accountConfig.AccountIdentifier, env.OrgIdentifier, env.ProjectIdentifier)
				if err != nil {
					log.Errorf("Unable to get service overrides for [%s] environment", env.Name)
				}
				overrideList = append(overrideList, overrides...)

				for i := range overrideList {
					overrideYaml, err := overrideList[i].ParseYAML()
					if err != nil {
						log.Errorf(color.RedString("Unable to parse service override YAML - %s", err))
					}
					var update = false
					for i := range overrideYaml.ServiceOverrides.Manifests {
						m := &overrideYaml.ServiceOverrides.Manifests[i]
						if m.Manifest.Spec.Store.Type == "Harness" {
							m.Manifest.Spec.Store.Type = conn.Type
							var files []string
							for _, file := range m.Manifest.Spec.Store.Spec.Files {
								files = append(files, fmt.Sprintf("filestore/%s/%s%s", env.OrgIdentifier, env.ProjectIdentifier, file))
							}
							var valueFiles []string
							if len(m.Manifest.Spec.ValuesPaths) > 0 {
								for _, v := range m.Manifest.Spec.ValuesPaths {
									valueFiles = append(valueFiles, fmt.Sprintf("filestore/%s/%s%s", env.OrgIdentifier, env.ProjectIdentifier, v))
								}
							}
							log.Infof("Setting following file paths : %+v", files)
							m.Manifest.Spec.Store.Spec.Paths = files
							m.Manifest.Spec.Store.Spec.Branch = accountConfig.GitDetails.BranchName
							m.Manifest.Spec.Store.Spec.ConnectorRef = accountConfig.GitDetails.ConnectorRef
							m.Manifest.Spec.Store.Spec.GitFetchType = "Branch"
							m.Manifest.Spec.ValuesPaths = valueFiles

							update = true
						} else if scope.ForceUpdateManifests {
							m.Manifest.Spec.Store.Type = conn.Type
							var files []string
							for _, file := range m.Manifest.Spec.Store.Spec.Files {
								files = append(files, fmt.Sprintf("filestore/%s/%s%s", env.OrgIdentifier, env.ProjectIdentifier, file))
							}
							var valueFiles []string
							if len(m.Manifest.Spec.ValuesPaths) > 0 {
								for _, v := range m.Manifest.Spec.ValuesPaths {
									valueFiles = append(valueFiles, fmt.Sprintf("filestore/%s/%s%s", env.OrgIdentifier, env.ProjectIdentifier, v))
								}
							}
							log.Infof("Setting following file paths : %+v", files)
							m.Manifest.Spec.Store.Spec.Paths = files
							m.Manifest.Spec.Store.Spec.Branch = accountConfig.GitDetails.BranchName
							m.Manifest.Spec.Store.Spec.ConnectorRef = accountConfig.GitDetails.ConnectorRef
							m.Manifest.Spec.Store.Spec.GitFetchType = "Branch"
							m.Manifest.Spec.ValuesPaths = valueFiles

							update = true
						} else {
							log.Infof("ServiceOverride [%s] for Environment [%s] is already remote!", m.Manifest.Identifier, overrideList[i].EnvironmentRef)
						}
						if update {
							// Marshal the modified ServiceYaml back to a YAML string
							modifiedYAML, err := yaml.Marshal(overrideYaml)
							if err != nil {
								log.Errorf(color.RedString("Unable to marshal modified service override YAML - [%s]", err))
								failedServices = append(failedServices, env.Name)
							} else {
								overrideList[i].YAML = string(modifiedYAML)
							}

							err = overrideList[i].UpdateEnvironment(&api)
							if err != nil {
								log.Errorf(color.RedString("Unable to move service override manifests for environment [%s]", env.Name))
								failedServices = append(failedServices, env.Name)
							}
						}
					}
					overridesBar.Increment()
				}
			}
			overridesBar.Finish()
		}
	}
}

func processInfraDefScope(log *logrus.Logger, api harness.APIRequest, customGitDetailsFilePath string, accountConfig harness.Config, p harness.Project, gitX bool) error {

	projectEnvironments, err := api.GetEnvironments(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier)
	if err != nil {
		log.Errorf(color.RedString("Unable to get environments - %s", err))
		return err
	}

	if len(projectEnvironments) > 0 {
		log.Infof("Moving infrastructures of %d environments to remote", len(projectEnvironments))

		pbTemplate := `{{ blue "Processing: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
		pbBar := pb.ProgressBarTemplate(pbTemplate).Start(len(projectEnvironments))

		for _, environment := range projectEnvironments {
			// ONLY TAKE CARE OF INFRA-DEF WHEN ENV IS REMOTE
			if environment.StoreType == "REMOTE" {
				infras, err := api.GetInfrastructures(accountConfig.AccountIdentifier, string(p.OrgIdentifier), p.Identifier, environment.Identifier)
				if err != nil {
					log.Errorf(color.RedString("Unable to list infrastructure of environment - %s [%s]", environment.Identifier, err))
					continue
				}

				if len(infras) > 0 {
					for _, infraDef := range infras {
						if infraDef.StoreType != "REMOTE" {
							accountConfig.GitDetails.FilePath = harness.GetInfrastructureFilePath(gitX, customGitDetailsFilePath, p, *environment, *infraDef)

							err = infraDef.MoveInfrastructureToRemote(&api, accountConfig, environment.Identifier)
							if err != nil {
								log.Errorf(color.RedString("Unable to move infrastruture [%s] from environment [%s] - %s", infraDef.Name, environment.Name, err))
							}
						}
					}
				}
			}
			pbBar.Increment()
		}
		pbBar.Finish()
	}

	return nil
}

func processOverridesV2(log *logrus.Logger, api harness.APIRequest, customGitDetailsFilePath string, cfg harness.Config, p harness.Project, gitX bool) error {

	overrideTypes := []harness.OverridesV2Type{harness.OV2_Global, harness.OV2_Service, harness.OV2_Infra, harness.OV2_ServiceInfra}
	var overrides []harness.OverridesV2Content

	for _, ovType := range overrideTypes {
		ov, err := api.GetOverridesV2(cfg.AccountIdentifier, string(p.OrgIdentifier), p.Identifier, ovType)

		if err != nil {
			log.Errorf("Failed to get service overrides V2 type %s - %s", ovType, err)
		} else {
			overrides = append(overrides, ov...)
		}
	}

	if len(overrides) > 0 {
		log.Infof("Moving %d service overrides V2 to remote", len(overrides))

		pbTemplate := `{{ blue "Processing: " }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} `
		pbBar := pb.ProgressBarTemplate(pbTemplate).Start(len(overrides))

		for _, override := range overrides {
			if override.StoreType != "REMOTE" {
				cfg.GitDetails.FilePath = harness.GetOverridesV2FilePath(gitX, customGitDetailsFilePath, p, override)

				if err := override.MoveToRemote(&api, cfg); err != nil {
					log.Errorf(color.RedString("Unable to move overrides V2 [%s] - %s", override, err))
				}
			}
			pbBar.Increment()
		}
		pbBar.Finish()
	}

	return nil
}

func servicesSummary(log *logrus.Logger, boldCyan *color.Color, failedTemplates []string, failedServices []string, alreadyRemoteServices []string, services []*harness.ServiceClass) {
	log.Infof(boldCyan.Sprintf("---Services---"))
	if len(failedTemplates) > 0 {
		log.Warnf(color.HiYellowString("These services (count:%d) failed while moving to remote: \n%s", len(failedServices), strings.Join(failedServices, ",\n")))
	}
	if len(alreadyRemoteServices) > 0 {
		log.Warnf(color.HiYellowString("These services (count:%d) already remote: \n%s", len(alreadyRemoteServices), strings.Join(alreadyRemoteServices, ",\n")))
	}
	log.Infof(color.GreenString("Processed total of %d services", len(services)))
	log.Infof(color.GreenString("------"))
	log.Infof(color.GreenString("Moved services to remote!"))
	log.Infof(color.GreenString("------"))
}

func templatesSummary(log *logrus.Logger, boldCyan *color.Color, failedTemplates []string, templates []harness.Template) {
	log.Infof(boldCyan.Sprintf("---Templates---"))
	if len(failedTemplates) > 0 {
		log.Warnf(color.HiYellowString("These templates (count:%d) failed while moving to remote: \n%s", len(failedTemplates), strings.Join(failedTemplates, ",\n")))
	}
	log.Infof(color.GreenString("Processed total of %d templates", len(templates)))
	log.Infof(color.GreenString("------"))
	log.Infof(color.GreenString("Moved templates to remote!"))
	log.Infof(color.GreenString("------"))
}

func pipelinesSummary(log *logrus.Logger, boldCyan *color.Color, failedPipelines []string, pipelines []harness.PipelineContent) {
	log.Infof(boldCyan.Sprintf("---Pipelines---"))
	if len(failedPipelines) > 0 {
		log.Warnf(color.HiYellowString("These pipelines (count:%d) failed while moving to remote: \n%s", len(failedPipelines), strings.Join(failedPipelines, ",\n")))
	}
	log.Infof(color.GreenString("Processed total of %d pipelines", len(pipelines)))
	log.Infof(color.GreenString("------"))
	log.Infof(color.GreenString("Moved pipelines to remote!"))
	log.Infof(color.GreenString("------"))
}

func environmentsSummary(log *logrus.Logger, boldCyan *color.Color, failedEnvs []string, envs []*harness.EnvironmentClass) {
	log.Infof(boldCyan.Sprintf("---Environments---"))
	if len(failedEnvs) > 0 {
		log.Warnf(color.HiYellowString("These environments (count:%d) failed while moving to remote: \n%s", len(failedEnvs), strings.Join(failedEnvs, ",\n")))
	}
	log.Infof(color.GreenString("Processed total of %d environments", len(envs)))
	log.Infof(color.GreenString("------"))
	log.Infof(color.GreenString("Moved environments to remote!"))
	log.Infof(color.GreenString("------"))
}
