import type { Metadata } from 'next'
import { Inter, Space_Grotesk, Fira_Code } from 'next/font/google'
import { ApolloWrapper } from '@web/providers/apollo-wrapper'
import { QueryWrapper } from '@web/providers/query-wrapper'
import { ErrorProvider } from '@web/components/features/providers/ErrorProvider'
import { MSWProvider } from '@/mocks/MSWProvider'

import { Toaster } from 'sonner'
import '@/app/globals.css'

const inter = Inter({
  subsets: ['latin'],
  weight: ['300', '400', '500', '600', '700'],
  variable: '--font-inter',
  display: 'swap',
})

const spaceGrotesk = Space_Grotesk({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700'],
  variable: '--font-space-grotesk',
  display: 'swap',
})

const firaCode = Fira_Code({
  subsets: ['latin'],
  weight: ['400', '500', '600'],
  variable: '--font-fira-code',
  display: 'swap',
})

export const metadata: Metadata = {
  title: 'ModelCraft - AI Native Data Infrastructure',
  description: 'ModelCraft is an AI native data access layer with GraphQL and CLI for secure, controllable database usage.',
  icons: {
    icon: '/favicon.svg',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="zh-CN">
      <body className={`${inter.variable} ${spaceGrotesk.variable} ${firaCode.variable} font-sans`}>
        <MSWProvider>
          <ApolloWrapper>
            <QueryWrapper>
              <ErrorProvider>
                {children}
              </ErrorProvider>
            </QueryWrapper>
          </ApolloWrapper>
        </MSWProvider>
        <Toaster position="top-right" richColors />
      </body>
    </html>
  )
}
