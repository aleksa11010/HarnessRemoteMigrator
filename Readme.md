# Harness Migration Utility: Inline to Remote Entities

## Introduction

The Harness Migration Utility is a tool designed to facilitate the migration of inline entities to remote entities. This document provides a comprehensive guide on the prerequisites, usage, and supported entities for the utility.

## Prerequisites

Before running this utility tool, ensure that you have a Git connector on your account with the appropriate permissions.

If you are providing the identifier of your Git connector, prefix it with 'account' or 'org' if the connector is not located under the project. 

### ***For migrating multiple projects, use an account or org level connector that has access to all projects.***

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
targetServices: # Similar to projects, we can target specific services
  - "service1": "project1" # Difference is we need to provide Service and Project ID
  - "service2": "project2"
excludeServices: # Exclude specific services by providing Service and Project IDs
  - "exclude_service1": "project1"
  - "exclude_service2": "project2"
gitDetails: # Setup remote location for pipelines/templates here
  branch_name: "migration" # Branch must exist before running 
  commit_message: "Migrating pipelines from inline to remote" # Your commit message
  connector_ref: "account.HarnessRemoteTest" # Git Connector Identifiers
  repo_name: "HarnessRemoteTest" # Your Repo name
fileStoreConfig:
  branch: "migration" # Branch is required to push File Store files
  # More Configuration (optional)
  url: "https://github.com/aleksa11010/FileStore.git" # Provide Repo URL
  organization: "default" # Connector settings
  project: "migration_project" # Connector settings
  connector_ref: "account.Connector2" # Connector Identifier
```

**Note:**
- If you provide Org and Project settings, the utility will attempt to pull the connector from there.
- If you are doing all projects on an account, we suggest using account level connector.

If no repo URL is provided, the connector from GitDetails/FileStoreConfig will be used to pull the URL from the spec.

### Running the Migration Utility
You can run the migration utility using the following commands:

**Run for all supported entities:**
```
./harness-remote-migrator -config /path/to/config.yaml -all
```
**Run ONLY for pipelines:**
```
./harness-remote-migrator -config /path/to/config.yaml -pipelines
```
**Run ONLY for templates:**
```
./harness-remote-migrator -config /path/to/config.yaml -templates
```

**Run ONLY for environments:**

```sh
./harness-remote-migrator -config /path/to/config.yaml -environments
```

**Run ONLY for infrastructure definition:**

```sh
./harness-remote-migrator -config /path/to/config.yaml -infraDef
```

**Run ONLY for file store:**
```
./harness-remote-migrator -config /path/to/config.yaml -filestore
```
**Run Service Manifest migration - MUST BE ACCOMPANIED BY file store flag:**
```
./harness-remote-migrator -config /path/to/config.yaml -filestore -service
```
**If the service has remote manifest already - we need to force the update to new file store location**
```
./harness-remote-migrator -config /path/to/config.yaml -filestore -service -update-service
```
**You can use any combination of above commands.**

## Utility Commands

**URL Encoding for strings**
- You can use the flag `url-encode-string` to enable the URL friendly encoding of remote paths

**CG Like Folder structure**
- You can use the flag `alt-path` to enable the CG-like folder structure in Git. 

**Custom Remote Path**

- You can use the flag `custom-remote-path` to point where to save YAMLs inside the remote repository.
- When using this argument you should avoid running multiple migrations at same time or all the files will be stored at the same path.

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

This folder structure will be used to reference the files in your Services/Environments.

## Supported Entities

1. Pipelines
1. Templates
1. Services
1. Environments
1. Infrastructure Definition
1. File Store
1. Service Manifests/Values
