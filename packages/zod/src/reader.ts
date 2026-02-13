import { z } from 'zod'
import { ZModel } from './utils.js'

export const ZReadingMode = z.enum(['normal', 'zen'])
export const ZThemeMode = z.enum(['light', 'dark', 'sepia', 'high_contrast'])
export const ZTypographyProfile = z.enum(['compact', 'comfortable', 'large'])

export const ZReadingProgress =
	z
		.object({
			userId: z.string().uuid(),
			ebookId: z.string().uuid(),
			location: z.string(),
			progressPercent: z.number().min(0).max(100).optional(),
			readingMode: ZReadingMode,
			rowVersion: z.number().int(),
			lastReadAt: z.string().datetime().optional(),
		})
		.extend(ZModel.shape)

export const ZStoreReadingProgressDTO = z.object({
	ebookId: z.string().uuid(),
	location: z.string().min(1),
	progressPercent: z.number().min(0).max(100).optional(),
	readingMode: ZReadingMode,
	lastReadAt: z.string().datetime().optional(),
})

export const ZUpdateReadingProgressDTO = z.object({
	location: z.string().min(1).optional(),
	progressPercent: z.number().min(0).max(100).optional(),
	readingMode: ZReadingMode.optional(),
	lastReadAt: z.string().datetime().optional(),
})

export const ZBookmark =
	z
		.object({
			userId: z.string().uuid(),
			ebookId: z.string().uuid(),
			location: z.string(),
			label: z.string().optional(),
			rowVersion: z.number().int(),
		})
		.extend(ZModel.shape)

export const ZStoreBookmarkDTO = z.object({
	ebookId: z.string().uuid(),
	location: z.string().min(1),
	label: z.string().optional(),
})

export const ZUpdateBookmarkDTO = z.object({
	label: z.string().optional(),
})

export const ZAnnotation =
	z
		.object({
			userId: z.string().uuid(),
			ebookId: z.string().uuid(),
			locationStart: z.string(),
			locationEnd: z.string(),
			highlightText: z.string().optional(),
			note: z.string().optional(),
			color: z.string().optional(),
			rowVersion: z.number().int(),
		})
		.extend(ZModel.shape)

export const ZStoreAnnotationDTO = z.object({
	ebookId: z.string().uuid(),
	locationStart: z.string().min(1),
	locationEnd: z.string().min(1),
	highlightText: z.string().optional(),
	note: z.string().optional(),
	color: z.string().optional(),
})

export const ZUpdateAnnotationDTO = z.object({
	locationStart: z.string().min(1).optional(),
	locationEnd: z.string().min(1).optional(),
	highlightText: z.string().optional(),
	note: z.string().optional(),
	color: z.string().optional(),
})

export const ZUserPreferences = z.object({
	userId: z.string().uuid(),
	readingMode: ZReadingMode,
	zenRestoreOnOpen: z.boolean(),
	themeMode: ZThemeMode,
	themeOverrides: z.record(z.string()),
	typographyProfile: ZTypographyProfile,
	rowVersion: z.number().int(),
	createdAt: z.string().datetime(),
	updatedAt: z.string().datetime(),
})

export const ZUpdateUserPreferencesDTO = z.object({
	readingMode: ZReadingMode.optional(),
	zenRestoreOnOpen: z.boolean().optional(),
	themeMode: ZThemeMode.optional(),
	themeOverrides: z.record(z.string()).optional(),
	typographyProfile: ZTypographyProfile.optional(),
})

export const ZUserReaderState = z.object({
	userId: z.string().uuid(),
	currentEbookId: z.string().uuid().optional(),
	currentLocation: z.string().optional(),
	readingMode: ZReadingMode,
	rowVersion: z.number().int(),
	lastOpenedAt: z.string().datetime().optional(),
	createdAt: z.string().datetime(),
	updatedAt: z.string().datetime(),
})

export const ZUpdateUserReaderStateDTO = z.object({
	currentEbookId: z.string().uuid().optional(),
	currentLocation: z.string().optional(),
	readingMode: ZReadingMode.optional(),
	lastOpenedAt: z.string().datetime().optional(),
})
