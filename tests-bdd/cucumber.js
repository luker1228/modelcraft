// CJS mode: use requireModule + require (do NOT mix ESM loader)
module.exports = {
  default: {
    requireModule: ['tsx/cjs'],
    require: [
      'support/**/*.ts',
      'step-definitions/**/*.ts',
    ],
    paths: ['features/**/*.feature'],
    format: [
      'progress',
      'json:reports/test-report.json',
    ],
  },
}
