/**
 * Next.js instrumentation hook - runs once on server startup.
 * Validates required BFF environment variables before the server accepts requests.
 */
export async function register() {
  // Only validate on the Node.js runtime (not Edge), and only once
  if (process.env.NEXT_RUNTIME !== 'nodejs') return

  const requiredEnvVars: Array<{ name: string; description: string }> = [
    {
      name: 'BACKEND_URL',
      description: 'Internal URL of the Go backend (e.g. http://9.135.32.8:8090)',
    },
  ]

  const missing = requiredEnvVars.filter(({ name }) => !process.env[name])

  if (missing.length > 0) {
    const lines = missing
      .map(({ name, description }) => `  - ${name}: ${description}`)
      .join('\n')

    console.error(`
╔══════════════════════════════════════════════════════════════════╗
║              BFF STARTUP FAILED: Missing Environment Variables   ║
╚══════════════════════════════════════════════════════════════════╝

The following required environment variables are not set:

${lines}

Please add them to your .env file.
See .env.example for reference values.
`)
    process.exit(1)
  }

  console.log('[BFF] Environment validation passed ✓')
}
