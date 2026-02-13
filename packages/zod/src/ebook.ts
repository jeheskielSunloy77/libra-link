import { z } from 'zod'
import { ZModel } from './utils.js'

export const ZEbookFormat = z.enum(['epub', 'pdf', 'txt'])

export const ZEbook =
	z
		.object({
			ownerUserId: z.string().uuid(),
			title: z.string(),
			description: z.string().optional(),
			format: ZEbookFormat,
			languageCode: z.string().optional(),
			storageKey: z.string(),
			fileSizeBytes: z.number().int(),
			checksumSha256: z.string(),
			importedAt: z.string().datetime(),
		})
		.extend(ZModel.shape)

export const ZStoreEbookDTO = z.object({
	title: z.string().min(1).max(255),
	description: z.string().optional(),
	format: ZEbookFormat,
	languageCode: z.string().max(16).optional(),
	storageKey: z.string().min(1),
	fileSizeBytes: z.number().int().positive(),
	checksumSha256: z.string().length(64),
	importedAt: z.string().datetime().optional(),
})

export const ZUpdateEbookDTO = z.object({
	title: z.string().min(1).max(255).optional(),
	description: z.string().optional(),
	languageCode: z.string().max(16).optional(),
})

export const ZEbookGoogleMetadata = z.object({
	ebookId: z.string().uuid(),
	googleBooksId: z.string(),
	isbn10: z.string().optional(),
	isbn13: z.string().optional(),
	publisher: z.string().optional(),
	publishedDate: z.string().optional(),
	pageCount: z.number().int().optional(),
	categories: z.array(z.string()).optional(),
	thumbnailUrl: z.string().url().optional(),
	infoLink: z.string().url().optional(),
	rawPayload: z.record(z.any()).optional(),
	attachedAt: z.string().datetime(),
	updatedAt: z.string().datetime(),
	deletedAt: z.string().datetime().optional(),
})

export const ZAttachGoogleMetadataDTO = z.object({
	googleBooksId: z.string().min(1),
	isbn10: z.string().length(10).optional(),
	isbn13: z.string().length(13).optional(),
	publisher: z.string().optional(),
	publishedDate: z.string().optional(),
	pageCount: z.number().int().positive().optional(),
	categories: z.array(z.string()).optional(),
	thumbnailUrl: z.string().url().optional(),
	infoLink: z.string().url().optional(),
	rawPayload: z.record(z.any()).optional(),
})
