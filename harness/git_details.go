package harness

import (
	"fmt"
)

func GetPipelineFilePath(gitX bool, customGitDetailsFilePath string, p Project, pipeline PipelineContent) string {
	if len(customGitDetailsFilePath) == 0 {
		if gitX {
			return fmt.Sprintf(".harness/orgs/%s/projects/%s/pipelines/%s.yaml", string(p.OrgIdentifier), p.Identifier, pipeline.Identifier)
		}
		return "pipelines/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + pipeline.Identifier + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + pipeline.Identifier + ".yaml"
	}
}

func GetTemplateFilePath(gitX bool, customGitDetailsFilePath string, p Project, template Template) string {
	if len(customGitDetailsFilePath) == 0 {
		if gitX {
			return fmt.Sprintf(".harness/orgs/%s/projects/%s/templates/%s/%s.yaml", string(p.OrgIdentifier), p.Identifier, template.Identifier, template.VersionLabel)
		}
		return "templates/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + template.Identifier + "-" + template.VersionLabel + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + template.Identifier + "-" + template.VersionLabel + ".yaml"
	}
}

func GetServiceFilePath(gitX bool, customGitDetailsFilePath string, p Project, service ServiceClass) string {
	if len(customGitDetailsFilePath) == 0 {
		if gitX {
			return fmt.Sprintf(".harness/orgs/%s/projects/%s/services/%s.yaml", string(p.OrgIdentifier), p.Identifier, service.Identifier)
		}
		return "services/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + service.Identifier + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + service.Identifier + ".yaml"
	}
}

func GetEnvironmentFilePath(gitX bool, customGitDetailsFilePath string, p Project, env EnvironmentClass) string {
	if len(customGitDetailsFilePath) == 0 {
		if gitX {
			return fmt.Sprintf(".harness/orgs/%s/projects/%s/envs/%s/%s.yaml", string(p.OrgIdentifier), p.Identifier, getEnvType(env), env.Identifier)
		}
		return "environments/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + env.Identifier + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + env.Identifier + ".yaml"
	}
}

func GetInfrastructureFilePath(gitX bool, customGitDetailsFilePath string, p Project, env EnvironmentClass, infraDef Infrastructure) string {
	if len(customGitDetailsFilePath) == 0 {
		if gitX {
			return fmt.Sprintf(".harness/orgs/%s/projects/%s/envs/%s/%s/infras/%s.yaml", string(p.OrgIdentifier), p.Identifier, getEnvType(env), env.Identifier, infraDef.Identifier)
		}
		return "environments/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + env.Identifier + "-" + infraDef.Identifier + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + env.Identifier + "-" + infraDef.Identifier + ".yaml"
	}
}

func GetOverridesV2FilePath(gitX bool, customGitDetailsFilePath string, p Project, ov OverridesV2Content) string {
	if len(customGitDetailsFilePath) == 0 {
		if gitX {
			switch ov.Type {
			case OV2_Global:
				return fmt.Sprintf(".harness/orgs/%s/projects/%s/overrides/%s/overrides.yaml", string(p.OrgIdentifier), p.Identifier, ov.EnvironmentRef)
			case OV2_Service:
				return fmt.Sprintf(".harness/orgs/%s/projects/%s/overrides/%s/services/%s/overrides.yaml",
					string(p.OrgIdentifier), p.Identifier, ov.EnvironmentRef, ov.ServiceRef)
			case OV2_Infra:
				return fmt.Sprintf(".harness/orgs/%s/projects/%s/overrides/%s/infras/%s/overrides.yaml",
					string(p.OrgIdentifier), p.Identifier, ov.EnvironmentRef, ov.InfraIdentifier)
			case OV2_ServiceInfra:
				return fmt.Sprintf(".harness/orgs/%s/projects/%s/overrides/%s/services/%s/infras/%s/overrides.yaml",
					string(p.OrgIdentifier), p.Identifier, ov.EnvironmentRef, ov.ServiceRef, ov.InfraIdentifier)
			default:
				panic(fmt.Sprintf("unrecognized overrides V2 type %s", ov.Type))
			}
		}
		return "overrides/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + ov.Identifier + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + ov.Identifier + ".yaml"
	}
}

func GetInputsetFilePath(gitX bool, customGitDetailsFilePath string, p Project, is *InputsetContent) string {
	if len(customGitDetailsFilePath) == 0 {
		if gitX {
			return fmt.Sprintf(".harness/orgs/%s/projects/%s/pipelines/%s/input_sets/%s.yaml", string(p.OrgIdentifier), p.Identifier, is.PipelineIdentifier, is.Identifier)
		}
		return "input_sets/" + string(p.OrgIdentifier) + "/" + p.Identifier + "/" + is.PipelineIdentifier + "/" + is.Identifier + ".yaml"
	} else {
		return customGitDetailsFilePath + "/" + is.PipelineIdentifier + "/" + is.Identifier + ".yaml"
	}
}

func getEnvType(env EnvironmentClass) string {
	switch env.Type {
	case "Production":
		return "production"
	case "PreProduction":
		return "pre_production"
	default:
		return "unknown"
	}
}
