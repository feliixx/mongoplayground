import { test as baseTest, expect, Page } from "@playwright/test";

let currentPage: Page

export const test = baseTest.extend({
    page: async ({ page }, use) => {
        currentPage = page
        await page.goto('/');
        await use(page);
    },
});

type Editor = 'config' | 'query' | 'result'

export async function expectEditorContent(name: Editor, expected: string) {
    expect.poll(async () => {
        try {
            return await getEditorContent(name)
        } catch (e: any) {
            return ''
        } 
    }).toBe(expected);
}

export async function setEditorContent(name: Editor, content: string) {
    const textarea = currentPage.locator(`#${name}`).getByRole('textbox')
    await textarea.click({ force: true })
    await textarea.press('Control+a+Delete')
    await textarea.fill(content)
}

async function getEditorContent(name: Editor) {
    const textLayer = currentPage.locator(`#${name} .ace_text-layer`)
    // textLayer looks like:
    // 
    //<div class="ace_line"><span>[</span></div>
    //<div class="ace_line"> <span>{</span></div>
    //<div class="ace_line"><span class="ace_indent-guide"> </span> <span class="ace_string">"key"</span>: <span class="ace_constant ace_numeric">1</span></div>
    //<div class="ace_line"> <span>}</span>,</div>
    //<div class="ace_line"><span>]</span></div>
    //
    // we can't use innerText() here because it doesn't work reliably with withespaces 
    return (await textLayer.innerHTML())
        .replace(/<\/div>/gi, "\n")     // replace all ace_line closing tag with line return
        .replace(/(<([^>]+)>)/ig, "")   // remove all HTML tags
        .replace(/\n+$/, "")            // remove trailing line return
        .replace(/(\n)\1+/, "\n")       // remove extra line returns caused by ace_line_group
}