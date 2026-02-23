import { test, expect, uploadTestFile } from './fixtures.js'

test.describe('Home view', () => {
    test('shows user sidebar info', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        // Should show the admin user's login in the sidebar user card
        // Use a more specific locator to avoid matching nav links
        await expect(page.locator('.glass-card').filter({ hasText: 'admin' }).first()).toBeVisible()
    })

    test('uploads tab shows upload cards', async ({ authenticatedPage: page }) => {
        // Create an upload first so there's something to show
        await uploadTestFile(page, 'home-test.txt', 'for home view')

        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        // Click Uploads sidebar nav button
        await page.getByRole('button', { name: 'Uploads', exact: true }).click()
        await page.waitForLoadState('networkidle')

        // Should show at least one upload card with the file name
        await expect(page.getByText('home-test.txt')).toBeVisible({ timeout: 5_000 })
    })

    test('tokens tab renders', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        // Click Tokens sidebar nav button
        await page.getByRole('button', { name: 'Tokens', exact: true }).click()
        await page.waitForLoadState('networkidle')

        // Should show the token creation area with "Create token" button
        await expect(page.getByRole('button', { name: /Create token/i })).toBeVisible({ timeout: 5_000 })
    })

    test('create token', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        // Go to Tokens tab
        await page.getByRole('button', { name: 'Tokens', exact: true }).click()
        await page.waitForLoadState('networkidle')

        // Create a new token
        await page.getByRole('button', { name: /Create token/i }).click()
        await page.waitForLoadState('networkidle')

        // A token row should now be visible (tokens are long hex strings)
        // A "Revoke" button should appear for the new token
        await expect(page.getByRole('button', { name: /Revoke/i }).first()).toBeVisible({ timeout: 5_000 })
    })
})

test.describe('User info card', () => {
    test('shows user login', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        const card = page.locator('aside .glass-card').first()
        await expect(card).toBeVisible({ timeout: 5_000 })
        await expect(card).toContainText('admin')
    })

    test('shows provider', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        const card = page.locator('aside .glass-card').first()
        await expect(card).toBeVisible({ timeout: 5_000 })
        await expect(card).toContainText('local')
    })

    test('admin badge shown for admin user', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        // The admin badge is a green rounded-full span with text "admin"
        const badge = page.locator('aside .glass-card .rounded-full').filter({ hasText: 'admin' })
        await expect(badge).toBeVisible({ timeout: 5_000 })
        // Verify the green styling
        await expect(badge).toHaveClass(/bg-emerald/)
    })
})

test.describe('User configuration panel', () => {
    test('shows user config labels', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        // Stats view (default) should show User Configuration panel
        const configPanel = page.locator('.glass-card').filter({ hasText: 'User Configuration' })
        await expect(configPanel).toBeVisible({ timeout: 5_000 })
        await expect(configPanel.getByText('Max File Size')).toBeVisible()
        await expect(configPanel.getByText('Max User Size')).toBeVisible()
        await expect(configPanel.getByText('Default TTL')).toBeVisible()
        await expect(configPanel.getByText('Max TTL')).toBeVisible()
    })
})

test.describe('User statistics panel', () => {
    test('shows user stats labels and values', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        const statsPanel = page.locator('.glass-card').filter({ hasText: 'User Statistics' })
        await expect(statsPanel).toBeVisible({ timeout: 5_000 })

        // Check labels
        await expect(statsPanel.getByText('Uploads', { exact: true })).toBeVisible()
        await expect(statsPanel.getByText('Files', { exact: true })).toBeVisible()
        await expect(statsPanel.getByText('Total Size', { exact: true })).toBeVisible()

        // Check that stat values are present and not NaN
        const values = statsPanel.locator('.text-2xl.font-bold')
        const count = await values.count()
        expect(count).toBe(3)

        for (let i = 0; i < count; i++) {
            const text = await values.nth(i).textContent()
            expect(text).toBeTruthy()
            expect(text).not.toBe('NaN')
        }
    })
})

test.describe('Edit account button', () => {
    test('visible for local provider', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        const btn = page.getByRole('button', { name: 'Edit account', exact: true })
        await expect(btn).toBeVisible({ timeout: 5_000 })
    })

    test('opens edit account modal', async ({ authenticatedPage: page }) => {
        await page.goto('/#/home')
        await page.waitForLoadState('networkidle')

        await page.getByRole('button', { name: 'Edit account', exact: true }).click()
        await expect(page.getByRole('heading', { name: 'Edit Account' })).toBeVisible({ timeout: 5_000 })
    })
})
