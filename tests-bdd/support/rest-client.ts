import 'dotenv/config'

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080'

export interface TokenResponse {
  accessToken: string
  tokenType: string
}

export class RestClient {
  /**
   * 用 OAuth code 换取 accessToken。
   * 在测试中，通过环境变量 TEST_ACCESS_TOKEN 直接提供 token，
   * 本方法留作 OAuth 完整流程备用。
   */
  async getTokenByCode(code: string): Promise<TokenResponse> {
    const res = await fetch(`${API_BASE_URL}/api/auth/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code }),
    })
    if (!res.ok) {
      throw new Error(`Auth failed: ${res.status} ${await res.text()}`)
    }
    return res.json() as Promise<TokenResponse>
  }
}
