#!/usr/bin/env node
/**
 * Fetches the canonical list of soul core creatures from TibiaWiki
 * (Category:Soul_Cores) and writes the sorted names to data/creatures.txt.
 *
 * Run: npx tsx scripts/fetch-soulcore-creatures.ts
 */

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import axios from 'axios'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const OUTPUT_FILE = path.resolve(__dirname, '../../data/creatures.txt')

const API_BASE = 'https://tibia.fandom.com/api.php'

/**
 * Extracts the creature name from a soul core wiki page title.
 *
 * TibiaWiki uses two naming conventions:
 *   Standard:  "Acid Blob Soul Core"          → "Acid Blob"
 *   Variant:   "Nomad Soul Core (Basic)"       → "Nomad"
 *              "Nomad Soul Core (Blue)"        → "Nomad (Blue)"
 *              "Nomad Soul Core (Female)"      → "Nomad (Female)"
 */
function extractCreatureName(pageTitle: string): string | null {
  // Variant format: "X Soul Core (Y)"
  const variantMatch = pageTitle.match(/^(.+?) Soul Core \((.+?)\)$/)
  if (variantMatch) {
    const base = variantMatch[1].trim()
    const variant = variantMatch[2].trim()
    return variant.toLowerCase() === 'basic' ? base : `${base} (${variant})`
  }

  // Standard format: "X Soul Core"
  if (pageTitle.endsWith(' Soul Core')) {
    return pageTitle.slice(0, -' Soul Core'.length)
  }

  return null
}

interface CategoryMember {
  pageid: number
  ns: number
  title: string
}

interface ApiResponse {
  batchcomplete?: string
  continue?: { cmcontinue: string; continue: string }
  query: { categorymembers: CategoryMember[] }
}

async function fetchAllSoulCorePages(): Promise<string[]> {
  const names: string[] = []
  let cmcontinue: string | undefined

  do {
    const params: Record<string, string> = {
      action: 'query',
      list: 'categorymembers',
      cmtitle: 'Category:Soul_Cores',
      cmnamespace: '0',
      cmlimit: '500',
      format: 'json',
    }
    if (cmcontinue) params.cmcontinue = cmcontinue

    const url = `${API_BASE}?${new URLSearchParams(params).toString()}`
    const resp = await axios.get<ApiResponse>(url, { timeout: 15_000 })
    const members = resp.data.query.categorymembers

    for (const member of members) {
      const name = extractCreatureName(member.title)
      if (name) names.push(name)
    }

    cmcontinue = resp.data.continue?.cmcontinue
    console.log(`  Fetched ${names.length} creatures so far...`)
  } while (cmcontinue)

  return names
}

/**
 * Some soul core wiki pages exist but were never added to Category:Soul_Cores
 * (e.g. Butterfly (Red), Horse (Gray)). For creatures that appear "removed",
 * verify their wiki page still exists before treating them as truly gone.
 * Returns the subset of names whose wiki page is confirmed missing.
 */
async function filterTrulyRemoved(names: string[]): Promise<string[]> {
  if (names.length === 0) return []

  // Build page titles to check — try both naming conventions
  const toCheck = names.flatMap((name) => [
    `${name} Soul Core`, // standard: "Butterfly (Red) Soul Core"
  ])

  const BATCH = 50
  const missing: string[] = []

  for (let i = 0; i < toCheck.length; i += BATCH) {
    const batch = toCheck.slice(i, i + BATCH)
    const params = new URLSearchParams({
      action: 'query',
      titles: batch.join('|'),
      format: 'json',
    })
    const resp = await axios.get<{
      query: { pages: Record<string, { missing?: string; title: string }> }
    }>(`${API_BASE}?${params.toString()}`, { timeout: 15_000 })

    const pages = Object.values(resp.data.query.pages)
    for (let j = 0; j < batch.length; j++) {
      const page = pages[j]
      // 'missing' property is present when the page doesn't exist
      if (page.missing !== undefined) {
        // Recover the original creature name
        const title = batch[j]
        const name = extractCreatureName(title)
        if (name) missing.push(name)
      }
    }
  }

  return missing
}

async function main() {
  console.log('Fetching soul core creature list from TibiaWiki...')

  const fromCategory = await fetchAllSoulCorePages()

  let previousNames: string[] = []
  if (fs.existsSync(OUTPUT_FILE)) {
    previousNames = fs
      .readFileSync(OUTPUT_FILE, 'utf-8')
      .split('\n')
      .map((l) => l.trim())
      .filter((l) => l.length > 0)
  }

  const previousSet = new Set(previousNames)
  const categorySet = new Set(fromCategory)

  const added = fromCategory.filter((n) => !previousSet.has(n))

  // Creatures in the previous list but not in the category fetch.
  // Some wiki pages exist but are simply not tagged with Category:Soul_Cores —
  // verify each one before treating it as removed.
  const candidatesForRemoval = previousNames.filter((n) => !categorySet.has(n))

  let trulyRemoved: string[] = []
  let uncategorizedButValid: string[] = []

  if (candidatesForRemoval.length > 0) {
    console.log(
      `\nVerifying ${candidatesForRemoval.length} creature(s) not found in category...`,
    )
    trulyRemoved = await filterTrulyRemoved(candidatesForRemoval)
    const trulyRemovedSet = new Set(trulyRemoved)
    uncategorizedButValid = candidatesForRemoval.filter((n) => !trulyRemovedSet.has(n))
  }

  // Final list = category results + creatures whose wiki page still exists
  const newNames = [...new Set([...fromCategory, ...uncategorizedButValid])].sort((a, b) =>
    a.localeCompare(b),
  )

  const newContent = newNames.join('\n') + '\n'
  fs.writeFileSync(OUTPUT_FILE, newContent, 'utf-8')

  console.log(`\nTotal creatures: ${newNames.length}`)

  if (added.length > 0) {
    console.log(`\nAdded (${added.length}):`)
    for (const name of added) console.log(`  + ${name}`)
  }

  if (uncategorizedButValid.length > 0) {
    console.log(`\nKept (wiki page exists but missing from Category:Soul_Cores) (${uncategorizedButValid.length}):`)
    for (const name of uncategorizedButValid) console.log(`  ~ ${name}`)
  }

  if (trulyRemoved.length > 0) {
    console.log(`\nTruly removed (wiki page gone) (${trulyRemoved.length}):`)
    for (const name of trulyRemoved) console.log(`  - ${name}`)
  }

  if (added.length === 0 && trulyRemoved.length === 0 && uncategorizedButValid.length === 0) {
    console.log('No changes — creatures.txt is already up to date.')
  } else {
    console.log(`\nWrote ${OUTPUT_FILE}`)
  }
}

main().catch((err) => {
  console.error('Error:', err instanceof Error ? err.message : String(err))
  process.exit(1)
})
