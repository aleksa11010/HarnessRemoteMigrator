package harness

type OverridesV2Response struct {
	Status  string          `json:"status"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    OverridesV2Data `json:"data"`
}

type OverridesV2Data struct {
	TotalPages    int64                `json:"totalPages"`
	TotalItems    int64                `json:"totalItems"`
	PageItemCount int64                `json:"pageItemCount"`
	PageSize      int64                `json:"pageSize"`
	Content       []OverridesV2Content `json:"content"`
	PageIndex     int64                `json:"pageIndex"`
	Empty         bool                 `json:"empty"`
}

type OverridesV2Content struct {
	Identifier        string          `json:"identifier"`
	AccountID         string          `json:"accountId"`
	OrgIdentifier     string          `json:"orgIdentifier"`
	ProjectIdentifier string          `json:"projectIdentifier"`
	EnvironmentRef    string          `json:"environmentRef"`
	ServiceRef        string          `json:"serviceRef,omitempty"`
	InfraIdentifier   string          `json:"infraIdentifier,omitempty"`
	Type              OverridesV2Type `json:"type"`
	StoreType         string          `json:"storeType,omitempty"`
	YAML              string          `json:"yaml"`
	Spec              OverrideSpec    `json:"spec,omitempty"`
}

type OverrideSpec struct {
	Variables []struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Value    string `json:"value"`
		Required bool   `json:"required"`
	} `json:"variables"`
	CliEnvironmentVariables []struct {
		Name  string `json:"name"`
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"cliEnvironmentVariables"`
	Manifests []struct {
		Manifest struct {
			Identifier string `json:"identifier"`
			Type       string `json:"type"`
			Spec       struct {
				Store struct {
					Type string `json:"type"`
					Spec struct {
						ConnectorRef              string   `json:"connectorRef"`
						GitFetchType              string   `json:"gitFetchType"`
						Paths                     []string `json:"paths"`
						Branch                    string   `json:"branch"`
						Files                     []string `json:"files"`
						SkipResourceVersioning    bool     `json:"skipResourceVersioning"`
						EnableDeclarativeRollback bool     `json:"enableDeclarativeRollback"`
					} `json:"spec"`
				} `json:"store"`
				ValuesPaths []string `json:"valuesPaths"`
			} `json:"spec"`
		} `json:"manifest"`
	} `json:"manifests"`
	ConfigFiles []struct {
		ConfigFile struct {
			Identifier string `json:"identifier"`
			Type       string `json:"type"`
			Spec       struct {
				Store struct {
					Type string `json:"type"`
					Spec struct {
						ConnectorRef string   `json:"connectorRef"`
						GitFetchType string   `json:"gitFetchType"`
						Paths        []string `json:"paths"`
						Branch       string   `json:"branch"`
						Files        []string `json:"files"`
						SecretFiles  []string `json:"SecretFiles"`
					} `json:"spec"`
				} `json:"store"`
			} `json:"spec"`
		} `json:"configFile"`
	} `json:"configFiles"`
}

type OverridesV2Type string

const (
	OV2_Global       OverridesV2Type = "ENV_GLOBAL_OVERRIDE"
	OV2_Service      OverridesV2Type = "ENV_SERVICE_OVERRIDE"
	OV2_Infra        OverridesV2Type = "INFRA_GLOBAL_OVERRIDE"
	OV2_ServiceInfra OverridesV2Type = "INFRA_SERVICE_OVERRIDE"
)
