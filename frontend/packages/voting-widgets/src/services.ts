import type { VotingService, MemberContext, BallotVote, BallotReceipt } from './types'
import { HttpError } from './types'

async function parseError(res: Response): Promise<string> {
  const body = await res.json().catch(() => ({ error: res.statusText }))
  return body.error ?? `Request failed (${res.status})`
}

export function createMemberVotingService(token: string, apiBaseUrl = ''): VotingService {
  const base = apiBaseUrl.replace(/\/$/, '')
  return {
    async getContext(): Promise<MemberContext> {
      const res = await fetch(`${base}/v1/api/member/gatherings/${token}`)
      if (!res.ok) throw new HttpError(await parseError(res), res.status)
      return res.json()
    },
    async submitBallot(content: Record<string, BallotVote>): Promise<BallotReceipt> {
      const res = await fetch(`${base}/v1/api/member/gatherings/${token}/ballot`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ballot_content: content })
      })
      if (!res.ok) throw new HttpError(await parseError(res), res.status)
      return res.json()
    }
  }
}
