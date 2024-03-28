package harness

type InfraDefResult struct {
	Status        string       `json:"status"`
	Data          InfraDefData `json:"data"`
	MetaData      interface{}  `json:"metaData"`
	CorrelationID string       `json:"correlationId"`
}

type InfraDefData struct {
	TotalPages    int64              `json:"totalPages"`
	TotalItems    int64              `json:"totalItems"`
	PageItemCount int64              `json:"pageItemCount"`
	PageSize      int64              `json:"pageSize"`
	Content       []*InfraDefContent `json:"content"`
	PageIndex     int64              `json:"pageIndex"`
	Empty         bool               `json:"empty"`
}

type InfraDefContent struct {
	Infrastructure Infrastructure `json:"infrastructure"`
	CreatedAt      int64          `json:"createdAt"`
	LastModifiedAt int64          `json:"lastModifiedAt"`
}

type Infrastructure struct {
	AccountID         string `json:"accountId"`
	Identifier        string `json:"identifier"`
	OrgIdentifier     string `json:"orgIdentifier"`
	ProjectIdentifier string `json:"projectIdentifier"`
	EnvironmentRef    string `json:"environmentRef"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	DeploymentType    string `json:"deploymentType"`
	YAML              string `json:"yaml"`
	StoreType         string `json:"storeType"`
}
