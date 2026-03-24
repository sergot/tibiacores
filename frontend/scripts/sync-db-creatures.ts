#!/usr/bin/env node
/**
 * Diffs data/creatures.txt (desired state from TibiaWiki) against
 * data/creatures-synced.txt (what is already in the DB) and generates a
 * timestamped goose migration file for any new creatures.
 *
 * Workflow:
 *   1. Run fetch-soulcore-creatures.ts  → updates data/creatures.txt
 *   2. Run download-soulcore-images.ts  → downloads GIFs for new creatures
 *   3. Run this script                  → generates backend/db/migrations/{ts}_add_new_creatures.sql
 *   4. Review the migration, then: make goose/up
 *   5. Update data/creatures-synced.txt to match data/creatures.txt
 *
 * Run: npx tsx scripts/sync-db-creatures.ts
 */

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const CREATURES_FILE = path.resolve(__dirname, '../../data/creatures.txt')
const SYNCED_FILE = path.resolve(__dirname, '../../data/creatures-synced.txt')
const MIGRATIONS_DIR = path.resolve(__dirname, '../../backend/db/migrations')

function readLines(file: string): string[] {
  return fs
    .readFileSync(file, 'utf-8')
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.length > 0)
}

function timestamp(): string {
  const now = new Date()
  const pad = (n: number, w = 2) => String(n).padStart(w, '0')
  return (
    String(now.getFullYear()) +
    pad(now.getMonth() + 1) +
    pad(now.getDate()) +
    pad(now.getHours()) +
    pad(now.getMinutes()) +
    pad(now.getSeconds())
  )
}

function escapeSql(s: string): string {
  return s.replaceAll("'", "''")
}

function generateMigration(added: string[], removed: string[]): string {
  const lines: string[] = ['-- +goose Up', '-- +goose StatementBegin']

  if (added.length > 0) {
    lines.push('')
    lines.push(`-- Add ${added.length} new creature(s)`)
    for (const name of added) {
      lines.push(`INSERT INTO creatures (name) VALUES ('${escapeSql(name)}') ON CONFLICT (name) DO NOTHING;`)
    }
  }

  if (removed.length > 0) {
    lines.push('')
    lines.push(`-- NOTE: ${removed.length} creature(s) removed from TibiaWiki.`)
    lines.push('-- Uncomment the lines below if you want to remove them from the DB.')
    lines.push('-- WARNING: This will also delete associated soul core data.')
    for (const name of removed) {
      lines.push(`-- DELETE FROM creatures WHERE name = '${escapeSql(name)}';`)
    }
  }

  lines.push('')
  lines.push('-- +goose StatementEnd')
  lines.push('')
  lines.push('-- +goose Down')
  lines.push('-- +goose StatementBegin')

  if (added.length > 0) {
    lines.push('')
    for (const name of added) {
      lines.push(`DELETE FROM creatures WHERE name = '${escapeSql(name)}';`)
    }
  }

  lines.push('')
  lines.push('-- +goose StatementEnd')

  return lines.join('\n')
}

function main() {
  if (!fs.existsSync(CREATURES_FILE)) {
    console.error(`creatures.txt not found: ${CREATURES_FILE}`)
    console.error('Run fetch-soulcore-creatures.ts first.')
    process.exit(1)
  }

  if (!fs.existsSync(SYNCED_FILE)) {
    console.error(`creatures-synced.txt not found: ${SYNCED_FILE}`)
    console.error(
      'This file should contain the creature names already present in the DB.\n' +
        'If this is your first time running this script, copy creatures.txt to creatures-synced.txt\n' +
        'to represent the current DB state, then re-run after fetching new creatures.',
    )
    process.exit(1)
  }

  const desired = new Set(readLines(CREATURES_FILE))
  const synced = new Set(readLines(SYNCED_FILE))

  const added = [...desired].filter((n) => !synced.has(n)).sort((a, b) => a.localeCompare(b))
  const removed = [...synced].filter((n) => !desired.has(n)).sort((a, b) => a.localeCompare(b))

  if (added.length === 0 && removed.length === 0) {
    console.log('DB is already in sync with creatures.txt — nothing to do.')
    return
  }

  if (added.length > 0) {
    console.log(`New creatures to add (${added.length}):`)
    for (const name of added) console.log(`  + ${name}`)
  }

  if (removed.length > 0) {
    console.log(`\nCreatures removed from TibiaWiki (${removed.length}):`)
    for (const name of removed) console.log(`  - ${name}`)
    console.log('  → These are commented out in the migration. Review before uncommenting.')
  }

  const ts = timestamp()
  const filename = `${ts}_add_new_creatures.sql`
  const outputPath = path.join(MIGRATIONS_DIR, filename)

  const content = generateMigration(added, removed)
  fs.writeFileSync(outputPath, content, 'utf-8')

  console.log(`\nMigration written: backend/db/migrations/${filename}`)
  console.log('\nNext steps:')
  console.log('  1. Review the migration file')
  console.log('  2. Run: make goose/up')
  console.log('  3. Run: cp data/creatures.txt data/creatures-synced.txt')
}

main()
