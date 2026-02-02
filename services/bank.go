package services

/* Centralized CustomerService struct */
type CustomerService struct{}

/* Define a HealthCheckService for checking server health */
type HealthCheckService struct{}

/* HealthCheck method to check server status */
func (h *HealthCheckService) Ping(args *string, reply *string) error {
	*reply = "Server is healthy"
	return nil
}

type InsertAuditLogRequest struct {
	CustomerId  int64  `json:"customer_id,omitempty"`
	UserId      int64  `json:"user_id,omitempty"`
	UserRole    string `json:"user_role,omitempty"`
	ActionType  string `json:"action_type,omitempty"`
	EntityId    string `json:"entity_id,omitempty"`
	EntityName  string `json:"entity_name,omitempty"`
	IpAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	RequestBody string `json:"request_body,omitempty"`
}
