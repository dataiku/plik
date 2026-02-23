import { test, expect, uploadTestFile } from './fixtures.js'

test.describe('Navigation and routing', () => {
    test('main upload page loads', async ({ page }) => {
        await page.goto('/')
        await page.waitForLoadState('networkidle')

        // The Upload link should be visible in the nav bar (it's a nav link, not a button)
        await expect(page.getByRole('link', { name: 'Upload', exact: true })).toBeVisible()
    })

    test('invalid upload ID shows error', async ({ page }) => {
        await page.goto('/#/?id=nonexistent123')
        await page.waitForLoadState('networkidle')

        // Error message should appear in the download view
        await expect(page.locator('.text-danger-500').first()).toBeVisible({ timeout: 5_000 })
    })

    test('clients page loads', async ({ page }) => {
        await page.goto('/#/clients')
        await page.waitForLoadState('networkidle')

        // Should show CLI client page content
        await expect(page.getByText(/client|download|plik/i).first()).toBeVisible()
    })
})
