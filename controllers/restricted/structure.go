package restricted

type DbResult1 struct {
	Id      int    `json:"id"`
	Status  string `json:"status"`
	Details string `json:"details"`
	Errors  string `json:"errors"`
}

// CountryRestrictedActionsService is the RPC service.
// Example RPC method string: "CountryRestrictedActionsService.GetErrorMessages".
type RestrictedService struct{}

type GetErrorMessagesRequest struct{}

// ErrorMessage mirrors the .NET ErrorMessageDto.
type ErrorMessage struct {
	ErrorCode string `json:"errorCode"`
	ErrorText string `json:"errorText"`
}

// DbResult mirrors the stored-proc output used across modules.
// NOTE: net/rpc uses gob which matches exported Go field names.
// Some clients decode into `ID`, others into `Id`, so we send both.

type DbResult struct {
	ID      int    `json:"id"`
	Id      int    `json:"Id,omitempty"` // agar aap dono rakhna chahte ho
	Status  string `json:"status"`
	Details any    `json:"details"`
	Errors  any    `json:"errors"`
}

type ErrorResult struct {
	ErrorType   string `json:"errorType"`
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}
type DbResultRPC struct {
	ID     int    `json:"id"`
	Id     int    `json:"Id,omitempty"`
	Status string `json:"status"`

	Details string `json:"details"` // ✅ ALWAYS string (raw JSON)
	Errors  string `json:"errors"`  // ✅ ALWAYS string
}

type Country struct {
	Name              string  `json:"name"`
	IsoCode           string  `json:"isoCode"`
	PersonalErrorCode *string `json:"personalErrorCode"` // null possible
	BusinessErrorCode *string `json:"businessErrorCode"` // null possible
	Id                int     `json:"id"`
}

type ProductService struct {
	Name        string  `json:"name"`
	Id          int     `json:"id"`
	Description *string `json:"description"` // null possible
}

type RegistrationDetails struct {
	Countries       []Country        `json:"countries"`
	ProductServices []ProductService `json:"productServices"`
}

type RegistrationDetailsResult struct {
	ID      int                 `json:"id"`
	Status  string              `json:"status"`
	Details RegistrationDetails `json:"details"`
	Errors  []string            `json:"errors"`
}
