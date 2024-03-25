package harness

func GetPipelineFilePath(customGitDetailsFilePath string, p Project, pipeline PipelineContent) string {
	if len(customGitDetailsFilePath) == 0 {
		return "pipelines/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + pipeline.Identifier + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + pipeline.Identifier + ".yaml"
	}
}

func GetTemplateFilePath(customGitDetailsFilePath string, p Project, template Template) string {
	if len(customGitDetailsFilePath) == 0 {
		return "templates/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + template.Identifier + "-" + template.VersionLabel + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + template.Identifier + "-" + template.VersionLabel + ".yaml"
	}
}

func GetServiceFilePath(customGitDetailsFilePath string, p Project, service ServiceClass) string {
	if len(customGitDetailsFilePath) == 0 {
		return "services/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + service.Identifier + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + service.Identifier + ".yaml"
	}
}
