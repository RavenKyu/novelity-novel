import { googleLoginUrl, localReturnTo } from '@novelity/api-client';
import type { PageServerLoad } from './$types';

const messages: Record<string, string> = {
	not_configured: '로그인은 준비 중입니다.',
	denied: '로그인이 취소되었습니다.',
	invalid_state: '로그인 요청이 만료되었습니다. 다시 시도해 주세요.',
	forbidden: '이 계정으로 로그인할 수 없습니다.'
};

export const load: PageServerLoad = ({ url }) => {
	const code = url.searchParams.get('error');
	return {
		loginUrl: googleLoginUrl(localReturnTo(url.searchParams.get('returnTo'))),
		errorMessage: code ? messages[code] || '로그인하지 못했습니다. 다시 시도해 주세요.' : null
	};
};
