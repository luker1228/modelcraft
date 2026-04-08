// 让 dotenv 加载 .env.test
process.env.DOTENV_CONFIG_PATH = process.env.DOTENV_CONFIG_PATH || '.env.test'

module.exports = {
  default: {
    requireModule: ['tsx/cjs'],
    require: [
      'support/**/*.ts',
      'step-definitions/**/*.ts',
    ],
    paths: ['features/**/*.feature'],
    format: [
      'progress-bar',
      'html:reports/test-report.html',
    ],
  },
}
