import {
	ZPaginatedResponse,
	ZResponseWithData,
	ZStoreSyncEventDTO,
	ZSyncEvent,
	ZSyncEventsQuery,
} from '@libra-link/zod'
import { initContract } from '@ts-rest/core'
import { failResponses, getSecurityMetadata } from '../utils.js'

const c = initContract()

export const syncContract = c.router({
	storeEvent: {
		summary: 'Store sync event',
		description: 'Submit one offline sync event for current user.',
		method: 'POST',
		path: '/api/v1/sync/events',
		body: ZStoreSyncEventDTO,
		responses: {
			201: ZResponseWithData(ZSyncEvent),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
	listEvents: {
		summary: 'List sync events',
		description: 'List sync events for current user since optional timestamp.',
		method: 'GET',
		path: '/api/v1/sync/events',
		query: ZSyncEventsQuery,
		responses: {
			200: ZPaginatedResponse(ZSyncEvent),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
})
