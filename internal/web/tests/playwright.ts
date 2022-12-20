import { test as baseTest, Page } from "@playwright/test";

let currentPage: Page

export const test = baseTest.extend({
    page: async ({ page }, use) => {
        currentPage = page
        await page.goto('/');
        await use(page);
    },
});


export async function get(name: string, allowEmpty = false): Promise<string> {
    await currentPage.waitForTimeout(5)
    const editor = currentPage.locator(`#${name} > .ace_scroller > .ace_content`)
    let content = await editor.innerText()
    if (name === 'config' || name === 'query' || allowEmpty) {
        return content
    } 
    while (!content || content === "running query...") {
        await currentPage.waitForTimeout(10)
        content = await editor.innerText()
    }
    return content
}

export async function set(name: string, content: string) {
    const textarea = currentPage.locator(`#${name}`).getByRole('textbox')
    await textarea.click({force: true})
    await textarea.press('Control+a+Delete')
    await textarea.fill(content)
}