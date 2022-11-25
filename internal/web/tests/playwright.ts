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
    if (name != 'result') {
        return editor.innerText()
    }
    let result = await editor.innerText()
    if (allowEmpty) {
        return result
    }

    while (!result || result === "running query...") {
        await currentPage.waitForTimeout(10)
        result = await editor.innerText()
    }
    return result
}

export async function set(name: string, content: string) {
    const textarea = currentPage.locator(`#${name}`).getByRole('textbox')
    await textarea.click({force: true})
    await textarea.press('Control+a+Delete')
    await textarea.fill(content)
}