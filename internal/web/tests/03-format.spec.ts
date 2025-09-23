import { test, setEditorContent, expectEditorContent } from './playwright'

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

  await expectEditorContent('config', configTxt)
  await expectEditorContent('query', queryTxt)

  await page.getByRole('button', { name: 'format' }).click()
  await expectEditorContent('config', configTxt)
  await expectEditorContent('query', queryTxt)
})

test('format with button', async ({ page }) => {

  await setEditorContent('query', 'db.collection.find({key:1})')
  await page.getByRole('button', { name: 'format' }).click()

  await expectEditorContent('query', `db.collection.find({
  key: 1
})`)
})

test('format with shortcut', async ({ page }) => {

  await setEditorContent('query', 'db.collection.find({key:1})')
  await page.getByText('Template').press('Control+s')

  await expectEditorContent('query', `db.collection.find({
  key: 1
})`)
})



