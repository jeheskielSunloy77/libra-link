import { initContract } from '@ts-rest/core'
import { authContract } from './auth.js'
import { ebookContract } from './ebook.js'
import { healthContract } from './health.js'
import { readerContract } from './reader.js'
import { shareContract } from './share.js'
import { syncContract } from './sync.js'
import { userContract } from './user.js'

const c = initContract()

export const apiContract = c.router({
	health: healthContract,
	auth: authContract,
	user: userContract,
	ebook: ebookContract,
	share: shareContract,
	reader: readerContract,
	sync: syncContract,
})
