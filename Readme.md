# Harness Migration Utility - Inline to Remote entities 


## Prerequisites
To run this utility tool you need to have a Git connector on your account with proper premissions.

When providing the identifier of your Git Connector you need to prefix account or org if account is not located under the project, if moving multiple projects please use account or org level connector that will have access to all projects.
## Usage
There are two ways to run the utlity, by providing a Configuration YAML file or by providing CLI arguments.
### Config File
Configuration file will contain the following information : 
```yaml
accountIdentifier: AccountIdentifier # Your Account ID
apiKey: pat.xxx.yyy.zzz # Your PAT/SAT token
targetProjects: # We can target only specific projects by providing their identifiers here
  - "target_project1" 
  - "target_project2"
excludeProjects: # Or we can exclude specific projects by providing their identifiers here
  - "exclude_project1"
  - "exclude_project2"
gitDetails: # This is where we setup remote location for pipelines/templates
  branch_name: "migration" # Branch must exists before running 
  commit_message: "Migrating piplines from inline to remote" # Your commit message
  connector_ref: "account.HarnessRemoteTest" # Git Connector Identifiers
  repo_name: "HarnessRemoteTest" # Your Repo name
```
### CLI Arguments

***To be added***

## Supported entities
1. Pipelines
2. Templates