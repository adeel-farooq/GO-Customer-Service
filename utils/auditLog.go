package utils

import (
	"fmt"
	"go-cloud-customer/db"
	"go-cloud-customer/services"
)

func InsertAuditLog(req services.InsertAuditLogRequest) int64 {
	var id int64
	query := `INSERT INTO audit_logs (
				customer_id, partner_id, user_id, user_role, action_type,
				entity_name, entity_id, new_data, 
				response_status, error_message, ip_address, user_agent
				) VALUES (
					$1, $2, $3, $4, $5, 
					$6, $7, $8, $9, $10, 
					$11, $12 
				)
				RETURNING id
				`
	err := db.DB.QueryRow(query,
		req.CustomerId, 15, req.UserId, req.UserRole, req.ActionType,
		req.EntityName, req.EntityId, string(req.RequestBody),
		"", "", req.IpAddress, req.UserAgent,
	).Scan(&id)
	if err != nil {
		fmt.Printf("Failed to insert audit log %v", err)
		return 0
	}

	return id
}

func UpdateAuditLog(id int64, response_status, error_message, api_url string) {
	query := `UPDATE audit_logs SET 
				response_status = $1, 
				error_message   = $2,
				api_url			= $3
			  WHERE id = $4
				`
	_, err := db.DB.Exec(query,
		response_status,
		error_message,
		api_url,
		id,
	)
	if err != nil {
		fmt.Printf("Failed to insert audit log %v", err)
	}
}
