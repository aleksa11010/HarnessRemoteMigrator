package harness

type ListInputsetResponse struct {
	Status        string           `json:"status"`
	Data          ListInputsetData `json:"data"`
	CorrelationID string           `json:"correlationId"`
}

type ListInputsetData struct {
	TotalPages    int64              `json:"totalPages"`
	TotalItems    int64              `json:"totalItems"`
	PageItemCount int64              `json:"pageItemCount"`
	PageSize      int64              `json:"pageSize"`
	Content       []*InputsetContent `json:"content"`
	PageIndex     int64              `json:"pageIndex"`
	Empty         bool               `json:"empty"`
}

type InputsetContent struct {
	Identifier            string                `json:"identifier"`
	Name                  string                `json:"name"`
	PipelineIdentifier    string                `json:"pipelineIdentifier"`
	InputSetType          string                `json:"inputSetType"`
	EntityValidityDetails EntityValidityDetails `json:"entityValidityDetails"`
	StoreType             string                `json:"storeType"`
}
