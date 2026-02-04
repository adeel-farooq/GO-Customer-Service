package profile

type GetUserInfoRequest struct {
	SiteUsersId int64 `json:"siteUsersId"`
	CustomersId int64 `json:"customersId"`
}

// gob-safe: concrete types on both sides
type GetUserInfoResult struct {
	ID      int64       `json:"id"`
	Details GetUserInfo `json:"details"`
	Status  string      `json:"status"`
	Errors  []string    `json:"errors"`
}

type ErrorResultDto struct {
	ErrorType   string `json:"errorType"`
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}

type GetUserInfo struct {
	ID                             int    `json:"id"`
	FirstName                      string `json:"firstName"`
	LastName                       string `json:"lastName"`
	AccountType                    string `json:"accountType"`
	BAccountSwitchAvailable        bool   `json:"bAccountSwitchAvailable"`
	BRequiresApprovalToCreatePayee bool   `json:"bRequiresApprovalToCreatePayee"`
	BMainAccountHolder             bool   `json:"bMainAccountHolder"`
	CustomerUsersCustomersId       int    `json:"customerUsersCustomersId"`
	OnboardingVersion              string `json:"onboardingVersion"`
	OtcDeskVersion                 string `json:"otcDeskVersion"`
	BCanBulkTransfer               bool   `json:"bCanBulkTransfer"`
	BCanReverseTransfer            bool   `json:"bCanReverseTransfer"`
	BFinancialInstitution          bool   `json:"bFinancialInstitution"`
	BPushToCardEnabled             bool   `json:"bPushToCardEnabled"`
	BExtendedUserList              bool   `json:"bExtendedUserList"`
	BQuickCryptoSaleEnabled        bool   `json:"bQuickCryptoSaleEnabled"`
	BNetworkDiscoverable           bool   `json:"bNetworkDiscoverable"`
	NetworkID                      string `json:"networkID"`
	ProductServices                string `json:"productServices"`
}
