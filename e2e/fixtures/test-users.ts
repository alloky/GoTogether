export const testUsers = {
  alice: {
    name: 'Alice Test',
    email: `alice-${Date.now()}@test.com`,
    password: 'password123',
  },
  bob: {
    name: 'Bob Test',
    email: `bob-${Date.now()}@test.com`,
    password: 'password123',
  },
  charlie: {
    name: 'Charlie Test',
    email: `charlie-${Date.now()}@test.com`,
    password: 'password123',
  },
};

export function uniqueEmail(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 7)}@test.com`;
}
