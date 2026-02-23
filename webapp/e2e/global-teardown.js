import { rmSync, readFileSync } from 'fs'

/**
 * Global teardown — cleans up the temp directory created by start-server.sh.
 */
export default function globalTeardown() {
    try {
        const dir = readFileSync('/tmp/plik-e2e-tmpdir', 'utf-8').trim()
        if (dir.startsWith('/tmp/plik-e2e.')) {
            rmSync(dir, { recursive: true, force: true })
        }
        rmSync('/tmp/plik-e2e-tmpdir', { force: true })
    } catch {
        // Already cleaned or never created — ignore
    }
}
