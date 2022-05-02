import { Chance } from 'chance'
import profile from '../fixtures/profile.json'

const c = new Chance()

declare global {
  namespace Cypress {
    interface Chainable {
      /** Creates a new user profile. */
      createUser: typeof createUser

      /** Creates multiple new user profiles. */
      createManyUsers: typeof createManyUsers

      /**
       * Resets the test user profile, including any existing contact methods.
       */
      resetProfile: typeof resetProfile

      /** Adds a contact method. If userID is missing, the test user's will be used. */
      addContactMethod: typeof addContactMethod

      /** Adds a notification rule. If userID is missing, the test user's will be used. */
      addNotificationRule: typeof addNotificationRule
    }
  }

  type UserRole = 'user' | 'admin'
  interface Profile {
    id: string
    name: string
    email: string
    role: UserRole
    username?: string
    passwordHash?: string
    isFavorite: boolean
  }

  interface UserOptions {
    name?: string
    email?: string
    role?: UserRole
    favorite?: boolean
  }

  type ContactMethodType = 'SMS' | 'VOICE'
  interface ContactMethod {
    id: string
    userID: string
    name: string
    type: ContactMethodType
    value: string
  }

  interface ContactMethodOptions {
    userID?: string
    name?: string
    type?: ContactMethodType
    value?: string
  }

  interface NotificationRule {
    id: string
    userID: string
    contactMethodID: string
    contactMethod: ContactMethod
    delayMinutes: number
  }

  interface NotificationRuleOptions {
    userID?: string
    delayMinutes?: number
    contactMethodID?: string
    contactMethod?: ContactMethodOptions
  }
}

function setFavoriteUser(id: string): Cypress.Chainable {
  const query = `
    mutation setFavorite($input: SetFavoriteInput!){
      setFavorite(input: $input)
    }
  `

  return cy.graphql(query, {
    input: {
      target: { type: 'user', id: id },
      favorite: true,
    },
  })
}

function createManyUsers(
  users: Array<UserOptions>,
): Cypress.Chainable<Array<Profile>> {
  const profiles: Array<Profile> = users.map((user) => ({
    id: c.guid(),
    name: user.name || c.word({ length: 12 }),
    email: user.email || c.email(),
    role: user.role || 'user',
    isFavorite: user.favorite || false,
  }))

  const dbQuery =
    `insert into users (id, name, email, role) values` +
    profiles
      .map((p) => `('${p.id}', '${p.name}', '${p.email}', '${p.role}')`)
      .join(',') +
    `;`

  return cy
    .sql(dbQuery)
    .then(() =>
      Promise.all(
        profiles.filter((p) => p.isFavorite).map((u) => setFavoriteUser(u.id)),
      ).then(() => profiles),
    )
}

function createUser(user?: UserOptions): Cypress.Chainable<Profile> {
  if (!user) user = {}
  return createManyUsers([user]).then((p) => p[0])
}

function addContactMethod(
  cm?: ContactMethodOptions,
): Cypress.Chainable<ContactMethod> {
  if (!cm) cm = {}
  if (!cm.userID) {
    return addContactMethod({ ...cm, userID: profile.id })
  }

  const mutation = `
    mutation ($input: CreateUserContactMethodInput!) {
      createUserContactMethod(input: $input) {
        id
        name
        type
        value
      }
    }
  `

  const newPhone = '+1763' + c.integer({ min: 3000000, max: 3999999 })
  return cy
    .graphql(mutation, {
      input: {
        userID: cm.userID,
        name: cm.name || 'SM CM ' + c.word({ length: 8 }),
        type: cm.type || c.pickone(['SMS', 'VOICE']),
        value: cm.value || newPhone,
      },
    })
    .then((res: GraphQLResponse) => {
      res = res.createUserContactMethod
      res.userID = cm && cm.userID
      return res as ContactMethod
    })
}

function addNotificationRule(
  nr?: NotificationRuleOptions,
): Cypress.Chainable<NotificationRule> {
  if (!nr) nr = {}
  if (!nr.userID) {
    return addNotificationRule({ ...nr, userID: profile.id })
  }

  if (!nr.contactMethodID) {
    return cy
      .addContactMethod({ ...nr.contactMethod, userID: nr.userID })
      .then((cm: ContactMethod) =>
        addNotificationRule({ ...nr, contactMethodID: cm.id }),
      )
  }

  const mutation = `
    mutation ($input: CreateUserNotificationRuleInput!) {
      createUserNotificationRule(input: $input) {
        id
        delayMinutes
        contactMethodID
        contactMethod {
          id
          name
          type
          value
        }
      }
    }
  `

  return cy
    .graphql(mutation, {
      input: {
        userID: nr.userID,
        contactMethodID: nr.contactMethodID,
        delayMinutes: nr.delayMinutes || c.integer({ min: 0, max: 15 }),
      },
    })
    .then((res: GraphQLResponse) => {
      res = res.createUserNotificationRule

      const userID = nr && nr.userID
      res.userID = userID
      res.contactMethod.userID = userID

      return res as NotificationRule
    })
}

function clearContactMethods(id: string): Cypress.Chainable {
  const query = `
    query($id: ID!) {
      user(id: $id) {
        contactMethods {
          id
        }
      }
    }
  `

  const mutation = `
    mutation($input: [TargetInput!]!) {
      deleteAll(input: $input)
    }
  `

  return cy.graphql(query, { id }).then((res: GraphQLResponse) => {
    if (!res.user.contactMethods.length) return

    res.user.contactMethods.forEach((cm: ContactMethod) => {
      cy.graphql(mutation, {
        input: [
          {
            type: 'contactMethod',
            id: cm.id,
          },
        ],
      })
    })
  })
}

function resetProfile(prof?: Profile): Cypress.Chainable {
  if (!prof) {
    return resetProfile(profile as unknown as Profile)
  }

  const mutation = `
    mutation updateUser($input: UpdateUserInput!) {
      updateUser(input: $input)
    }
  `

  return clearContactMethods(prof.id).graphql(mutation, {
    input: {
      id: prof.id,
      name: prof.name,
      email: prof.email,
    },
  })
}

Cypress.Commands.add('createUser', createUser)
Cypress.Commands.add('createManyUsers', createManyUsers)
Cypress.Commands.add('resetProfile', resetProfile)
Cypress.Commands.add('addContactMethod', addContactMethod)
Cypress.Commands.add('addNotificationRule', addNotificationRule)
