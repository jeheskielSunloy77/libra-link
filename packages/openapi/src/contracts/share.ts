import {
	ZBorrow,
	ZBorrowShareDTO,
	ZCreateShareReportDTO,
	ZEmpty,
	ZShare,
	ZShareReport,
	ZShareReview,
	ZStoreShareDTO,
	ZUpdateShareDTO,
	ZUpsertShareReviewDTO,
	ZResponseWithData,
} from '@libra-link/zod'
import { initContract } from '@ts-rest/core'
import { z } from 'zod'
import { failResponses, getSecurityMetadata } from '../utils.js'
import { createResourceContract } from './resource.js'

const c = initContract()

const idParams = z.object({ id: z.string().uuid() })

export const shareContract = c.router({
	...createResourceContract({
		path: '/api/v1/shares',
		resource: 'Share',
		resourcePlural: 'Shares',
		schemas: {
			entity: ZShare,
			createDTO: ZStoreShareDTO,
			updateDTO: ZUpdateShareDTO,
		},
	}),
	borrow: {
		summary: 'Borrow share',
		description: 'Borrow a shared ebook if rules allow.',
		method: 'POST',
		path: '/api/v1/shares/:id/borrow',
		pathParams: idParams,
		body: ZBorrowShareDTO,
		responses: {
			201: ZResponseWithData(ZBorrow),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
	returnBorrow: {
		summary: 'Return borrow',
		description: 'Return an active borrow by borrow id.',
		method: 'POST',
		path: '/api/v1/borrows/:id/return',
		pathParams: idParams,
		body: ZEmpty,
		responses: {
			200: ZResponseWithData(ZBorrow),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
	upsertReview: {
		summary: 'Upsert share review',
		description: 'Create or update the current user review for a share.',
		method: 'PUT',
		path: '/api/v1/shares/:id/review',
		pathParams: idParams,
		body: ZUpsertShareReviewDTO,
		responses: {
			200: ZResponseWithData(ZShareReview),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
	createReport: {
		summary: 'Create share report',
		description: 'Report problematic share content.',
		method: 'POST',
		path: '/api/v1/shares/:id/report',
		pathParams: idParams,
		body: ZCreateShareReportDTO,
		responses: {
			201: ZResponseWithData(ZShareReport),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
})
