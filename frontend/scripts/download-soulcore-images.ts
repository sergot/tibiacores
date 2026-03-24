#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import axios from 'axios'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

const CREATURES_FILE = path.resolve(__dirname, '../../data/creatures.txt')
const OUTPUT_DIR = path.resolve(__dirname, '../public/assets/soulcores')
const DELAY_MS = 500

const HEADERS = {
  'User-Agent':
    'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36',
  Accept: 'image/gif,image/*;q=0.9,*/*;q=0.8',
  Referer: 'https://tibia.fandom.com/',
}

function toSoulCoreFilename(creatureName: string): string {
  return `${creatureName.replaceAll(' ', '_')}_Soul_Core.gif`
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/** Resolves the real CDN URL via the MediaWiki API. Returns null if the file doesn't exist. */
async function resolveImageUrl(filename: string): Promise<string | null> {
  const apiUrl =
    `https://tibia.fandom.com/api.php?action=query` +
    `&titles=File:${encodeURIComponent(filename)}` +
    `&prop=imageinfo&iiprop=url&format=json`
  const resp = await axios.get<{
    query: { pages: Record<string, { missing?: boolean; imageinfo?: Array<{ url: string }> }> }
  }>(apiUrl, { timeout: 15_000, headers: HEADERS })
  const pages = resp.data.query.pages
  const page = Object.values(pages)[0]
  if (page.missing || !page.imageinfo?.length) return null
  return page.imageinfo[0].url
}

async function downloadImage(filename: string): Promise<boolean> {
  const cdnUrl = await resolveImageUrl(filename)
  if (!cdnUrl) return false

  const response = await axios.get(`${cdnUrl}&format=original`, {
    responseType: 'arraybuffer',
    timeout: 15_000,
    headers: HEADERS,
  })
  const outputPath = path.join(OUTPUT_DIR, filename)
  fs.writeFileSync(outputPath, Buffer.from(response.data as ArrayBuffer))
  return true
}

async function main() {
  if (!fs.existsSync(CREATURES_FILE)) {
    console.error(`Creatures file not found: ${CREATURES_FILE}`)
    process.exit(1)
  }

  fs.mkdirSync(OUTPUT_DIR, { recursive: true })

  const creatures = fs
    .readFileSync(CREATURES_FILE, 'utf-8')
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.length > 0)

  console.log(`Downloading soul core images for ${creatures.length} creatures...`)

  const failed: string[] = []

  for (let i = 0; i < creatures.length; i++) {
    const name = creatures[i]
    const filename = toSoulCoreFilename(name)
    const outputPath = path.join(OUTPUT_DIR, filename)

    if (fs.existsSync(outputPath)) {
      console.log(`[${i + 1}/${creatures.length}] Skipping (already exists): ${filename}`)
      continue
    }

    process.stdout.write(`[${i + 1}/${creatures.length}] Downloading: ${filename} ... `)

    try {
      const ok = await downloadImage(filename)
      if (ok) {
        console.log('OK')
      } else {
        console.log('NOT FOUND (file missing on wiki)')
        failed.push(name)
      }
    } catch (err) {
      console.log(`ERROR: ${err instanceof Error ? err.message : String(err)}`)
      failed.push(name)
    }

    if (i < creatures.length - 1) {
      await sleep(DELAY_MS)
    }
  }

  console.log('\nDone.')
  if (failed.length > 0) {
    console.log(`\nFailed to download (${failed.length}):`)
    for (const name of failed) {
      console.log(`  - ${name}`)
    }
  } else {
    console.log('All images downloaded successfully.')
  }
}

main()
