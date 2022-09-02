/** @type {import('@playwright/test').PlaywrightTestConfig} */
const config = {
  webServer: {
    command: 'yarn build && yarn preview',
    port: 4173,
  },
}

export default config
