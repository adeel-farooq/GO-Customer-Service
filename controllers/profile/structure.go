package profile

type GetUserInfoRequest struct {
	SiteUsersId int64 `json:"siteUsersId"`
	CustomersId int64 `json:"customersId"`
}

// gob-safe: keep Details/Errors as string
type DbResultRPC struct {
	ID      int64  `json:"id"`
	Id      int64  `json:"Id"`
	Status  string `json:"status"`
	Details string `json:"details"` // RAW JSON string
	Errors  string `json:"errors"`  // DB error string/JSON array string or ""
}

type ErrorResultDto struct {
	ErrorType   string `json:"errorType"`
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}
