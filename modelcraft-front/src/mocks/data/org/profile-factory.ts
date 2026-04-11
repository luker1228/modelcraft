import { faker } from '@faker-js/faker'

export interface MockMyUserProfileScenario {
  type: 'success' | 'profileNotFound' | 'invalidInput'
}

type MockUserStatus = 'REGISTERED' | 'ACTIVE' | 'SUSPENDED'

export interface MockProfile {
  id: string
  userId: string
  nickname: string
  avatarUrl?: string
  bio?: string
  createdAt: string
  updatedAt: string
}

export interface MockUserWithProfile {
  id: string
  phone: string
  userName: string
  status: MockUserStatus
  createdAt: string
  updatedAt: string
  profile: MockProfile
}

export interface MockMyUserProfilePayload {
  user: MockUserWithProfile | null
  error:
    | {
        __typename: 'UserNotFound' | 'ProfileNotFound'
        message: string
      }
    | null
}

export interface MockUpdateMyProfilePayload {
  profile: MockProfile | null
  error:
    | {
        __typename: 'ProfileNotFound'
        message: string
      }
    | {
        __typename: 'InvalidProfileInput'
        message: string
        suggestion: string
      }
    | null
}

function createMockProfile(override: Partial<MockProfile> = {}): MockProfile {
  const userId = override.userId ?? faker.string.uuid()
  const createdAt = override.createdAt ?? faker.date.recent().toISOString()

  const profile: MockProfile = {
    id: faker.string.uuid(),
    userId,
    nickname: `user_${faker.string.alphanumeric({ length: 6, casing: 'upper' })}`,
    avatarUrl: `mock://avatar/default-${faker.number.int({ min: 1, max: 6 })}.png`,
    bio: faker.lorem.sentence(),
    createdAt,
    updatedAt: override.updatedAt ?? createdAt,
  }

  return {
    ...profile,
    ...override,
  }
}

function createMockUserWithProfile(override: Partial<MockUserWithProfile> = {}): MockUserWithProfile {
  const id = override.id ?? faker.string.uuid()
  const createdAt = override.createdAt ?? faker.date.recent().toISOString()
  const profile = override.profile ?? createMockProfile({ userId: id })

  const user: MockUserWithProfile = {
    id,
    phone: `1${faker.string.numeric(10)}`,
    userName: faker.internet.username().toLowerCase(),
    status: faker.helpers.arrayElement<MockUserStatus>(['REGISTERED', 'ACTIVE', 'SUSPENDED']),
    createdAt,
    updatedAt: override.updatedAt ?? createdAt,
    profile,
  }

  return {
    ...user,
    ...override,
    profile: override.profile ?? profile,
  }
}

export function createMockMyUserProfilePayload(
  scenario: MockMyUserProfileScenario
): MockMyUserProfilePayload {
  switch (scenario.type) {
    case 'profileNotFound':
      return {
        user: null,
        error: {
          __typename: 'ProfileNotFound',
          message: 'Profile not found for current user',
        },
      }
    case 'invalidInput':
    case 'success':
    default:
      return {
        user: createMockUserWithProfile(),
        error: null,
      }
  }
}

export function createMockUpdateMyProfilePayload(
  scenario: MockMyUserProfileScenario
): MockUpdateMyProfilePayload {
  switch (scenario.type) {
    case 'profileNotFound':
      return {
        profile: null,
        error: {
          __typename: 'ProfileNotFound',
          message: 'Profile not found for current user',
        },
      }
    case 'invalidInput':
      return {
        profile: null,
        error: {
          __typename: 'InvalidProfileInput',
          message: 'At least one updatable profile field is required',
          suggestion: 'Provide nickname, avatarUrl, or bio to update profile',
        },
      }
    case 'success':
    default:
      return {
        profile: createMockProfile(),
        error: null,
      }
  }
}
