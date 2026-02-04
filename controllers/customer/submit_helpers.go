package customer

func ShouldSendSignificantPartyEmails(d *SubmitDetails) bool {
	if d == nil {
		return false
	}
	if d.ListInnerErrors == nil {
		return true
	}
	return len(d.ListInnerErrors) == 0
}
