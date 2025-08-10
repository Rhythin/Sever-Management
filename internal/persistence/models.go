package persistence

import (
	"time"
)

// Server represents a virtual server instance in the DB

type Server struct {
	ID           string `gorm:"primaryKey;type:text"`
	Region       string
	Type         string
	IPID         *uint // Foreign key to IPAddress
	IP           *IPAddress
	State        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    *time.Time
	StoppedAt    *time.Time
	TerminatedAt *time.Time
	Billing      *Billing    `gorm:"foreignKey:ServerID"`
	Events       []*EventLog `gorm:"foreignKey:ServerID"`
}

// TableName specifies the table name for Server
func (Server) TableName() string {
	return "servers"
}

// IPAddress tracks allocated IPs

type IPAddress struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Address   string `gorm:"uniqueIndex"`
	Allocated bool
	ServerID  *string // Nullable, FK to Server
}

// TableName specifies the table name for IPAddress
func (IPAddress) TableName() string {
	return "ip_addresses"
}

// Billing tracks server cost and uptime

type Billing struct {
	ID                 uint   `gorm:"primaryKey;autoIncrement"`
	ServerID           string `gorm:"uniqueIndex"`
	AccumulatedSeconds int64
	LastBilledAt       *time.Time
	TotalCost          float64
}

// TableName specifies the table name for Billing
func (Billing) TableName() string {
	return "billing"
}

// EventLog stores server lifecycle events

type EventLog struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	ServerID  string `gorm:"index"`
	Timestamp time.Time
	Type      string
	Message   string
}

// TableName specifies the table name for EventLog
func (EventLog) TableName() string {
	return "event_logs"
}
