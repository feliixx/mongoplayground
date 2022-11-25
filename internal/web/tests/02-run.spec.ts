import { expect } from '@playwright/test';
import { test, get, set } from './playwright'

test('run default page with button', async ({ page }) => {
  expect(await get('result', true)).toBe('')

  await page.getByRole('button', { name: 'run' }).click()
  expect(await get('result')).toBe(`[
  {
    "_id": ObjectId("5a934e000102030405000000"),
    "key": 1
  },
  {
    "_id": ObjectId("5a934e000102030405000001"),
    "key": 2
  }
]`)
})

test('run default page with shortcut', async ({ page }) => {
  expect(await get('result', true)).toBe('')

  await page.getByText('Mongo Playground').press('Control+Enter')
  expect(await get('result')).toBe(`[
  {
    "_id": ObjectId("5a934e000102030405000000"),
    "key": 1
  },
  {
    "_id": ObjectId("5a934e000102030405000001"),
    "key": 2
  }
]`)
})

test('run incorrect config', async ({ page }) => {

  await set('config', `[
  {
    a: invalid
  }
]`)

  await page.getByText('3').hover()
  await expect(page.locator('#config > .ace_tooltip')).toBeVisible()
  await expect(page.locator('#config > .ace_tooltip')).toHaveText(`Unknown type: 'invalid'`)
  await page.getByRole('button', { name: 'run' }).click()

  await expect(page.locator('#resultPanel')).toHaveClass('text_red')
  expect(await get('result')).toBe(`Invalid configuration:
Line 3: Unknown type: 'invalid'`)
})

test('run no result', async ({ page }) => {

  await set('config', `[{k:1}]`)
  await set('query', `db.collection.find({k:2})`)
  await page.getByRole('button', { name: 'run' }).click()

  await expect(page.locator('#resultPanel')).not.toHaveClass('text_red')
  expect(await get('result')).toBe(`no document found`)
})


test('aggregation query with stages', async ({ page }) => {

  await set('query', `db.collection.aggregate([
    {
      "$project": {
        "_id": 0
      }
    },
    {
      "$match": {
        "key": 1
      }
    }
  ])`)

  await page.getByRole('button', { name: '$match'}).click()
  await expect(page.locator('#custom-aggregation_stages ul').getByText('$project')).toBeVisible()
  await expect(page.locator('#custom-aggregation_stages ul').getByText('$match')).toBeVisible()

  await page.getByRole('button', { name: 'run' }).click()
  expect(await get('result')).toBe(`[
  {
    "key": 1
  }
]`)

await page.getByRole('button', { name: '$match'}).click()
await page.locator('#custom-aggregation_stages ul').getByText('$project').click()

  await page.getByRole('button', { name: 'run' }).click()
  expect(await get('result')).toBe(`[
  {
    "key": 1
  },
  {
    "key": 2
  }
]`)
})


