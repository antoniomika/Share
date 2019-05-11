package utils

import (
	"time"

	"github.com/h2non/filetype/types"
)

// UserObject is the base user
type UserObject struct {
	Email    string
	Password string
}

// FirebaseConfig is the firebase configuration
type FirebaseConfig struct {
	APIKey            string
	AuthDomain        string
	DatabaseURL       string
	ProjectID         string
	StorageBucket     string
	MessagingSenderID string
	EditorURL         string
	IPAddress         string
}

// LinkObject is the link location
type LinkObject struct {
	URL          string
	Token        string
	Clicks       int
	Clickers     []string
	ShortURL     string
	CreateTime   time.Time
	ExpireTime   time.Time
	ExpireClicks int
}

// UploadObject is the data model for an upload
type UploadObject struct {
	StorageKey   string
	Filename     string
	Token        string
	Clicks       int
	Clickers     []string
	ShortURL     string
	ContentType  types.Type
	CreateTime   time.Time
	ExpireTime   time.Time
	ExpireClicks int
}
