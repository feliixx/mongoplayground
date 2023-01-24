import { expect } from '@playwright/test'
import { test, get, set } from './playwright'

test('format default page', async ({ page }) => {

  const configTxt = `[
  {
    "key": 1
  },
  {
    "key": 2
  }
]`
  const queryTxt = 'db.collection.find()'

  expect(await get('config')).toBe(configTxt)
  expect(await get('query')).toBe(queryTxt)

  await page.getByRole('button', { name: 'format' }).click()
  expect(await get('config')).toBe(configTxt)
  expect(await get('query')).toBe(queryTxt)
})

test('format with button', async ({ page }) => {

  await set('query', 'db.collection.find({key:1})')
  await page.getByRole('button', { name: 'format' }).click()

  expect(await get('query')).toBe(`db.collection.find({
  key: 1
})`)
})

test('format with shortcut', async ({ page }) => {

  await set('query', 'db.collection.find({key:1})')
  await page.getByText('Template').press('Control+s')

  expect(await get('query')).toBe(`db.collection.find({
  key: 1
})`)
})



