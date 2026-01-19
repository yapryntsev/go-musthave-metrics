package audit

type (
	// Auditor listens for incoming events and processes them.
	Auditor interface {
		// ID returns auditor's unique ID.
		ID() AuditorID
		// Process gets the payload from the auditable object and processes it.
		Process(payload Payload)
	}

	// Auditable notifies registered auditors about certain events.
	Auditable interface {
		// Audit register a new auditor. When new events occur, auditor will be notified.
		Audit(auditor Auditor)
		// Neglect remove the auditor with the provided ID from the notification list.
		Neglect(auditorID AuditorID)
	}

	AuditorID string

	// Payload contains info about specific event.
	Payload struct {
		Timestamp string   `json:"ts"`
		Metrics   []string `json:"metrics"`
		IPAddress string   `json:"ip_address"`
	}
)
