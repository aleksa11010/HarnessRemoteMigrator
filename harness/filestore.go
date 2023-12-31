package harness

// Generated by https://quicktype.io

type FileStore struct {
	Status        string        `json:"status"`
	Data          FileStoreData `json:"data"`
	MetaData      interface{}   `json:"metaData"`
	CorrelationID string        `json:"correlationId"`
}

type FileStoreData struct {
	Content          []FileStoreContent `json:"content"`
	Pageable         FileStorePageable  `json:"pageable"`
	Last             bool               `json:"last"`
	TotalPages       int64              `json:"totalPages"`
	TotalElements    int64              `json:"totalElements"`
	Sort             FileStoreSort      `json:"sort"`
	Number           int64              `json:"number"`
	First            bool               `json:"first"`
	NumberOfElements int64              `json:"numberOfElements"`
	Size             int64              `json:"size"`
	Empty            bool               `json:"empty"`
}

type FileStoreContent struct {
	AccountIdentifier string        `json:"accountIdentifier"`
	Identifier        string        `json:"identifier"`
	Name              string        `json:"name"`
	FileUsage         string        `json:"fileUsage"`
	Type              string        `json:"type"`
	ParentIdentifier  string        `json:"parentIdentifier"`
	Tags              []interface{} `json:"tags"`
	MIMEType          string        `json:"mimeType"`
	Path              string        `json:"path"`
	CreatedBy         EdBy          `json:"createdBy"`
	LastModifiedBy    EdBy          `json:"lastModifiedBy"`
	LastModifiedAt    int64         `json:"lastModifiedAt"`
}

type EdBy struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type FileStorePageable struct {
	Sort       Sort  `json:"sort"`
	PageSize   int64 `json:"pageSize"`
	PageNumber int64 `json:"pageNumber"`
	Offset     int64 `json:"offset"`
	Paged      bool  `json:"paged"`
	Unpaged    bool  `json:"unpaged"`
}

type FileStoreSort struct {
	Sorted   bool `json:"sorted"`
	Unsorted bool `json:"unsorted"`
	Empty    bool `json:"empty"`
}
