package domain

type EbookFormat string

const (
	EbookFormatEPUB EbookFormat = "epub"
	EbookFormatPDF  EbookFormat = "pdf"
	EbookFormatTXT  EbookFormat = "txt"
)

type ReadingMode string

const (
	ReadingModeNormal ReadingMode = "normal"
	ReadingModeZen    ReadingMode = "zen"
)

type ThemeMode string

const (
	ThemeModeLight        ThemeMode = "light"
	ThemeModeDark         ThemeMode = "dark"
	ThemeModeSepia        ThemeMode = "sepia"
	ThemeModeHighContrast ThemeMode = "high_contrast"
)

type TypographyProfile string

const (
	TypographyProfileCompact     TypographyProfile = "compact"
	TypographyProfileComfortable TypographyProfile = "comfortable"
	TypographyProfileLarge       TypographyProfile = "large"
)

type ShareVisibility string

const (
	ShareVisibilityPublic   ShareVisibility = "public"
	ShareVisibilityUnlisted ShareVisibility = "unlisted"
)

type ShareStatus string

const (
	ShareStatusActive   ShareStatus = "active"
	ShareStatusDisabled ShareStatus = "disabled"
	ShareStatusRemoved  ShareStatus = "removed"
)

type BorrowStatus string

const (
	BorrowStatusActive   BorrowStatus = "active"
	BorrowStatusReturned BorrowStatus = "returned"
	BorrowStatusExpired  BorrowStatus = "expired"
	BorrowStatusRevoked  BorrowStatus = "revoked"
)

type ReportReason string

const (
	ReportReasonCopyright ReportReason = "copyright"
	ReportReasonAbuse     ReportReason = "abuse"
	ReportReasonSpam      ReportReason = "spam"
	ReportReasonOther     ReportReason = "other"
)

type ReportStatus string

const (
	ReportStatusOpen     ReportStatus = "open"
	ReportStatusInReview ReportStatus = "in_review"
	ReportStatusResolved ReportStatus = "resolved"
	ReportStatusRejected ReportStatus = "rejected"
)

type SyncEntityType string

const (
	SyncEntityTypeProgress   SyncEntityType = "progress"
	SyncEntityTypeAnnotation SyncEntityType = "annotation"
	SyncEntityTypeBookmark   SyncEntityType = "bookmark"
	SyncEntityTypePreference SyncEntityType = "preference"
	SyncEntityTypeReader     SyncEntityType = "reader_state"
)

type SyncOperation string

const (
	SyncOperationUpsert SyncOperation = "upsert"
	SyncOperationDelete SyncOperation = "delete"
)
