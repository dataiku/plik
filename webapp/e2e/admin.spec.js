import { test, expect, ADMIN_LOGIN } from './fixtures.js'

test.describe('Admin view', () => {
    test('stats tab shows server info', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        // Stats view is shown by default — should show "Server Configuration" or "Server Statistics"
        await expect(page.getByText('Server Configuration')).toBeVisible({ timeout: 5_000 })
        await expect(page.getByText('Server Statistics')).toBeVisible()
    })

    test('uploads tab shows upload list', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        // Click Uploads sidebar nav button
        await page.getByRole('button', { name: 'Uploads', exact: true }).click()
        await page.waitForLoadState('networkidle')

        // Should show upload list (may be empty with "No uploads" or have entries)
        const content = page.getByText('No uploads').or(page.locator('.glass-card'))
        await expect(content.first()).toBeVisible({ timeout: 5_000 })
    })

    test('users tab shows admin user', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        // Click Users sidebar nav button
        await page.getByRole('button', { name: 'Users', exact: true }).click()
        await page.waitForLoadState('networkidle')

        // Should show the admin user login in the main content area
        // The admin user name appears twice in the same card (login + admin badge),
        // so use .first() to avoid strict mode violation
        const mainContent = page.locator('main')
        await expect(mainContent.getByText(ADMIN_LOGIN).first()).toBeVisible({ timeout: 5_000 })
    })

    test('create and delete user', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        // Click "Create User" sidebar button to open modal
        await page.getByRole('button', { name: 'Create User', exact: true }).click()

        // Wait for modal to appear
        await expect(page.getByText('Create User').nth(1)).toBeVisible({ timeout: 3_000 })

        // Fill the modal form
        await page.locator('input[placeholder="min 4 chars"]').fill('testuser')
        await page.locator('input[placeholder="min 8 chars"]').fill('testpass123')

        // Click the Create button in the modal
        await page.getByRole('button', { name: 'Create', exact: true }).click()
        await page.waitForLoadState('networkidle')

        // Switch to Users tab to see the new user
        await page.getByRole('button', { name: 'Users', exact: true }).click()
        await page.waitForLoadState('networkidle')

        // Wait for the new user to appear in the main content area
        const mainContent = page.locator('main')
        await expect(mainContent.getByText('testuser').first()).toBeVisible({ timeout: 5_000 })

        // Delete the user — find the Delete button in the testuser's row
        const userCards = mainContent.locator('.glass-card').filter({ hasText: 'testuser' })
        await userCards.first().getByRole('button', { name: 'Delete' }).click()

        // Confirm deletion in the dialog (use exact match to avoid matching disabled Delete buttons)
        await page.getByRole('button', { name: 'Confirm', exact: true }).click()
        await page.waitForLoadState('networkidle')

        // User should no longer appear (give some time for UI to update)
        await expect(mainContent.getByText('testuser')).not.toBeVisible({ timeout: 5_000 })
    })
})

test.describe('Admin server info card', () => {
    test('shows version', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        // Version should be displayed in sidebar (e.g. "v1.4.0-rc4")
        const versionText = page.locator('aside .glass-card .font-mono')
        await expect(versionText).toBeVisible({ timeout: 5_000 })
        await expect(versionText).toContainText('v')
    })

    test('shows Go version', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        // Go version text below the version (e.g. "go1.24.0")
        const sidebar = page.locator('aside .glass-card').first()
        await expect(sidebar).toContainText('go1.', { timeout: 5_000 })
    })

    test('release and mint badges green when true', async ({ authenticatedPage: page, withVersion }) => {
        await withVersion({ isRelease: true, isMint: true })
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        const sidebar = page.locator('aside .glass-card').first()
        const releaseBadge = sidebar.locator('.rounded-full').filter({ hasText: 'release' })
        const mintBadge = sidebar.locator('.rounded-full').filter({ hasText: 'mint' })
        await expect(releaseBadge).toBeVisible({ timeout: 5_000 })
        await expect(releaseBadge).toHaveClass(/bg-emerald/)
        await expect(mintBadge).toBeVisible()
        await expect(mintBadge).toHaveClass(/bg-emerald/)
    })

    test('release and mint badges red when false', async ({ authenticatedPage: page, withVersion }) => {
        await withVersion({ isRelease: false, isMint: false })
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        const sidebar = page.locator('aside .glass-card').first()
        const releaseBadge = sidebar.locator('.rounded-full').filter({ hasText: 'release' })
        const mintBadge = sidebar.locator('.rounded-full').filter({ hasText: 'mint' })
        await expect(releaseBadge).toBeVisible({ timeout: 5_000 })
        await expect(releaseBadge).toHaveClass(/bg-red/)
        await expect(mintBadge).toBeVisible()
        await expect(mintBadge).toHaveClass(/bg-red/)
    })
})

test.describe('Admin server configuration', () => {
    test('shows Max File Size', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        const configPanel = page.locator('.glass-card').filter({ hasText: 'Server Configuration' })
        await expect(configPanel).toBeVisible({ timeout: 5_000 })
        await expect(configPanel.getByText('Max File Size')).toBeVisible()
    })

    test('shows Max TTL', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        const configPanel = page.locator('.glass-card').filter({ hasText: 'Server Configuration' })
        await expect(configPanel).toBeVisible({ timeout: 5_000 })
        await expect(configPanel.getByText('Max TTL')).toBeVisible()
    })

    test('all config values are formatted', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        const configPanel = page.locator('.glass-card').filter({ hasText: 'Server Configuration' })
        await expect(configPanel).toBeVisible({ timeout: 5_000 })

        // All 4 labels should be visible with formatted values
        for (const label of ['Max File Size', 'Max User Size', 'Default TTL', 'Max TTL']) {
            await expect(configPanel.getByText(label)).toBeVisible()
        }

        // The config values should have formatted content (not empty)
        const values = configPanel.locator('.text-surface-200.font-medium')
        const count = await values.count()
        expect(count).toBe(4)

        for (let i = 0; i < count; i++) {
            const text = await values.nth(i).textContent()
            expect(text).toBeTruthy()
        }
    })
})

test.describe('Admin server statistics', () => {
    test('shows all stat labels', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        const statsPanel = page.locator('.glass-card').filter({ hasText: 'Server Statistics' })
        await expect(statsPanel).toBeVisible({ timeout: 5_000 })

        for (const label of ['Users', 'Uploads', 'Anonymous Uploads', 'Files', 'Total Size', 'Anonymous Size']) {
            await expect(statsPanel.getByText(label, { exact: true })).toBeVisible()
        }
    })

    test('stat values are not empty or NaN', async ({ authenticatedPage: page }) => {
        await page.goto('/#/admin')
        await page.waitForLoadState('networkidle')

        const statsPanel = page.locator('.glass-card').filter({ hasText: 'Server Statistics' })
        await expect(statsPanel).toBeVisible({ timeout: 5_000 })

        // Check that the bold stat values exist and aren't NaN
        const values = statsPanel.locator('.text-2xl.font-bold')
        const count = await values.count()
        expect(count).toBe(6)

        for (let i = 0; i < count; i++) {
            const text = await values.nth(i).textContent()
            expect(text).toBeTruthy()
            expect(text).not.toBe('NaN')
        }
    })
})

