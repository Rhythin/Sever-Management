package packets

import "github.com/rhythin/sever-management/internal/domain"

type ProvisionRequest struct {
	Region string `json:"region"`
	Type   string `json:"type"`
}
type ProvisionResponse struct {
	ID string `json:"id"`
}
type ServerResponse struct {
	ID        string           `json:"id"`
	State     string           `json:"state"`
	Region    string           `json:"region,omitempty"`
	Type      string           `json:"type,omitempty"`
	Billing   *BillingResponse `json:"billing,omitempty"`
	IPAddress string           `json:"ip_address,omitempty"`
}

type BillingResponse struct {
	AccumulatedSeconds int64   `json:"accumulated_seconds"`
	LastBilledAt       *string `json:"last_billed_at,omitempty"`
	TotalCost          float64 `json:"total_cost"`
}

type ActionRequest struct {
	Action domain.ServerAction `json:"action"` // must be one of start|stop|reboot|terminate
}
type ActionResponse struct {
	Result string `json:"result"`
}
type EventLogResponse struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Message   string `json:"message"`
}
