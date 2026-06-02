import { HttpError } from './types';
async function parseError(res) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    return body.error ?? `Request failed (${res.status})`;
}
export function createMemberVotingService(token, apiBaseUrl = '') {
    const base = apiBaseUrl.replace(/\/$/, '');
    return {
        async getContext() {
            const res = await fetch(`${base}/v1/api/member/gatherings/${token}`);
            if (!res.ok)
                throw new HttpError(await parseError(res), res.status);
            return res.json();
        },
        async submitBallot(content) {
            const res = await fetch(`${base}/v1/api/member/gatherings/${token}/ballot`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ballot_content: content })
            });
            if (!res.ok)
                throw new HttpError(await parseError(res), res.status);
            return res.json();
        }
    };
}
