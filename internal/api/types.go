package api

import "time"

type AgentOTPRequest struct {
	Phone          string `json:"phone"`
	ClientInstance string `json:"client_instance"`
}

type AgentOTPVerificationRequest struct {
	RequestID      string `json:"request_id"`
	Code           string `json:"code"`
	ClientInstance string `json:"client_instance"`
}

type AgentRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type OTPRequestAccepted struct {
	RequestID       string `json:"request_id"`
	Status          string `json:"status"`
	CooldownSeconds int    `json:"cooldown_seconds"`
}

type AgentSessionResponse struct {
	AccessToken             string         `json:"access_token"`
	AccessExpiresInSeconds  int            `json:"access_expires_in_seconds"`
	RefreshToken            string         `json:"refresh_token,omitempty"`
	RefreshExpiresInSeconds int            `json:"refresh_expires_in_seconds,omitempty"`
	Account                 CurrentAccount `json:"account"`
	Scopes                  []string       `json:"scopes"`
}

type CurrentAccount struct {
	ID                   string `json:"id"`
	Status               string `json:"status"`
	RegistrationPlatform string `json:"registration_platform"`
	PhoneMasked          string `json:"phone_masked"`
	StateVersion         int64  `json:"state_version"`
}

type AgentRecorderCardPage struct {
	Items []AgentRecorderCard `json:"items"`
}

type AgentRecorderCard struct {
	ID                string    `json:"id"`
	BindingEpochID    string    `json:"binding_epoch_id"`
	IdentityMasked    string    `json:"identity_masked"`
	BindingStatus     string    `json:"binding_status"`
	CredentialVersion int64     `json:"credential_version,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type AgentQuotaSummary struct {
	AvailableTokens           int64      `json:"available_tokens"`
	GiftAvailableTokens       int64      `json:"gift_available_tokens"`
	PaidAvailableTokens       int64      `json:"paid_available_tokens"`
	AdjustmentAvailableTokens int64      `json:"adjustment_available_tokens"`
	NextExpiresAt             *time.Time `json:"next_expires_at,omitempty"`
	CalculatedAt              time.Time  `json:"calculated_at"`
}

type AgentModelUsageSummary struct {
	Period                            string    `json:"period"`
	PeriodStartsAt                    time.Time `json:"period_starts_at"`
	PeriodEndsAt                      time.Time `json:"period_ends_at"`
	ProviderTotalTokens               int64     `json:"provider_total_tokens"`
	ChargedQuotaTokens                int64     `json:"charged_quota_tokens"`
	RunCount                          int64     `json:"run_count"`
	AttemptCount                      int64     `json:"attempt_count"`
	GiftChargedTokens                 int64     `json:"gift_charged_tokens"`
	PaidChargedTokens                 int64     `json:"paid_charged_tokens"`
	EstimatedPlatformSavingsMicrounit int64     `json:"estimated_platform_savings_microunits,omitempty"`
	ValueBasis                        string    `json:"value_basis"`
	CalculatedAt                      time.Time `json:"calculated_at"`
}

type AgentMemoryPage struct {
	Items      []AgentMemory `json:"items"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

type AgentMemory struct {
	ID                    string    `json:"id"`
	Platform              string    `json:"platform"`
	ObjectType            string    `json:"object_type"`
	ClientObjectID        string    `json:"client_object_id"`
	RevisionID            string    `json:"revision_id"`
	Revision              int64     `json:"revision"`
	SchemaID              string    `json:"schema_id"`
	SchemaVersion         int64     `json:"schema_version"`
	SourceAvailability    string    `json:"source_availability"`
	PayloadPlaintextBytes int64     `json:"payload_plaintext_bytes"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type ListMemoriesParams struct {
	Platform string
	Limit    int
	Cursor   string
}
