package api

import "time"

type User struct {
	ID       string
	Email    string
	Username string
}

type Ebook struct {
	ID          string
	Title       string
	Format      string
	StorageKey  string
	ImportedAt  time.Time
	Description string
}

type Share struct {
	ID          string
	EbookID     string
	OwnerUserID string
	Status      string
	Visibility  string
	Title       string
}

type Preferences struct {
	UserID            string
	ReadingMode       string
	ZenRestoreOnOpen  bool
	ThemeMode         string
	ThemeOverrides    map[string]string
	TypographyProfile string
	RowVersion        int
}

type ReaderState struct {
	UserID          string
	CurrentEbookID  string
	CurrentLocation string
	ReadingMode     string
	RowVersion      int
	LastOpenedAt    *time.Time
}

type SyncEvent struct {
	EntityType     string
	EntityID       string
	Operation      string
	Payload        map[string]any
	BaseVersion    *int
	ClientTS       time.Time
	IdempotencyKey string
}

type GoogleDeviceStart struct {
	DeviceCode      string
	AuthURL         string
	ExpiresAt       time.Time
	IntervalSeconds int
}

type GoogleDevicePoll struct {
	Status       string
	User         *User
	AccessToken  string
	RefreshToken string
}
