package audit

type Auditor interface {
	ID() AuditorID
	Process(payload Payload)
}

type Auditable interface {
	Audit(auditor Auditor)
	Neglect(auditorID AuditorID)
}

type AuditorID string
type Payload struct {
	Timestamp string   `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
