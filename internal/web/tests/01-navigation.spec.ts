import { expect } from '@playwright/test'
import { test } from './playwright'

test('homepage', async ({ page }) => {

  await expect(page).toHaveTitle(/Mongo playground/)

  await expect(page.getByRole('link', { name: 'Report an issue' })).toBeVisible()

  await expect(page.getByText(/MongoDB version \d.\d.\d+/)).toBeVisible()
})

test('about page', async ({ page }) => {

  await page.getByRole('link', { name: 'About this playground' }).click()
  await expect(page).toHaveURL(/about.html/)
  await expect(page).toHaveTitle(/Mongo playground/)

  await expect(page.getByRole('link', { name: '@feliixx' })).toBeVisible()

  await page.getByText('Mongo Playground').click()
  await expect(page).toHaveURL(/\//)
})

test('documentation page', async ({ page }) => {

  await page.getByRole('button', { name: 'docs' }).click()
  await expect(page).toHaveURL(/\//)
  await expect(page).toHaveTitle(/Mongo playground/)

  await expect(page.getByRole('heading', { name: 'Summary' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Database' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Query' })).toBeHidden()
  await expect(page.getByRole('heading', { name: 'Result' })).toBeHidden()

  await page.getByRole('button', { name: 'docs' }).click()

  await expect(page.getByRole('heading', { name: 'Summary' })).toBeHidden()
  await expect(page.getByRole('heading', { name: 'Database' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Query' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Result' })).toBeVisible()
})


test('run should close documentation page', async ({ page }) => {

  await page.getByRole('button', { name: 'docs' }).click()

  await expect(page.getByRole('heading', { name: 'Summary' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Database' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Query' })).toBeHidden()
  await expect(page.getByRole('heading', { name: 'Result' })).toBeHidden()

  await page.getByRole('button', { name: 'run' }).click()

  await expect(page.getByRole('heading', { name: 'Summary' })).toBeHidden()
  await expect(page.getByRole('heading', { name: 'Database' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Query' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Result' })).toBeVisible()
})
