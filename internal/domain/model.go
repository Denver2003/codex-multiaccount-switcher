package domain

type Metadata struct {
	SchemaVersion    int              `json:"schema_version"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
	LastSwitchAt     string           `json:"last_switch_at,omitempty"`
	CurrentProfileID string           `json:"current_profile_id,omitempty"`
	Profiles         []ProfileSummary `json:"profiles"`
}

type ProfileSummary struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Email      string `json:"email,omitempty"`
	CreatedAt  string `json:"created_at"`
	LastUsedAt string `json:"last_used_at,omitempty"`
	AuthHash   string `json:"auth_hash"`
}
