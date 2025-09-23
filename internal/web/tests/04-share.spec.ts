import { expect } from '@playwright/test';
import { test, setEditorContent, expectEditorContent } from './playwright'

test('changing config enables share', async ({ page }) => {

  const shareButton = page.getByRole('button', { name: 'share' })
  await expect(shareButton).toBeDisabled()
  await setEditorContent('config', '[]')
  await shareButton.click()
  await expect(page).toHaveURL(/p\/4btTeezhQ_i/)
  await expect(shareButton).toBeDisabled()
})

test('changing query enables share', async ({ page }) => {

  const shareButton = page.getByRole('button', { name: 'share' })
  await expect(shareButton).toBeDisabled()
  await setEditorContent('query', 'db.c.find({v:"a"})')
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

  await setEditorContent('config', '[{}]')
  await page.getByRole('button', { name: 'share' }).click()
  await expect(page).toHaveURL(/p\/4cOeA7NGLru/)
  await expectEditorContent('config', `[
  {}
]`)
})

test('run after share does not change URL', async ({ page }) => {

  await setEditorContent('config', 'db={"a":[{k:1}]}')
  await setEditorContent('query', 'db.a.find({},{_id:0})')

  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "k": 1
  }
]`)
  await page.getByRole('button', { name: 'share' }).click()
  await expect(page).toHaveURL(/p\/iKNbEa-etwo/)
  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "k": 1
  }
]`)
  await expect(page).toHaveURL(/p\/iKNbEa-etwo/)
})

test('sharing show copied tooltip', async ({ page }) => {
  await expect(page.getByText("Copied")).toBeHidden()

  await setEditorContent('config', '{')
  await page.getByRole('button', { name: 'share' }).click()
  await expect(page).toHaveURL(/p\/MMrQg5UYwYX/)

  await expect(page.getByText("Copied")).toBeVisible()
})

test('saving the same playground twice returns the same URL', async ({ page }) => {
  await expect(page.getByText("Copied")).toBeHidden()

  await setEditorContent('config', '{"_id":1}')
  await page.getByRole('button', { name: 'share' }).click()
  await expect(page).toHaveURL(/p\/Cz5OkFt6TSH/)

  await setEditorContent('config', '')
  await setEditorContent('config', '{"_id":1}')
  await page.getByRole('button', { name: 'share' }).click()
  await expect(page).toHaveURL(/p\/Cz5OkFt6TSH/)
})