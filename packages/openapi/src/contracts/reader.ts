import {
	ZAnnotation,
	ZBookmark,
	ZReadingProgress,
	ZResponseWithData,
	ZStoreAnnotationDTO,
	ZStoreBookmarkDTO,
	ZStoreReadingProgressDTO,
	ZUpdateAnnotationDTO,
	ZUpdateBookmarkDTO,
	ZUpdateReadingProgressDTO,
	ZUpdateUserPreferencesDTO,
	ZUpdateUserReaderStateDTO,
	ZUserPreferences,
	ZUserReaderState,
} from '@libra-link/zod'
import { initContract } from '@ts-rest/core'
import { failResponses, getSecurityMetadata } from '../utils.js'
import { createResourceContract } from './resource.js'

const c = initContract()

export const readerContract = c.router({
	readingProgress: createResourceContract({
		path: '/api/v1/reading-progress',
		resource: 'Reading Progress',
		resourcePlural: 'Reading Progress Entries',
		schemas: {
			entity: ZReadingProgress,
			createDTO: ZStoreReadingProgressDTO,
			updateDTO: ZUpdateReadingProgressDTO,
		},
	}),
	bookmarks: createResourceContract({
		path: '/api/v1/bookmarks',
		resource: 'Bookmark',
		resourcePlural: 'Bookmarks',
		schemas: {
			entity: ZBookmark,
			createDTO: ZStoreBookmarkDTO,
			updateDTO: ZUpdateBookmarkDTO,
		},
	}),
	annotations: createResourceContract({
		path: '/api/v1/annotations',
		resource: 'Annotation',
		resourcePlural: 'Annotations',
		schemas: {
			entity: ZAnnotation,
			createDTO: ZStoreAnnotationDTO,
			updateDTO: ZUpdateAnnotationDTO,
		},
	}),
	getUserPreferences: {
		summary: 'Get user preferences',
		description: 'Get reading preferences for current user.',
		method: 'GET',
		path: '/api/v1/users/preferences',
		responses: {
			200: ZResponseWithData(ZUserPreferences),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
	patchUserPreferences: {
		summary: 'Patch user preferences',
		description: 'Partially update reading preferences for current user.',
		method: 'PATCH',
		path: '/api/v1/users/preferences',
		body: ZUpdateUserPreferencesDTO,
		responses: {
			200: ZResponseWithData(ZUserPreferences),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
	getUserReaderState: {
		summary: 'Get user reader state',
		description: 'Get reader session state for current user.',
		method: 'GET',
		path: '/api/v1/users/reader-state',
		responses: {
			200: ZResponseWithData(ZUserReaderState),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
	patchUserReaderState: {
		summary: 'Patch user reader state',
		description: 'Partially update current reader state for current user.',
		method: 'PATCH',
		path: '/api/v1/users/reader-state',
		body: ZUpdateUserReaderStateDTO,
		responses: {
			200: ZResponseWithData(ZUserReaderState),
			...failResponses,
		},
		metadata: getSecurityMetadata(),
	},
})
