import { test, expect } from '@playwright/test';

const ROOT = 'https://ljluestc.github.io/';

function isSameOrigin(url: string): boolean {
	try {
		const u = new URL(url);
		const r = new URL(ROOT);
		return u.origin === r.origin;
	} catch {
		return false;
	}
}

function isSkippable(url: string): boolean {
	return url.startsWith('mailto:') || url.startsWith('tel:') || url.includes('#');
}

test('site-wide link crawler has no 404s or Page Not Found', async ({ browser }) => {
	const context = await browser.newContext();
	const page = await context.newPage();
	const visited = new Set<string>();
	const queue: string[] = [ROOT];
	const maxPages = 200;

	while (queue.length && visited.size < maxPages) {
		const url = queue.shift()!;
		if (visited.has(url)) continue;
		visited.add(url);

		const response = await page.goto(url, { waitUntil: 'load' });
		expect(response, `No response for ${url}`).toBeTruthy();
		const status = response!.status();
		expect(status, `Bad status ${status} at ${url}`).toBeLessThan(400);

		const content = await page.content();
		expect.soft(/Page Not Found/i.test(content)).toBeFalsy();

		const links = await page.$$eval('a[href]', as => as.map(a => (a as HTMLAnchorElement).href));
		for (const href of links) {
			if (isSkippable(href)) continue;
			if (!isSameOrigin(href)) continue;
			if (!visited.has(href)) queue.push(href);
		}
	}

	expect(visited.size).toBeGreaterThan(0);
	await context.close();
});
