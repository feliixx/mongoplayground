import { expect } from '@playwright/test';
import { test, get, set } from './playwright'

test('changing config enables share', async ({ page }) => {

  const shareButton = page.getByRole('button', { name: 'share' })
  await expect(shareButton).toBeDisabled()
  await set('config', '[]')
  await shareButton.click()
  await expect(page).toHaveURL(/p\/4btTeezhQ_i/)
  await expect(shareButton).toBeDisabled()
})

test('changing query enables share', async ({ page }) => {

  const shareButton = page.getByRole('button', { name: 'share' })
  await expect(shareButton).toBeDisabled()
  await set('query', 'db.c.find({v:"a"})')
  await shareButton.click()
  await expect(page).toHaveURL(/p\/IIAf09j3hnm/)
  await expect(shareButton).toBeDisabled()
})

test('changing mode enables share', async ({ page }) => {

  const shareButton = page.getByRole('button', { name: 'share' })
  await expect(shareButton).toBeDisabled()
  await page.locator('#custom-mode').getByRole('button', { name: 'bson' }).click()
  await page.locator('#custom-mode').getByText('mgodatagen').click()
  await shareButton.click()
  await expect(page).toHaveURL(/p\/cJxvGAak3VQ/)
  await expect(shareButton).toBeDisabled()
})

test('sharing format the playground', async ({ page }) => {

  await set('config', '[{}]')
  await page.getByRole('button', { name: 'share' }).click()
  await expect(page).toHaveURL(/p\/4cOeA7NGLru/)
  expect(await get('config')).toBe(`[
  {}
]`)
})

test('run after share does not change URL', async ({ page }) => {

  await set('config', 'db={"a":[{k:1}]}')
  await set('query', 'db.a.find({},{_id:0})')

  await page.getByRole('button', { name: 'run' }).click()
  expect(await get('result')).toBe(`[
  {
    "k": 1
  }
]`)
  await page.getByRole('button', { name: 'share' }).click()
  await expect(page).toHaveURL(/p\/iKNbEa-etwo/)
  await page.getByRole('button', { name: 'run' }).click()
  expect(await get('result')).toBe(`[
  {
    "k": 1
  }
]`)
  await expect(page).toHaveURL(/p\/iKNbEa-etwo/)
})

test('sharing show copied tooltip', async ({ page }) => {
  await expect(page.getByText("Copied")).toBeHidden()

  await set('config', '{')
  await page.getByRole('button', { name: 'share' }).click()
  await expect(page).toHaveURL(/p\/MMrQg5UYwYX/)

  await expect(page.getByText("Copied")).toBeVisible()
})
