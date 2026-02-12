package auth_models

import "time"

// OTPRecord represents a numeric code sent to a user for password reset
type OTPRecord struct {
	Email     string    `bson:"email" json:"email"`
	Code      string    `bson:"code" json:"code"`
	ExpiresAt time.Time `bson:"expires_at" json:"expires_at"`
}
