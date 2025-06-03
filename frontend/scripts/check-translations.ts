#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { glob } from 'glob'

interface TranslationObject {
  [key: string]: string | TranslationObject
}

// Parse command line arguments
const args = process.argv.slice(2)
const checkOnly = args.includes('--check-only') || args.includes('-c')

function extractKeys(obj: TranslationObject, prefix: string = ''): Set<string> {
  const keys = new Set<string>()

  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key

    if (typeof value === 'object' && value !== null) {
      const nestedKeys = extractKeys(value as TranslationObject, fullKey)
      nestedKeys.forEach((k) => keys.add(k))
    } else {
      keys.add(fullKey)
    }
  }

  return keys
}

async function findUsedTranslationKeys(frontendDir: string): Promise<Set<string>> {
  const usedKeys = new Set<string>()

  // Standard translation keys pattern (for quoted strings)
  const standardTranslationPattern =
    /(?<=(?:^|[^\w$])(?:t|useTranslation|\$t|i18n\.(?:global\.)?t)\s*\(['"`])([a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)*)(?=['"`](?:\s*\)|\s*,))/gm

  // Template literal with dynamic expression pattern
  // For cases like: t(`profile.email.status.${emailVerified ? 'verified' : 'notVerified'}`)
  const templateLiteralPattern =
    /(?:t|useTranslation|\$t|i18n\.(?:global\.)?t)\s*\(`(.*?)\$\{.*?['"]([a-zA-Z0-9_]+)['"].*?(?:['"]([a-zA-Z0-9_]+)['"]).?\}`\)/gm

  try {
    const files = await glob('**/*.{ts,tsx,js,jsx,vue}', {
      cwd: frontendDir,
      ignore: ['**/node_modules/**', '**/dist/**'],
    })

    for (const file of files) {
      const filePath = path.join(frontendDir, file)
      try {
        const content = await fs.promises.readFile(filePath, 'utf-8')
        let match

        // Match standard translation keys
        while ((match = standardTranslationPattern.exec(content)) !== null) {
          if (!match[1].includes('${') && !match[1].includes('}')) {
            usedKeys.add(match[1])
          }
        }

        // Match template literal patterns
        while ((match = templateLiteralPattern.exec(content)) !== null) {
          if (match[1]) {
            const prefix = match[1]
            const firstSuffix = match[2]

            // Add the first possible key (e.g., profile.email.status.verified)
            if (prefix && firstSuffix) {
              usedKeys.add(`${prefix}${firstSuffix}`)
            }

            // If there's a second suffix in the ternary (the part after the colon),
            // add it as another key (e.g., profile.email.status.notVerified)
            if (prefix && match[3]) {
              usedKeys.add(`${prefix}${match[3]}`)
            }
          }
        }
      } catch (error) {
        console.error(`Error reading file ${filePath}:`, error)
      }
    }
  } catch (error) {
    console.error('Error scanning frontend directory:', error)
  }
  return usedKeys
}

function getValueFromPath(obj: TranslationObject, path: string): string | undefined {
  const parts = path.split('.')
  let current: TranslationObject | string = obj

  for (const part of parts) {
    if (current && typeof current === 'object' && part in current) {
      current = current[part]
    } else {
      return undefined
    }
  }

  return typeof current === 'string' ? current : undefined
}

function setValue(obj: TranslationObject, key: string, value: string | TranslationObject): void {
  const parts = key.split('.')
  let current: TranslationObject = obj

  for (let i = 0; i < parts.length - 1; i++) {
    const part = parts[i]
    if (part === '__proto__' || part === 'constructor') {
      return
    }
    if (!(part in current)) {
      current[part] = {}
    }
    current = current[part] as TranslationObject
  }

  const lastPart = parts[parts.length - 1]
  if (lastPart !== '__proto__' && lastPart !== 'constructor') {
    current[lastPart] = value
  }
}

function cleanTranslations(
  translations: TranslationObject,
  usedKeys: Set<string>,
): TranslationObject {
  const cleaned: TranslationObject = {}

  for (const key of Array.from(usedKeys).sort()) {
    const value = getValueFromPath(translations, key)
    if (value !== undefined) {
      setValue(cleaned, key, value)
    }
  }

  return cleaned
}

function checkInterpolationFormats(translations: Record<string, TranslationObject>): {
  hasIssues: boolean,
  issues: { lang: string, key: string, value: string }[]
} {
  const issues: { lang: string, key: string, value: string }[] = []
  let hasIssues = false

  // Regex to find {{variable}} format (incorrect for Vue I18n)
  const incorrectFormatRegex = /\{\{\s*[a-zA-Z0-9_]+\s*\}\}/g

  for (const [lang, langTranslations] of Object.entries(translations)) {
    const extractAndCheckValues = (obj: TranslationObject, prefix: string = '') => {
      for (const [key, value] of Object.entries(obj)) {
        const fullKey = prefix ? `${prefix}.${key}` : key

        if (typeof value === 'string') {
          if (incorrectFormatRegex.test(value)) {
            hasIssues = true
            issues.push({
              lang,
              key: fullKey,
              value
            })
          }
        } else if (value !== null && typeof value === 'object') {
          extractAndCheckValues(value as TranslationObject, fullKey)
        }
      }
    }

    extractAndCheckValues(langTranslations)
  }

  return { hasIssues, issues }
}

async function main() {
  // Get the project root directory (parent of frontend directory)
  const scriptDir = path.dirname(fileURLToPath(import.meta.url))
  const frontendDir = path.resolve(scriptDir, '..')
  const localesDir = path.join(frontendDir, 'src', 'i18n', 'locales')

  // Tracking if any issues were found that should cause the script to exit with an error
  let hasIssues = false

  // Check if directories exist
  if (!fs.existsSync(frontendDir)) {
    console.error(`Error: Frontend directory not found at ${frontendDir}`)
    process.exit(1)
  }
  if (!fs.existsSync(localesDir)) {
    console.error(`Error: Locales directory not found at ${localesDir}`)
    process.exit(1)
  }

  // Initialize language dictionaries
  const languages = ['en', 'de', 'es', 'pl', 'pt']
  const translations: Record<string, TranslationObject> = {}

  // Load translation files
  for (const lang of languages) {
    const filePath = path.join(localesDir, `${lang}.json`)
    try {
      const content = await fs.promises.readFile(filePath, 'utf-8')
      translations[lang] = JSON.parse(content)
    } catch (error) {
      console.error(`Error loading ${lang}.json:`, error)
      translations[lang] = {}
    }
  }

  // Get all keys from English file
  const enKeys = extractKeys(translations['en'])

  // Find all translation keys used in the code
  const usedKeys = await findUsedTranslationKeys(frontendDir)
  const allTranslationKeys = new Set<string>()

  // Combine all keys from all translation files
  for (const lang of languages) {
    const langKeys = extractKeys(translations[lang])
    langKeys.forEach((key) => allTranslationKeys.add(key))
  }

  // Find missing translations that are actually used in code
  const missingUsedKeys = new Set<string>()
  for (const key of usedKeys) {
    if (!allTranslationKeys.has(key)) {
      missingUsedKeys.add(key)
      hasIssues = true
    }
  }

  // Find unused keys that should be removed
  const unusedKeys = new Set<string>()
  for (const key of allTranslationKeys) {
    if (!usedKeys.has(key)) {
      unusedKeys.add(key)
    }
  }

  // Print summary
  console.log('\nTranslation Key Summary:')
  console.log('='.repeat(50))
  console.log(`Total keys in English file: ${enKeys.size}`)
  console.log(`Total unique keys across all languages: ${allTranslationKeys.size}`)
  console.log(`Total keys used in code: ${usedKeys.size}`)
  console.log(`Missing translations used in code: ${missingUsedKeys.size}`)
  console.log(`Unused translations: ${unusedKeys.size}`)

  if (missingUsedKeys.size > 0) {
    console.log('\nMissing Translations Used in Code:')
    console.log('='.repeat(50))
    for (const key of Array.from(missingUsedKeys).sort()) {
      console.log(`- ${key}`)
    }
  }

  // Check for missing translations in each language
  console.log('\nMissing Translations:')
  console.log('='.repeat(50))
  for (const lang of languages) {
    if (lang === 'en') continue // Skip English as it's our reference

    const langKeys = extractKeys(translations[lang])
    const missingKeys = new Set<string>()

    for (const key of enKeys) {
      if (!langKeys.has(key)) {
        missingKeys.add(key)
        hasIssues = true
      }
    }

    if (missingKeys.size > 0) {
      console.log(`\nMissing translations in ${lang}:`)
      for (const key of Array.from(missingKeys).sort()) {
        console.log(`- ${key}`)
      }
    } else {
      console.log(`\nNo missing translations in ${lang}`)
    }
  }

  // Check for incorrect interpolation formats
  const interpolationCheck = checkInterpolationFormats(translations)
  if (interpolationCheck.hasIssues) {
    console.log('\nIncorrect Interpolation Formats:')
    console.log('='.repeat(50))
    for (const issue of interpolationCheck.issues) {
      console.log(`Language: ${issue.lang}, Key: ${issue.key}, Value: ${issue.value}`)
    }
    hasIssues = true
  }

  // Create cleaned translation files if not in check-only mode
  if (!checkOnly) {
    for (const lang of languages) {
      const originalKeys = extractKeys(translations[lang])
      const cleaned = cleanTranslations(translations[lang], usedKeys)
      const cleanedKeys = extractKeys(cleaned)

      // Only write and report if keys were actually removed
      if (originalKeys.size > cleanedKeys.size) {
        const outputPath = path.join(localesDir, `${lang}.json`)
        try {
          await fs.promises.writeFile(outputPath, JSON.stringify(cleaned, null, 2) + '\n', 'utf-8')
          console.log(
            `\nCleaned ${outputPath} (removed ${originalKeys.size - cleanedKeys.size} unused keys)`,
          )
        } catch (error) {
          console.error(`Error writing ${lang}.json:`, error)
        }
      }
    }
  } else if (unusedKeys.size > 0) {
    console.log('\nNote: Unused translations were not removed because --check-only flag was used.')
  }

  if (unusedKeys.size > 0) {
    console.log('\nUnused Translation Keys:')
    console.log('='.repeat(50))
    for (const key of Array.from(unusedKeys).sort()) {
      console.log(`- ${key}`)
    }
  }

  // Exit with error code if issues were found
  if (hasIssues) {
    console.error('\n⚠️ Translation issues found. Please fix missing translations.')
    process.exit(1)
  } else {
    console.log('\n✅ No translation issues found.')
  }
}

main().catch((error) => {
  console.error('Script failed:', error)
  process.exit(1)
})
