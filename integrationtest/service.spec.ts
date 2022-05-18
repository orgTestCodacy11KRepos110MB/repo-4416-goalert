import { test, expect } from '@playwright/test'
import { gotoDetails, createService, deleteService } from './util'

import Chance from 'chance'
import { userSession } from './login'
const c = new Chance()

test.describe.configure({ mode: 'parallel' })
test.use({ storageState: userSession })
test('editing a service', async ({ page }) => {
  await page.goto('./')

  const svc = await createService(page)
  await gotoDetails(page, 'Services', svc.name)

  const newName = 'zz_test_svc ' + c.word({ length: 12 })
  const desc = c.sentence()
  await page.locator('[aria-label="Edit"]').click()
  await page.locator('input[name="name"]').fill(newName)
  await page.locator('[name="description"]').fill(desc)
  await page.click('text=Submit')

  await expect(page.locator('h2', { hasText: newName })).toBeVisible()
  await expect(
    page.locator('[data-cy=details]', { hasText: desc }),
  ).toBeVisible()

  await deleteService(page, newName)
})

test('graphql', async ({ page }) => {
  await page.goto('./api/graphql/explore')
  await page.click('[aria-label="Query Editor"]')
  await page.keyboard.press('Control+A')
  await page.keyboard.type('qu')
  await expect(page.locator('#cm-complete-0-0')).toHaveText('query')
  await page.keyboard.press('Enter')
  await page.keyboard.type('{')
  await page.keyboard.press('Enter')
  await page.keyboard.type('u')
  await page.locator('#cm-complete-0').locator('li >> "user"').click()
  await expect(
    await page.locator('span.cm-property', { hasText: 'user' }),
  ).toBeVisible()
  await page.fill('.query-editor textarea', '{name}')
  await page.waitForTimeout(100) // graphiql has a hard-coded 100ms change delay, we need to match it before running the query
  await page.keyboard.press('Control+Enter')

  await expect(page.locator('.result-window')).toContainText('test.user', {
    timeout: 3000,
  })
})

test('integration keys', async ({ page }) => {
  await page.goto('./')
  const svc = await createService(page)
  await gotoDetails(page, 'Services', svc.name)

  await page.locator('a', { hasText: 'Integration Keys' }).click()
  await page.locator('[aria-label="Create Integration Key"]').click()

  const intName = 'zz_test_int ' + c.word({ length: 12 })
  await page.fill('input[name="name"]', intName)
  await page.click('text=Submit')

  const intURL = (await page
    .locator('li', { hasText: intName })
    .locator('a', { hasText: 'Copy' })
    .getAttribute('href')) as string
  await expect(intURL).toBeTruthy()

  const summary = c.word({ length: 12 })
  const summary2 = c.word({ length: 12 })
  const details = c.sentence()
  await Promise.all([
    page.request.post(intURL, {
      failOnStatusCode: true,
      form: {
        summary,
        details,
      },
    }),
    page.request.post(intURL, {
      failOnStatusCode: true,
      form: {
        summary: summary2,
        details,
      },
    }),
  ])

  // go back to details
  await page.locator('a', { hasText: svc.name }).click()
  await page.locator('a', { hasText: 'Manage alerts' }).click()

  await expect(page.locator('li', { hasText: summary2 })).toBeVisible()

  const closeByAPI = page.request.post(intURL, {
    failOnStatusCode: true,
    form: {
      summary: summary2,
      details,
      action: 'close',
    },
  })

  await page.locator('li', { hasText: summary }).click()
  await page.locator('button[aria-label="Close"]').click()
  await closeByAPI
  await page.goBack()

  await expect(page.locator('ul', { hasText: 'No results' })).toBeVisible()

  await deleteService(page, svc.name)
})
