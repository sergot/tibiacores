#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { glob } from 'glob';

interface TranslationObject {
    [key: string]: string | TranslationObject;
}

function extractKeys(obj: TranslationObject, prefix: string = ''): Set<string> {
    const keys = new Set<string>();
    
    for (const [key, value] of Object.entries(obj)) {
        const fullKey = prefix ? `${prefix}.${key}` : key;
        
        if (typeof value === 'object' && value !== null) {
            const nestedKeys = extractKeys(value as TranslationObject, fullKey);
            nestedKeys.forEach(k => keys.add(k));
        } else {
            keys.add(fullKey);
        }
    }
    
    return keys;
}

async function checkKeyUsage(key: string, frontendDir: string): Promise<boolean> {
    const patterns = [
        `t("${key}")`,
        `t('${key}')`,
        `t(\`${key}\`)`,
        `useTranslation("${key}")`,
        `useTranslation('${key}')`,
        `useTranslation(\`${key}\`)`,
        `{t("${key}")}`,
        `{t('${key}')}`,
        `{t(\`${key}\`)}`,
        `$t("${key}")`,
        `$t('${key}')`,
        `$t(\`${key}\`)`,
        `{ $t("${key}") }`,
        `{ $t('${key}') }`,
        `{ $t(\`${key}\`) }`,
        `i18n.t("${key}")`,
        `i18n.t('${key}')`,
        `i18n.t(\`${key}\`)`,
        `i18n.global.t("${key}")`,
        `i18n.global.t('${key}')`,
        `i18n.global.t(\`${key}\`)`
    ];

    try {
        const files = await glob('**/*.{ts,tsx,js,jsx,vue}', {
            cwd: frontendDir,
            ignore: ['**/node_modules/**', '**/dist/**']
        });

        for (const file of files) {
            const filePath = path.join(frontendDir, file);
            try {
                const content = await fs.promises.readFile(filePath, 'utf-8');
                if (patterns.some(pattern => content.includes(pattern))) {
                    return true;
                }
            } catch (error) {
                console.error(`Error reading file ${filePath}:`, error);
            }
        }
    } catch (error) {
        console.error('Error scanning frontend directory:', error);
        return false;
    }
    return false;
}

async function findUsedTranslationKeys(frontendDir: string): Promise<Set<string>> {
    const usedKeys = new Set<string>();
    const translationPattern = /(?:t|useTranslation|\$t|i18n\.(?:global\.)?t)\(['"`]([a-z][a-z0-9]*(?:\.[a-z][a-z0-9]*)+)['"`](?:\s*,\s*\{[^}]*\})?/gi;
    
    try {
        const files = await glob('**/*.{ts,tsx,js,jsx,vue}', {
            cwd: frontendDir,
            ignore: ['**/node_modules/**', '**/dist/**']
        });

        for (const file of files) {
            const filePath = path.join(frontendDir, file);
            try {
                const content = await fs.promises.readFile(filePath, 'utf-8');
                let match;
                while ((match = translationPattern.exec(content)) !== null) {
                    if (!match[1].includes('${') && !match[1].includes('}')) {
                        usedKeys.add(match[1]);
                    }
                }
            } catch (error) {
                console.error(`Error reading file ${filePath}:`, error);
            }
        }
    } catch (error) {
        console.error('Error scanning frontend directory:', error);
    }
    return usedKeys;
}

function getValueFromPath(obj: TranslationObject, path: string): string | undefined {
    const parts = path.split('.');
    let current: any = obj;
    
    for (const part of parts) {
        if (current && typeof current === 'object' && part in current) {
            current = current[part];
        } else {
            return undefined;
        }
    }
    
    return typeof current === 'string' ? current : undefined;
}

function setValue(obj: TranslationObject, key: string, value: string | TranslationObject): void {
    const parts = key.split('.');
    let current: TranslationObject = obj;

    for (let i = 0; i < parts.length - 1; i++) {
        const part = parts[i];
        if (part === '__proto__' || part === 'constructor') {
            return;
        }
        if (!(part in current)) {
            current[part] = {};
        }
        current = current[part] as TranslationObject;
    }

    const lastPart = parts[parts.length - 1];
    if (lastPart !== '__proto__' && lastPart !== 'constructor') {
        current[lastPart] = value;
    }
}

function cleanTranslations(translations: TranslationObject, usedKeys: Set<string>): TranslationObject {
    const cleaned: TranslationObject = {};
    
    for (const key of Array.from(usedKeys).sort()) {
        const value = getValueFromPath(translations, key);
        if (value !== undefined) {
            setValue(cleaned, key, value);
        }
    }
    
    return cleaned;
}

async function main() {
    // Get the project root directory (parent of frontend directory)
    const scriptDir = path.dirname(fileURLToPath(import.meta.url));
    const frontendDir = path.resolve(scriptDir, '..');
    const localesDir = path.join(frontendDir, 'src', 'i18n', 'locales');

    // Check if directories exist
    if (!fs.existsSync(frontendDir)) {
        console.error(`Error: Frontend directory not found at ${frontendDir}`);
        process.exit(1);
    }
    if (!fs.existsSync(localesDir)) {
        console.error(`Error: Locales directory not found at ${localesDir}`);
        process.exit(1);
    }

    // Initialize language dictionaries
    const languages = ['en', 'de', 'es', 'pl'];
    const translations: Record<string, TranslationObject> = {};

    // Load translation files
    for (const lang of languages) {
        const filePath = path.join(localesDir, `${lang}.json`);
        try {
            const content = await fs.promises.readFile(filePath, 'utf-8');
            translations[lang] = JSON.parse(content);
        } catch (error) {
            console.error(`Error loading ${lang}.json:`, error);
            translations[lang] = {};
        }
    }

    // Get all keys from English file
    const enKeys = extractKeys(translations['en']);

    // Find all translation keys used in the code
    const usedKeys = await findUsedTranslationKeys(frontendDir);
    const allTranslationKeys = new Set<string>();
    
    // Combine all keys from all translation files
    for (const lang of languages) {
        const langKeys = extractKeys(translations[lang]);
        langKeys.forEach(key => allTranslationKeys.add(key));
    }

    // Find missing translations that are actually used in code
    const missingUsedKeys = new Set<string>();
    for (const key of usedKeys) {
        if (!allTranslationKeys.has(key)) {
            missingUsedKeys.add(key);
        }
    }

    // Find unused keys that should be removed
    const unusedKeys = new Set<string>();
    for (const key of allTranslationKeys) {
        if (!usedKeys.has(key)) {
            unusedKeys.add(key);
        }
    }

    // Print summary
    console.log('\nTranslation Key Summary:');
    console.log('='.repeat(50));
    console.log(`Total keys in English file: ${enKeys.size}`);
    console.log(`Total unique keys across all languages: ${allTranslationKeys.size}`);
    console.log(`Total keys used in code: ${usedKeys.size}`);
    console.log(`Missing translations used in code: ${missingUsedKeys.size}`);
    console.log(`Unused translations: ${unusedKeys.size}`);

    if (missingUsedKeys.size > 0) {
        console.log('\nMissing Translations Used in Code:');
        console.log('='.repeat(50));
        for (const key of Array.from(missingUsedKeys).sort()) {
            console.log(`- ${key}`);
        }
    }

    // Check for missing translations in each language
    console.log('\nMissing Translations:');
    console.log('='.repeat(50));
    for (const lang of languages) {
        if (lang === 'en') continue; // Skip English as it's our reference
        
        const langKeys = extractKeys(translations[lang]);
        const missingKeys = new Set<string>();
        
        for (const key of enKeys) {
            if (!langKeys.has(key)) {
                missingKeys.add(key);
            }
        }
        
        if (missingKeys.size > 0) {
            console.log(`\nMissing translations in ${lang}:`);
            for (const key of Array.from(missingKeys).sort()) {
                console.log(`- ${key}`);
            }
        } else {
            console.log(`\nNo missing translations in ${lang}`);
        }
    }

    // Create cleaned translation files
    for (const lang of languages) {
        const originalKeys = extractKeys(translations[lang]);
        const cleaned = cleanTranslations(translations[lang], usedKeys);
        const cleanedKeys = extractKeys(cleaned);

        // Only write and report if keys were actually removed
        if (originalKeys.size > cleanedKeys.size) {
            const outputPath = path.join(localesDir, `${lang}.json`);
            try {
                await fs.promises.writeFile(
                    outputPath,
                    JSON.stringify(cleaned, null, 2) + '\n',
                    'utf-8'
                );
                console.log(
                    `\nCleaned ${outputPath} (removed ${originalKeys.size - cleanedKeys.size} unused keys)`
                );
            } catch (error) {
                console.error(`Error writing ${lang}.json:`, error);
            }
        }
    }

    if (unusedKeys.size > 0) {
        console.log('\nUnused Translation Keys:');
        console.log('='.repeat(50));
        for (const key of Array.from(unusedKeys).sort()) {
            console.log(`- ${key}`);
        }
    }
}

main().catch(error => {
    console.error('Script failed:', error);
    process.exit(1);
}); 