package customer

type SubmitDetails struct {
	CompanyName        string             `json:"companyName"`
	ListInnerErrors    []any              `json:"listInnerErrors"`
	SignificantParties []SignificantParty `json:"significantParties"`
}

type SignificantParty struct {
	Id           int64  `json:"id"`
	EmailAddress string `json:"emailAddress"`
}
