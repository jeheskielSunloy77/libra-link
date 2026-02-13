import { z } from 'zod'
import { ZModel } from './utils.js'

export const ZShareVisibility = z.enum(['public', 'unlisted'])
export const ZShareStatus = z.enum(['active', 'disabled', 'removed'])
export const ZBorrowStatus = z.enum(['active', 'returned', 'expired', 'revoked'])
export const ZReportReason = z.enum(['copyright', 'abuse', 'spam', 'other'])
export const ZReportStatus = z.enum(['open', 'in_review', 'resolved', 'rejected'])

export const ZShare =
	z
		.object({
			ebookId: z.string().uuid(),
			ownerUserId: z.string().uuid(),
			titleOverride: z.string().optional(),
			description: z.string().optional(),
			visibility: ZShareVisibility,
			status: ZShareStatus,
			borrowDurationHours: z.number().int().positive(),
			maxConcurrentBorrows: z.number().int().positive(),
		})
		.extend(ZModel.shape)

export const ZStoreShareDTO = z.object({
	ebookId: z.string().uuid(),
	titleOverride: z.string().max(255).optional(),
	description: z.string().optional(),
	visibility: ZShareVisibility.optional(),
	status: ZShareStatus.optional(),
	borrowDurationHours: z.number().int().positive(),
	maxConcurrentBorrows: z.number().int().positive().optional(),
})

export const ZUpdateShareDTO = z.object({
	titleOverride: z.string().max(255).optional(),
	description: z.string().optional(),
	visibility: ZShareVisibility.optional(),
	status: ZShareStatus.optional(),
	borrowDurationHours: z.number().int().positive().optional(),
	maxConcurrentBorrows: z.number().int().positive().optional(),
})

export const ZBorrow = z.object({
	id: z.string().uuid(),
	createdAt: z.string().datetime(),
	updatedAt: z.string().datetime(),
	shareId: z.string().uuid(),
	borrowerUserId: z.string().uuid(),
	startedAt: z.string().datetime(),
	dueAt: z.string().datetime(),
	returnedAt: z.string().datetime().optional(),
	expiredAt: z.string().datetime().optional(),
	status: ZBorrowStatus,
	legalAcknowledgedAt: z.string().datetime(),
})

export const ZBorrowShareDTO = z.object({
	legalAcknowledged: z.literal(true),
})

export const ZShareReview =
	z
		.object({
			shareId: z.string().uuid(),
			userId: z.string().uuid(),
			rating: z.number().int().min(1).max(5),
			reviewText: z.string().optional(),
		})
		.extend(ZModel.shape)

export const ZUpsertShareReviewDTO = z.object({
	rating: z.number().int().min(1).max(5),
	reviewText: z.string().optional(),
})

export const ZShareReport = z.object({
	id: z.string().uuid(),
	createdAt: z.string().datetime(),
	updatedAt: z.string().datetime(),
	shareId: z.string().uuid(),
	reporterUserId: z.string().uuid(),
	reason: ZReportReason,
	details: z.string().optional(),
	status: ZReportStatus,
	reviewedByUserId: z.string().uuid().optional(),
	reviewedAt: z.string().datetime().optional(),
	resolutionNote: z.string().optional(),
})

export const ZCreateShareReportDTO = z.object({
	reason: ZReportReason,
	details: z.string().optional(),
})
