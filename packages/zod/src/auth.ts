import { z } from 'zod'

import { ZStoreUserDTO, ZUser } from './user.js'

export const ZAuthToken = z.object({
	token: z.string(),
	expiresAt: z.string().datetime(),
})

export const ZAuthResult = ZUser

export const ZAuthRegisterDTO = ZStoreUserDTO.pick({
	email: true,
	username: true,
	password: true,
})

export const ZAuthLoginDTO = z.object({
	identifier: z.string(),
	password: z.string(),
})

export const ZAuthGoogleCallbackQuery = z.object({
	code: z.string(),
	state: z.string(),
})

export const ZAuthGoogleDeviceStart = z.object({
	deviceCode: z.string(),
	authUrl: z.string().url(),
	expiresAt: z.string().datetime(),
	intervalSeconds: z.number().int().gte(1),
})

export const ZAuthGoogleDevicePollDTO = z.object({
	deviceCode: z.string().min(16),
})

export const ZAuthResultEnvelope = z.object({
	user: ZUser,
	token: ZAuthToken,
	refreshToken: ZAuthToken,
})

export const ZAuthGoogleDevicePollResponse = z.object({
	status: z.enum(['pending', 'approved', 'expired', 'failed']),
	result: ZAuthResultEnvelope.optional(),
})

export const ZAuthVerifyEmailDTO = z.object({
	email: z.string().email(),
	code: z.string().min(4).max(10),
})

export const ZAuthVerifyEmailResponse = ZUser
