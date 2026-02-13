import { z } from 'zod'
import { ZGetManyQuery } from './utils.js'

export const ZSyncEntityType = z.enum([
	'progress',
	'annotation',
	'bookmark',
	'preference',
	'reader_state',
])
export const ZSyncOperation = z.enum(['upsert', 'delete'])

export const ZSyncEvent = z.object({
	id: z.string().uuid(),
	userId: z.string().uuid(),
	entityType: ZSyncEntityType,
	entityId: z.string().uuid(),
	operation: ZSyncOperation,
	payload: z.record(z.any()).optional(),
	baseVersion: z.number().int().optional(),
	clientTimestamp: z.string().datetime(),
	serverTimestamp: z.string().datetime(),
	idempotencyKey: z.string(),
	createdAt: z.string().datetime(),
})

export const ZStoreSyncEventDTO = z.object({
	entityType: ZSyncEntityType,
	entityId: z.string().uuid(),
	operation: ZSyncOperation,
	payload: z.record(z.any()).optional(),
	baseVersion: z.number().int().optional(),
	clientTimestamp: z.string().datetime().optional(),
	idempotencyKey: z.string().min(8),
})

export const ZSyncEventsQuery = ZGetManyQuery.extend({
	since: z.string().datetime().optional(),
})
