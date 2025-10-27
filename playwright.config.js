/** @type {import('@playwright/test').PlaywrightTestConfig} */
const config = {
	timeout: 30000,
	reporter: [['list']],
	use: {
		headless: true,
		navigationTimeout: 15000,
	},
	workers: 2,
	testDir: 'tests',
};

module.exports = config;
