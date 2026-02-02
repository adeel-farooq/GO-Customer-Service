package customer

// CustomerService is the RPC service.
// Example RPC method string: "CustomerService.RegisterCustomerUser".
type CustomerService struct{}

// This struct matches the .NET V2 DTO:
// AvamaeTemplate.Core.Types.Dtos.PublicRole.RegistrationV2Module.CustomerUserRegistrationDto
//
// NOTE: If your gateway payload includes extra fields, Go will ignore unknown JSON fields.
type CustomerUserRegistrationRequest struct {
	FirstName    string `json:"firstName"`
	MiddleNames  string `json:"middleNames"`
	LastName     string `json:"lastName"`
	CurrencyCode string `json:"currencyCode"`

	EmailAddress          string `json:"emailAddress"`
	Password              string `json:"password"`
	ConfirmPassword       string `json:"confirmPassword"`
	RegisteredCountryCode string `json:"registeredCountryCode"`

	PhoneNumber string `json:"phoneNumber"`

	// NOTE: keep as bool (not *bool) to be RPC/gob friendly when gateway sends a non-pointer bool.
	BConductsThirdPartyPayments bool `json:"bConductsThirdPartyPayments"`

	RequestedProducts []int `json:"requestedProducts"`

	// Values normally derived from HTTP request in .NET; gateway can pass them.
	Domain                      string `json:"domain,omitempty"`
	CultureInfo                 string `json:"cultureInfo,omitempty"`
	ConsentIpAddress            string `json:"consentIpAddress,omitempty"`
	ConsentCountryIso2Code      string `json:"consentCountryIso2Code,omitempty"`
	ConsentStateProvinceIsoCode string `json:"consentStateProvinceIsoCode,omitempty"`
}

// Mirrors AvamaeTemplate.Core.Types.Results.DbResultDto
// type DbResult struct {
// 	// NOTE: net/rpc uses gob which matches by Go field names.
// 	// Some clients use `ID`, others use `Id`. We send both to maximize compatibility.
// 	ID      int    `json:"id"`
// 	Id      int    `json:"-"`
// 	Status  string `json:"status"`
// 	Details string `json:"details"`
// 	Errors  string `json:"errors"`
// }
type DbResult struct {
	ID      int    `json:"id"`
	Id      int    `json:"-"`
	Status  string `json:"status"`
	Details any    `json:"details"` // optional
	Errors  any    `json:"errors"`  // ✅ important
}

// Mirrors .NET ErrorResultDto where ErrorType is JsonIgnore.
// We keep it to let the gateway map HTTP status codes, but do not serialize it.
type ErrorResult struct {
	ErrorType   string `json:"-"`
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}

// Mirrors .NET ResultDto output JSON:
// { "id": <long>, "status": "1"|"0", "errors": [ {fieldName,messageCode}, ... ] }
type Result struct {
	Id     int64         `json:"id"`
	Status string        `json:"status"`
	Errors []ErrorResult `json:"errors"`
}

func NewSuccessResult(id int64) Result {
	return Result{Id: id, Status: "1", Errors: []ErrorResult{}}
}

func NewFailureResult(errs ...ErrorResult) Result {
	return Result{Id: 0, Status: "0", Errors: errs}
}
