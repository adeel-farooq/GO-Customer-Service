package customer

var TokenHelper SignificantPartyTokenHelper = HMACTokenHelper{Secret: "dev-secret-change-me"}
var NotificationsClient SignificantPartyNotifier = NoopNotifier{}
