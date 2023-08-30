

# Harness Migration Utility: Inline to Remote Entities

## Introduction
The Harness Migration Utility is a tool designed to facilitate the migration of inline entities to remote entities. This document provides a comprehensive guide on the prerequisites, usage, and supported entities for the utility.

## Prerequisites
Before running this utility tool, ensure that you have a Git connector on your account with the appropriate permissions.

If you are providing the identifier of your Git connector, prefix it with 'account' or 'org' if the connector is not located under the project. For migrating multiple projects, use an account or org level connector that has access to all projects.

## Usage
The utility can be run in two ways: by providing a Configuration YAML file or by providing CLI arguments.

### Config File
The configuration file contains the following information:
```yaml
accountIdentifier: AccountIdentifier # Your Account ID
apiKey: pat.xxx.yyy.zzz # Your PAT/SAT token
targetProjects: # Target specific projects by providing their identifiers here
  - "target_project1" 
  - "target_project2"
excludeProjects: # Exclude specific projects by providing their identifiers here
  - "exclude_project1"
  - "exclude_project2"
gitDetails: # Setup remote location for pipelines/templates here
  branch_name: "migration" # Branch must exist before running 
  commit_message: "Migrating pipelines from inline to remote" # Your commit message
  connector_ref: "account.HarnessRemoteTest" # Git Connector Identifiers
  repo_name: "HarnessRemoteTest" # Your Repo name
fileStoreConfig:
  branch: "migration" # Branch required to push File Store files
  # More Configuration (optional)
  url: "https://github.com/aleksa11010/FileStore.git" # Provide Repo URL
  organization: "default" # Connector settings
  project: "migration_project" # Connector settings
  connector_ref: "account.Connector2"
```

**Note:**
- If you provide Org and Project settings, the utility will attempt to pull the connector from there.
- If you are doing all projects on an account, we suggest using account level connector.

If no repo URL is provided, the connector from GitDetails/FileStoreConfig will be used to pull the URL from the spec.

### Running the Migration Utility
You can run the migration utility using the following command:
```
./harness-remote-migrator -config /path/to/config.yaml
```

### CLI Arguments
***To be added***

## File Store Migration
The utility migrates File Store files to a remote Git repository by creating local copies from your Harness account and initializing a Git repo.

The files are structured in following order Account, Org, and Project levels. 
Example structure:
```
.
├── account
│   └── Test
└── org
    └── default
        ├── HelloWorld
        └── example_project
            ├── Bar
            └── Foo
```
After downloading and structuring the files, they are committed and pushed to repo based on your configuration.

## Supported Entities
1. Pipelines
2. Templates
3. File Store
