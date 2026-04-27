export interface E2EAccount {
  phone: string
  userName: string
  password: string
}

export function makeAccount(): E2EAccount {
  const seed = `${Date.now()}${Math.floor(Math.random() * 1000000)
    .toString()
    .padStart(6, '0')}`
  const suffix = seed.slice(-10)
  const secondDigit = (3 + (Number(suffix[0]) % 7)).toString()

  return {
    phone: `1${secondDigit}${suffix.slice(1)}`,
    userName: `u${suffix}`,
    password: `Pwd_${suffix}`,
  }
}
