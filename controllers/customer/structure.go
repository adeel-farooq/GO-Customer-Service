package customer

import "encoding/json"

type SubmitVerificationFormRequest struct {
	CustomersId int64           `json:"customersId"`
	SiteUsersId int64           `json:"siteUsersId"`
	AddedBy     string          `json:"addedBy"`
	Domain      string          `json:"domain"`
	Command     json.RawMessage `json:"command"`
}

type DbResultRPC struct {
	ID      int64  `json:"id"`
	Id      int64  `json:"Id"`
	Status  string `json:"status"`
	Details string `json:"details"`
	Errors  string `json:"errors"`
}

// ---- Business Module V3 ----
type GetVerificationFormRequest struct {
	CustomersId int64 `json:"customersId"`
	SiteUsersId int64 `json:"siteUsersId"`
}

type GetDropdownDataRequest struct {
	CustomersId int64 `json:"customersId"`
	SiteUsersId int64 `json:"siteUsersId"`
}

// These match .NET GetVerificationFormDto JSON shape.
// We keep fields flexible: FormData is raw JSON.
type GetVerificationFormDto struct {
	RawFormData     string           `json:"rawFormData"`
	InnerErrors     string           `json:"innerErrors"`
	ListInnerErrors []ErrorResultDto `json:"listInnerErrors"`
	FormData        json.RawMessage  `json:"formData"`
}

type ErrorResultDto struct {
	ErrorType   string `json:"errorType"`
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}

// type DbResultRPC struct {
// 	ID      int64  `json:"id"`
// 	Id      int64  `json:"Id"`
// 	Status  string `json:"status"`
// 	Details string `json:"details"` // RAW JSON string
// 	Errors  string `json:"errors"`  // RAW JSON string OR "" OR plain string
// }

type SaveFormResponse struct {
	CompanyName        string             `json:"companyName"`
	ListInnerErrors    []any              `json:"listInnerErrors"`
	SignificantParties []SignificantParty `json:"significantParties"`
}

// type SignificantParty struct {
// 	Id           int64  `json:"id"`
// 	EmailAddress string `json:"emailAddress"`
// }

// DocumentGroupList request/response structs
type GetDocumentGroupListRequest struct {
	CustomersId int64 `json:"customersId"`
	SiteUsersId int64 `json:"siteUsersId"`
}

// ---- KYC Document V3 ----
type AddKYCDocumentRequest struct {
	CustomersId int64  `json:"customersId"`
	SiteUsersId int64  `json:"siteUsersId"`
	AddedBy     string `json:"addedBy"`

	DocumentId int64  `json:"documentId"`
	Note       string `json:"note"`

	CustomersBusinessDocumentsId *int64 `json:"customersBusinessDocumentsId,omitempty"`

	// file (optional)
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	FileBytes   []byte `json:"fileBytes,omitempty"`
}

type UploadedDocumentDetails struct {
	FileName string `json:"fileName"`
}

// CustomerService is the RPC service.
// Example RPC method string: "CustomerService.RegisterCustomerUser".
type CustomerService struct{}
type BusinessService struct{}

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

type UploadDocumentRequest struct {
	// file payload
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	FileBytes   []byte `json:"fileBytes"` // gateway sends raw bytes (base64 happens in JSON transport automatically in Go RPC)

	// metadata
	CustomersId int64 `json:"customersId"`
	// DocumentTypeId int64  `json:"documentTypeId"`
	DocumentId   int64  `json:"documentId"`
	DocumentName string `json:"documentName,omitempty"`
	Description  string `json:"description,omitempty"`

	// context
	SiteUsersId int64 `json:"siteUsersId"` // if you have auth context
}

type SaveVerificationFormRequest struct {
	CustomersId int64  `json:"customersId"`
	SiteUsersId int64  `json:"siteUsersId"`
	AddedBy     string `json:"addedBy"`

	// raw command JSON (same as .NET VerificationFormDto)
	Command json.RawMessage `json:"command"`
}

// Minimal parts we need for missing files + ownership graph
type VerificationForm struct {
	BusinessVerificationStep string            `json:"businessVerificationStep"`
	OwnerInformation         *OwnerInformation `json:"ownerInformation,omitempty"`
	RegistrationInformation  json.RawMessage   `json:"registrationInformation,omitempty"`
	OperationsInformation    *OperationsInfo   `json:"operationsInformation,omitempty"`
}

type OwnerInformation struct {
	BeneficialOwnersStructure  json.RawMessage             `json:"beneficialOwnersStructure,omitempty"`
	IndividualBeneficialOwners []IndividualBeneficialOwner `json:"individualBeneficialOwners,omitempty"`
}

type IndividualBeneficialOwner struct {
	ProofOfAddressFilename string `json:"proofOfAddressFilename,omitempty"`
}

type OperationsInfo struct {
	FinancialInstitutionFormFileName string `json:"financialInstitutionFormFileName,omitempty"`
}

// Ownership graph structs (ported from OwnershipTreeDto.cs)
type OwnershipTreeNode struct {
	BIsBusiness       *bool               `json:"bIsBusiness,omitempty"`
	OwnersGuid        string              `json:"ownersGuid,omitempty"`
	PercentageOwned   float64             `json:"percentageOwned,omitempty"`
	BControllingParty *bool               `json:"bControllingParty,omitempty"`
	PositionAtCompany string              `json:"positionAtCompany,omitempty"`
	Children          []OwnershipTreeNode `json:"children,omitempty"`
}

type OwnershipGraph struct {
	Vertices []OwnershipVertex `json:"vertices"`
	Edges    []OwnershipEdge   `json:"edges"`
}

type OwnershipVertex struct {
	OwnersGuid string              `json:"ownersGuid"`
	VertexData OwnershipVertexData `json:"vertexData"`
}

type OwnershipVertexData struct {
	BIsBusiness *bool `json:"bIsBusiness,omitempty"`
}

type OwnershipEdge struct {
	SourceVertexGuid string            `json:"sourceVertexGuid"`
	TargetVertexGuid string            `json:"targetVertexGuid"`
	EdgeData         OwnershipEdgeData `json:"edgeData"`
	BInSpanningTree  bool              `json:"bInSpanningTree"`
}

type OwnershipEdgeData struct {
	PercentageOwned   float64 `json:"percentageOwned,omitempty"`
	PositionAtCompany string  `json:"positionAtCompany,omitempty"`
	BControllingParty *bool   `json:"bControllingParty,omitempty"`
}

// ResultDto shape (simplified) – .NET JSON format compatible
type ResultDto struct {
	Id     int64             `json:"id"`
	Errors []ErrorResultFile `json:"errors"`
}

type DbResultFile struct {
	ID      int64  `json:"id"`
	Id      int64  `json:"Id"`
	Status  string `json:"status"`
	Details string `json:"details"` // RAW JSON string (safe for net/rpc gob)
	Errors  string `json:"errors"`  // RAW JSON string (array) or empty
}

// type ErrorResultFile struct {
// 	ErrorType   string `json:"errorType"`
// 	FieldName   string `json:"fieldName"`
// 	MessageCode string `json:"messageCode"`
// }
type DocumentRequest struct {
	CustomersId int64  `json:"customersId"`
	SiteUsersId int64  `json:"siteUsersId"`
	DocumentId  int64  `json:"documentId"`
	AddedBy     string `json:"addedBy"`
}
type RejectedDocuments struct {
	BDocumentsUpload         bool    `json:"bDocumentsUpload"`
	FinancialInstitutionForm *string `json:"financialInstitutionForm"` // null possible
}

// "details" object
type BusinessVerificationDetails struct {
	BusinessVerificationStatus   *string           `json:"businessVerificationStatus"` // null
	BusinessVerificationStep     string            `json:"businessVerificationStep"`
	BusinessVerificationNextStep string            `json:"businessVerificationNextStep"`
	RawRejectedDocuments         string            `json:"rawRejectedDocuments"` // JSON string from DB
	RejectedDocuments            RejectedDocuments `json:"rejectedDocuments"`
}

// Final RPC / API response
type BusinessVerificationStatusResult struct {
	ID      int64                       `json:"id"`
	Details BusinessVerificationDetails `json:"details"`
	Status  string                      `json:"status"`
	Errors  []ErrorItem                 `json:"errors"`
}

// common "label/value" items (entityTypes, businessActivities, etc.)
type LabelValue struct {
	Label string `json:"label"`
	Value string `json:"value"` // API me "1000" string hai
}

type StateProvince struct {
	ID                int    `json:"id"`
	StateProvinceName string `json:"stateProvinceName"`
	AnsiCode          string `json:"ansiCode"`
}

type AvailableCountry struct {
	ID          int             `json:"id"`
	CountryName string          `json:"countryName"`
	IsoCode     string          `json:"isoCode"`
	States      []StateProvince `json:"states"`
}

type OpenedDocument struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	InternalName        string `json:"internalName"`
	OrderNumber         int    `json:"orderNumber"`
	BAdditionalDocument bool   `json:"bAdditionalDocument"`
}

type BusinessRegistrationMetadataDetails struct {
	EntityTypes            []LabelValue       `json:"entityTypes"`
	AvailableCountries     []AvailableCountry `json:"availableCountries"`
	BusinessActivities     []LabelValue       `json:"businessActivities"`
	FundingSources         []LabelValue       `json:"fundingSources"`
	TransactionVolumes     []LabelValue       `json:"transactionVolumes"`
	MonthlyUsdEquivalents  []LabelValue       `json:"monthlyUsdEquivalents"`
	UboReportingExemptions []LabelValue       `json:"uboReportingExemptions"`
	RegionsOfOperation     []LabelValue       `json:"regionsOfOperation"`
	OpenedDocumentsList    []OpenedDocument   `json:"openedDocumentsList"`
}

type BusinessRegistrationMetadataResult struct {
	ID      int64                               `json:"id"`
	Details BusinessRegistrationMetadataDetails `json:"details"`
	Status  string                              `json:"status"`
	Errors  []ErrorItem                         `json:"errors"`
}
type ErrorItem struct {
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}
