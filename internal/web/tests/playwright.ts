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
    let content = await getEditorContent(name)
    if (name === 'config' || name === 'query' || allowEmpty) {
        return content
    }
    while (!content || content === "running query...") {
        await currentPage.waitForTimeout(10)
        content = await getEditorContent(name)
    }
    return content
}

export async function set(name: string, content: string) {
    const textarea = currentPage.locator(`#${name}`).getByRole('textbox')
    await textarea.click({ force: true })
    await textarea.press('Control+a+Delete')
    await textarea.fill(content)
}

async function getEditorContent(name: string) {
    const textLayer = currentPage.locator(`#${name} > .ace_scroller > .ace_content > .ace_text-layer`)
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