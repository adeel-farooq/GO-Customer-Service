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

// ---- Typed result model (camelCase keys) ----
// This matches the gateway-facing JSON shape you shared.
type GetVerificationFormResult struct {
	ID      int                        `json:"id"`
	Details GetVerificationFormDetails `json:"details"`
	Status  string                     `json:"status"`
	Errors  []ErrorItem                `json:"errors"`
}

type GetVerificationFormDetails struct {
	FormData        VerificationFormData `json:"formData"`
	InnerErrors     string               `json:"innerErrors"`
	ListInnerErrors []ErrorItem          `json:"listInnerErrors"`
}

// ===== FORM DATA =====
type VerificationFormData struct {
	Test                         any    `json:"test"`
	BusinessVerificationNextStep string `json:"businessVerificationNextStep"`
	BusinessVerificationStep     string `json:"businessVerificationStep"`

	RegistrationInformation RegistrationInformation `json:"registrationInformation"`
	OperationsInformation   OperationsInformation   `json:"operationsInformation"`
	OwnerInformation        OwnerInformation        `json:"ownerInformation"`
	Terms                   Terms                   `json:"terms"`
	DocumentsUpload         []DocumentsUploadItem   `json:"documentsUpload"`
}

type RegistrationInformation struct {
	EntityName                         string  `json:"entityName"`
	EntityTypesID                      int     `json:"entityTypesID"`
	EntityTypesOther                   *string `json:"entityTypesOther"`
	RegistrationDate                   string  `json:"registrationDate"`
	RegistrationNumber                 string  `json:"registrationNumber"`
	OfficialBusinessRegistrationNumber string  `json:"officialBusinessRegistrationNumber"`
	TaxNumber                          string  `json:"taxNumber"`
	DoingBusinessAs                    string  `json:"doingBusinessAs"`

	AddressStreet   string `json:"addressStreet"`
	AddressNumber   string `json:"addressNumber"`
	AddressPostCode string `json:"addressPostCode"`
	AddressCity     string `json:"addressCity"`
	AddressState    string `json:"addressState"`
	AddressCountry  string `json:"addressCountry"`

	OperatingAddressStreet   *string `json:"operatingAddressStreet"`
	OperatingAddressNumber   *string `json:"operatingAddressNumber"`
	OperatingAddressPostCode *string `json:"operatingAddressPostCode"`
	OperatingAddressCity     *string `json:"operatingAddressCity"`
	OperatingAddressState    *string `json:"operatingAddressState"`
	OperatingAddressCountry  *string `json:"operatingAddressCountry"`
}

type OperationsInformation struct {
	WebAddress         string `json:"webAddress"`
	PhoneNumber        string `json:"phoneNumber"`
	SupportEmail       string `json:"supportEmail"`
	BPubliclyListed    string `json:"bPubliclyListed"`
	Ticker             string `json:"ticker"`
	Exchanges          string `json:"exchanges"`
	OperationRegionIds []int  `json:"operationRegionIds"`
	BLicenseRequired   string `json:"bLicenseRequired"`

	AdditionalLicensingInfo *string `json:"additionalLicensingInfo"`
	PrimaryRegulator        *string `json:"primaryRegulator"`
	LicenseNumber           *string `json:"licenseNumber"`

	FinancialInstitutionFormFileName                     *string `json:"financialInstitutionFormFileName"`
	FinancialInstitutionFormCustomersBusinessDocumentsId int     `json:"financialInstitutionFormCustomersBusinessDocumentsId"`

	BusinessActivityIds []int `json:"businessActivityIds"`

	Web3ChainsAndAddresses any     `json:"web3ChainsAndAddresses"`
	SourceOfFundsIds       []int   `json:"sourceOfFundsIds"`
	SourceOfFundsOther     *string `json:"sourceOfFundsOther"`
	ActiveBanks            any     `json:"activeBanks"`

	YearlyTransactionsId *int `json:"yearlyTransactionsId"`
	MonthlyUsdValueId    *int `json:"monthlyUsdValueId"`

	BusinessAndIndustryType string `json:"businessAndIndustryType"`
	CryptoCurrencyOffered   string `json:"cryptoCurrencyOffered"`

	URLS string `json:"urLs"`

	ProcessingTraffic string `json:"processingTraffic"`
	WalletAddresses   string `json:"walletAddresses"`

	DirectorShareholderLocation     string `json:"director_ShareholderLocation"`
	CompanyAuthorizedRepresentative string `json:"companyAuthorizedRepresentative"`

	BCardPayment string `json:"bCardPayment"`
	ProviderName string `json:"providerName"`

	ProjectedAnnualTurnover float64 `json:"projectedAnnualTurnover"`
	AverageTransactionValue float64 `json:"averageTransactionValue"`
	MinTransactionValues    float64 `json:"minTransactionValues"`
	MaxTransactionValues    float64 `json:"maxTransactionValues"`
	ChargeBack              float64 `json:"chargeBack"`
	RefundValues            float64 `json:"refundValues"`

	BOwnTokenSale               string `json:"bOwnTokenSale"`
	CryptoTechnicalArchitecture string `json:"cryptoTechnicalArchitecture"`
	FormalRegulatory            string `json:"formalRegulatory"`
	BusinessPlan                string `json:"businessPlan"`
	BNFTSale                    string `json:"bNFTSale"`
	ExecutiveSummaryCrypto      string `json:"executiveSummaryCrypto"`
	ExecutiveSummaryNFT         string `json:"executiveSummaryNFT"`
	NFTDynamic                  string `json:"nftDynamic"`
	NFTCustody                  string `json:"nftCustody"`
	IsPortlProduct              bool   `json:"isPortlProduct"`

	BFundsReceivedUSA *string `json:"bFundsReceivedUSA"`
	BFundsSentUSA     *string `json:"bFundsSentUSA"`

	PurposeOfAccountId    *int   `json:"purposeOfAccountId"`
	PurposeOfAccountOther string `json:"purposeOfAccountOther"`

	AnticipatedMonthlyTransactionCountId int `json:"anticipatedMonthlyTransactionCountId"`
	AnticipatedMonthlyVolumeUsdId        int `json:"anticipatedMonthlyVolumeUsdId"`

	CountriesSendingPaymentsTo     string `json:"countriesSendingPaymentsTo"`
	CountriesReceivingPaymentsFrom string `json:"countriesReceivingPaymentsFrom"`

	AnticipatedMonthlyCryptoTransactionsId *int `json:"anticipatedMonthlyCryptoTransactionsId"`
	AnticipatedMonthlyCryptoVolumeUsdId    *int `json:"anticipatedMonthlyCryptoVolumeUsdId"`

	IntegrationMethodId               *int `json:"integrationMethodId"`
	ExpectedMonthlyPaymentVolumeUsdId *int `json:"expectedMonthlyPaymentVolumeUsdId"`
	SettlementPreferenceId            *int `json:"settlementPreferenceId"`

	CustomerTrafficCountries any `json:"customerTrafficCountries"`
	HistoricalChargebackRate any `json:"historicalChargebackRate"`
	HistoricalRefundRate     any `json:"historicalRefundRate"`
	ProductsServicesSold     any `json:"productsServicesSold"`
}

type OwnerInformation struct {
	ExemptionId int `json:"exemptionId"`

	BusinessBeneficialOwners   []any `json:"businessBeneficialOwners"`
	IndividualBeneficialOwners []any `json:"individualBeneficialOwners"`
	BeneficialOwnersStructure  []any `json:"beneficialOwnersStructure"`

	AuthorizedSigner string `json:"authorizedSigner"`
	BConfirmAboveTen any    `json:"bConfirmAboveTen"`
}

type Terms struct {
	BTaxAcknowledgement      bool `json:"bTaxAcknowledgement"`
	BFatcaAcknowledgement    bool `json:"bFatcaAcknowledgement"`
	BConfirmTrueAndCorrect   bool `json:"bConfirmTrueAndCorrect"`
	BEConsentDisclosure      bool `json:"bEConsentDisclosure"`
	BDepositAccountAgreement bool `json:"bDepositAccountAgreement"`
	BUSAPatriotActNotice     bool `json:"bUSAPatriotActNotice"`
	BAcceptTerms             bool `json:"bAcceptTerms"`
}

type DocumentsUploadItem struct {
	DocumentTypesID int    `json:"documentTypesID"`
	FileName        string `json:"fileName"`
}

// ---- Typed model for rawFormData JSON (so numbers/bools/arrays stay typed) ----
// NOTE: Names are prefixed to avoid clashing with existing request structs.

// ---- root of rawFormData JSON ----
type BusinessFormData struct {
	Test                         any    `json:"Test"` // null possible, unknown type
	BusinessVerificationNextStep string `json:"BusinessVerificationNextStep"`
	BusinessVerificationStep     string `json:"BusinessVerificationStep"`

	RegistrationInformation BusinessFormRegistrationInformation `json:"RegistrationInformation"`
	OperationsInformation   BusinessFormOperationsInformation   `json:"OperationsInformation"`
	OwnerInformation        BusinessFormOwnerInformation        `json:"OwnerInformation"`
	Terms                   BusinessFormTerms                   `json:"Terms"`
	DocumentsUpload         []BusinessFormDocumentsUploadItem   `json:"DocumentsUpload"`
}

// ---- RegistrationInformation ----
type BusinessFormRegistrationInformation struct {
	EntityName                         string  `json:"EntityName"`
	EntityTypesID                      int     `json:"EntityTypesID"`
	EntityTypesOther                   *string `json:"EntityTypesOther"`
	RegistrationDate                   string  `json:"RegistrationDate"` // keep string (ISO)
	RegistrationNumber                 string  `json:"RegistrationNumber"`
	OfficialBusinessRegistrationNumber string  `json:"OfficialBusinessRegistrationNumber"`
	TaxNumber                          string  `json:"TaxNumber"`
	DoingBusinessAs                    string  `json:"DoingBusinessAs"`

	AddressStreet   string `json:"AddressStreet"`
	AddressNumber   string `json:"AddressNumber"`
	AddressPostCode string `json:"AddressPostCode"`
	AddressCity     string `json:"AddressCity"`
	AddressState    string `json:"AddressState"`
	AddressCountry  string `json:"AddressCountry"`

	OperatingAddressStreet   *string `json:"OperatingAddressStreet"`
	OperatingAddressNumber   *string `json:"OperatingAddressNumber"`
	OperatingAddressPostCode *string `json:"OperatingAddressPostCode"`
	OperatingAddressCity     *string `json:"OperatingAddressCity"`
	OperatingAddressState    *string `json:"OperatingAddressState"`
	OperatingAddressCountry  *string `json:"OperatingAddressCountry"`
}

// ---- OperationsInformation ----
type BusinessFormOperationsInformation struct {
	WebAddress      string `json:"WebAddress"`
	PhoneNumber     string `json:"PhoneNumber"`
	SupportEmail    string `json:"SupportEmail"`
	BPubliclyListed string `json:"bPubliclyListed"` // "Yes"/"No"
	Ticker          string `json:"Ticker"`
	Exchanges       string `json:"Exchanges"`

	OperationRegionIds []int  `json:"OperationRegionIds"`
	BLicenseRequired   string `json:"bLicenseRequired"` // "Yes"/"No"

	AdditionalLicensingInfo *string `json:"AdditionalLicensingInfo"`
	PrimaryRegulator        *string `json:"PrimaryRegulator"`
	LicenseNumber           *string `json:"LicenseNumber"`

	FinancialInstitutionFormFileName                     *string `json:"FinancialInstitutionFormFileName"`
	FinancialInstitutionFormCustomersBusinessDocumentsId int     `json:"FinancialInstitutionFormCustomersBusinessDocumentsId"`

	BusinessActivityIds []int `json:"BusinessActivityIds"`

	Web3ChainsAndAddresses any     `json:"Web3ChainsAndAddresses"`
	SourceOfFundsIds       []int   `json:"SourceOfFundsIds"`
	SourceOfFundsOther     *string `json:"SourceOfFundsOther"`
	ActiveBanks            any     `json:"ActiveBanks"`

	YearlyTransactionsId *int `json:"YearlyTransactionsId"`
	MonthlyUsdValueId    *int `json:"MonthlyUsdValueId"`

	BusinessAndIndustryType string `json:"BusinessAndIndustryType"`
	CryptoCurrencyOffered   string `json:"CryptoCurrencyOffered"`

	URLs              string `json:"URLs"` // IMPORTANT: raw json me "URLs" hai
	ProcessingTraffic string `json:"ProcessingTraffic"`
	WalletAddresses   string `json:"WalletAddresses"`

	Director_ShareholderLocation    string `json:"Director_ShareholderLocation"`
	CompanyAuthorizedRepresentative string `json:"CompanyAuthorizedRepresentative"`

	BCardPayment string `json:"bCardPayment"` // "Yes"/"No"
	ProviderName string `json:"ProviderName"`

	ProjectedAnnualTurnover float64 `json:"ProjectedAnnualTurnover"`
	AverageTransactionValue float64 `json:"AverageTransactionValue"`
	MinTransactionValues    float64 `json:"MinTransactionValues"`
	MaxTransactionValues    float64 `json:"MaxTransactionValues"`
	ChargeBack              float64 `json:"ChargeBack"`
	RefundValues            float64 `json:"RefundValues"`

	BOwnTokenSale               string `json:"bOwnTokenSale"` // "Yes"/"No"
	CryptoTechnicalArchitecture string `json:"CryptoTechnicalArchitecture"`
	FormalRegulatory            string `json:"FormalRegulatory"`
	BusinessPlan                string `json:"BusinessPlan"`
	BNFTSale                    string `json:"bNFTSale"` // "Yes"/"No"
	ExecutiveSummaryCrypto      string `json:"ExecutiveSummaryCrypto"`
	ExecutiveSummaryNFT         string `json:"ExecutiveSummaryNFT"`
	NFTDynamic                  string `json:"NFTDynamic"`
	NFTCustody                  string `json:"NFTCustody"`
	IsPortlProduct              bool   `json:"IsPortlProduct"`

	BFundsReceivedUSA *string `json:"bFundsReceivedUSA"`
	BFundsSentUSA     *string `json:"bFundsSentUSA"`

	PurposeOfAccountId    *int   `json:"PurposeOfAccountId"`
	PurposeOfAccountOther string `json:"PurposeOfAccountOther"`

	AnticipatedMonthlyTransactionCountId int `json:"AnticipatedMonthlyTransactionCountId"`
	AnticipatedMonthlyVolumeUsdId        int `json:"AnticipatedMonthlyVolumeUsdId"`

	CountriesSendingPaymentsTo     string `json:"CountriesSendingPaymentsTo"`
	CountriesReceivingPaymentsFrom string `json:"CountriesReceivingPaymentsFrom"`

	AnticipatedMonthlyCryptoTransactionsId *int `json:"AnticipatedMonthlyCryptoTransactionsId"`
	AnticipatedMonthlyCryptoVolumeUsdId    *int `json:"AnticipatedMonthlyCryptoVolumeUsdId"`

	IntegrationMethodId               *int `json:"IntegrationMethodId"`
	ExpectedMonthlyPaymentVolumeUsdId *int `json:"ExpectedMonthlyPaymentVolumeUsdId"`
	SettlementPreferenceId            *int `json:"SettlementPreferenceId"`

	CustomerTrafficCountries any `json:"CustomerTrafficCountries"`
	HistoricalChargebackRate any `json:"HistoricalChargebackRate"`
	HistoricalRefundRate     any `json:"HistoricalRefundRate"`
	ProductsServicesSold     any `json:"ProductsServicesSold"`
}

// ---- OwnerInformation ----
type BusinessFormOwnerInformation struct {
	ExemptionId int `json:"ExemptionId"`

	BusinessBeneficialOwners   []any `json:"BusinessBeneficialOwners"`
	IndividualBeneficialOwners []any `json:"IndividualBeneficialOwners"`
	BeneficialOwnersStructure  []any `json:"BeneficialOwnersStructure"`

	AuthorizedSigner string `json:"AuthorizedSigner"`
	BConfirmAboveTen any    `json:"bConfirmAboveTen"` // null possible
}

// ---- Terms ----
type BusinessFormTerms struct {
	BTaxAcknowledgement      bool `json:"bTaxAcknowledgement"`
	BFatcaAcknowledgement    bool `json:"bFatcaAcknowledgement"`
	BConfirmTrueAndCorrect   bool `json:"bConfirmTrueAndCorrect"`
	BEConsentDisclosure      bool `json:"bEConsentDisclosure"`
	BDepositAccountAgreement bool `json:"bDepositAccountAgreement"`
	BUSAPatriotActNotice     bool `json:"bUSAPatriotActNotice"`
	BAcceptTerms             bool `json:"bAcceptTerms"`
}

// ---- DocumentsUpload ----
type BusinessFormDocumentsUploadItem struct {
	DocumentTypesID int    `json:"DocumentTypesID"`
	FileName        string `json:"FileName"`
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

// Command payload shape coming from gateway for SaveVerificationForm.
// We keep sub-objects as RawMessage so we can forward them to the SP without
// type coercion (e.g. some IDs come as strings in the payload).
type SaveVerificationFormCommandPayload struct {
	BusinessVerificationStep string          `json:"businessVerificationStep"`
	RegistrationInformation  json.RawMessage `json:"registrationInformation"`
	OperationsInformation    json.RawMessage `json:"operationsInformation"`
	OwnerInformation         json.RawMessage `json:"ownerInformation"`
	Terms                    json.RawMessage `json:"terms"`
	AccountCurrencySelection json.RawMessage `json:"accountCurrencySelection"`
	DocumentsUpload          json.RawMessage `json:"documentsUpload"`
	AddedBy                  string          `json:"addedBy,omitempty"`
}

// Minimal parts we need for missing files + ownership graph
type VerificationForm struct {
	BusinessVerificationStep string                        `json:"businessVerificationStep"`
	OwnerInformation         *VerificationOwnerInformation `json:"ownerInformation,omitempty"`
	RegistrationInformation  json.RawMessage               `json:"registrationInformation,omitempty"`
	OperationsInformation    *OperationsInfo               `json:"operationsInformation,omitempty"`
}

type VerificationOwnerInformation struct {
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
type SaveVerificationFormResponse struct {
	CustomerId                   int64              `json:"customerId"`
	BusinessVerificationNextStep string             `json:"businessVerificationNextStep"`
	BusinessVerificationStep     string             `json:"businessVerificationStep"`
	SignificantParties           []SignificantParty `json:"significantParties"`
	CompanyName                  *string            `json:"companyName"`
	InnerErrors                  string             `json:"innerErrors"`
	ListInnerErrors              []ErrorItem        `json:"listInnerErrors"`
}

// Gateway-facing wrapper for SaveVerificationForm (matches .NET DbResultDto style)
type SaveVerificationFormResult struct {
	ID      int64                        `json:"id"`
	Details SaveVerificationFormResponse `json:"details"`
	Status  string                       `json:"status"`
	Errors  []ErrorItem                  `json:"errors"`
}

type BusinessDocumentsResult struct {
	ID      int                     `json:"id"`
	Details []BusinessDocumentGroup `json:"details"`
	Status  string                  `json:"status"`
	Errors  []ErrorItem             `json:"errors"`
}

// ===== Details: category/group =====

type BusinessDocumentGroup struct {
	CategoryKey  string             `json:"categoryKey"`
	CategoryName string             `json:"categoryName"`
	Order        int                `json:"order"`
	Documents    []BusinessDocument `json:"documents"`
}

// ===== Individual document =====

type BusinessDocument struct {
	DocumentId                   int     `json:"documentId"`
	Name                         string  `json:"name"`
	OrderNumber                  int     `json:"orderNumber"`
	DocumentStatus               int     `json:"documentStatus"`
	DocumentStatusString         string  `json:"documentStatusString"`
	BAlwaysRequired              bool    `json:"bAlwaysRequired"`
	DocumentNotes                string  `json:"documentNotes"`
	UploadNotes                  *string `json:"uploadNotes"`     // null possible
	RejectionReason              *string `json:"rejectionReason"` // null possible
	BIsAdditionalDocument        bool    `json:"bIsAdditionalDocument"`
	CustomersBusinessDocumentsId *int    `json:"customersBusinessDocumentsId"`
	Filename                     *string `json:"filename"`
	DocumentCategory             string  `json:"documentCategory"`
}

// input (flat list)
type FlatBusinessDocument struct {
	DocumentId                   int     `json:"DocumentId"`
	Name                         string  `json:"Name"`
	OrderNumber                  int     `json:"OrderNumber"`
	DocumentStatus               int     `json:"DocumentStatus"`
	DocumentStatusString         string  `json:"DocumentStatusString"`
	BAlwaysRequired              bool    `json:"bAlwaysRequired"`
	DocumentNotes                string  `json:"DocumentNotes"`
	Filename                     *string `json:"Filename,omitempty"`
	CustomersBusinessDocumentsId *int    `json:"CustomersBusinessDocumentsId,omitempty"`
	BIsAdditionalDocument        bool    `json:"bIsAdditionalDocument"`
	DocumentCategory             string  `json:"documentCategory"`
}
