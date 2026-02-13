import {
	ZAttachGoogleMetadataDTO,
	ZEbook,
	ZEbookGoogleMetadata,
	ZResponse,
	ZResponseWithData,
	ZStoreEbookDTO,
	ZUpdateEbookDTO,
} from '@libra-link/zod'
import { initContract } from '@ts-rest/core'
import { z } from 'zod'
import { failResponses, getSecurityMetadata } from '../utils.js'
import { createResourceContract } from './resource.js'

const c = initContract()

const idParams = z.object({ id: z.string().uuid() })

export const ebookContract = c.router({
	...createResourceContract({
		path: '/api/v1/ebooks',
		resource: 'Ebook',
		resourcePlural: 'Ebooks',
		schemas: {
			entity: ZEbook,
			createDTO: ZStoreEbookDTO,
			updateDTO: ZUpdateEbookDTO,
		},
	}),
	attachMetadata: {
		summary: 'Attach Google metadata to ebook',
		description: 'Upsert Google Books metadata for an ebook.',
		method: 'POST',
		path: '/api/v1/ebooks/:id/metadata',
		pathParams: idParams,
		body: ZAttachGoogleMetadataDTO,
		responses: {
			200: ZResponseWithData(ZEbookGoogleMetadata),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
	removeMetadata: {
		summary: 'Detach Google metadata from ebook',
		description: 'Soft-delete current Google Books metadata attachment for an ebook.',
		method: 'DELETE',
		path: '/api/v1/ebooks/:id/metadata',
		pathParams: idParams,
		responses: {
			200: ZResponse,
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
})
