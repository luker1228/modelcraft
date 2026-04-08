import { faker } from '@faker-js/faker'

export function createMockModel(override: Record<string, unknown> = {}) {
  return {
    id: faker.string.uuid(),
    name: faker.helpers.slugify(faker.word.noun()).toLowerCase(),
    displayName: faker.commerce.productName(),
    description: faker.lorem.sentence(),
    createdAt: faker.date.recent().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    ...override,
  }
}

export function createMockModelList(count = 5) {
  return Array.from({ length: count }, () => createMockModel())
}
