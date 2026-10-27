package model

import c "github.com/sihuangtech/probelune-contracts"

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password_hash,omitempty"`
	Role     string `json:"role"`
	Disabled bool   `json:"disabled"`
}
type Session struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Expires int64  `json:"expires"`
	Device  string `json:"device"`
}
type Pair struct {
	UserID  string `json:"user_id"`
	Expires int64  `json:"expires"`
}
type Node struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Region          string   `json:"region"`
	Network         string   `json:"network"`
	TokenHash       string   `json:"token_hash,omitempty"`
	RegisterHash    string   `json:"register_hash,omitempty"`
	RegisterExpires int64    `json:"register_expires"`
	Enabled         bool     `json:"enabled"`
	Revoked         bool     `json:"revoked"`
	LastSeen        int64    `json:"last_seen"`
	CreatedAt       int64    `json:"created_at"`
	Policy          c.Policy `json:"policy"`
	Info            c.Poll   `json:"info"`
	ClockSkew       int64    `json:"clock_skew"`
}
type State struct {
	ID        string   `json:"id"`
	MonitorID string   `json:"monitor_id"`
	NodeID    string   `json:"node_id"`
	Version   int      `json:"version"`
	Status    string   `json:"status"`
	Failures  int      `json:"failures"`
	Successes int      `json:"successes"`
	LastAt    int64    `json:"last_at"`
	Result    c.Result `json:"result"`
}
type Event struct {
	At      int64  `json:"at"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
	User    string `json:"user"`
}
type Incident struct {
	ID            string   `json:"id"`
	MonitorID     string   `json:"monitor_id"`
	NodeID        string   `json:"node_id"`
	Name          string   `json:"name"`
	OpenedAt      int64    `json:"opened_at"`
	RecoveredAt   int64    `json:"recovered_at"`
	AckBy         string   `json:"ack_by"`
	SilencedUntil int64    `json:"silenced_until"`
	Affected      []string `json:"affected"`
	Events        []Event  `json:"events"`
	Suppressed    bool     `json:"suppressed"`
}
type Maintenance struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	MonitorIDs []string `json:"monitor_ids"`
	Start      int64    `json:"start"`
	End        int64    `json:"end"`
}
type Channel struct {
	ReminderMinutes int      `json:"reminder_minutes"`
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Enabled         bool     `json:"enabled"`
	URL             string   `json:"url"`
	Secret          string   `json:"secret,omitempty"`
	SMTPHost        string   `json:"smtp_host"`
	SMTPPort        int      `json:"smtp_port"`
	Username        string   `json:"username"`
	Password        string   `json:"password,omitempty"`
	From            string   `json:"from"`
	To              string   `json:"to"`
	TLSMode         string   `json:"tls_mode"`
	Policy          c.Policy `json:"policy"`
}
type Delivery struct {
	ID         string `json:"id"`
	ChannelID  string `json:"channel_id"`
	IncidentID string `json:"incident_id"`
	Kind       string `json:"kind"`
	Status     string `json:"status"`
	Attempts   int    `json:"attempts"`
	NextAt     int64  `json:"next_at"`
	Error      string `json:"error"`
	CreatedAt  int64  `json:"created_at"`
}
type Audit struct {
	ID       string `json:"id"`
	At       int64  `json:"at"`
	User     string `json:"user"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
}
type Device struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Token     string `json:"token,omitempty"`
	Platform  string `json:"platform"`
	CreatedAt int64  `json:"created_at"`
}
type Settings struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	PublicURL    string `json:"public_url"`
	Timezone     string `json:"timezone"`
	RawDays      int    `json:"raw_days"`
	EventDays    int    `json:"event_days"`
	PublicStatus bool   `json:"public_status"`
}
