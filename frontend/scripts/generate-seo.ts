#!/usr/bin/env tsx
// Script to generate sitemap.xml and robots.txt for TibiaCores

import { writeFileSync, readFileSync, existsSync } from 'fs'
import { join } from 'path'
import { generateRobotsTxt, SitemapGenerator } from '../src/utils/sitemap.js'

const PUBLIC_DIR = join(process.cwd(), 'public')
const BASE_URL = 'https://tibiacores.com'

async function generateSitemapWithLocalData(baseUrl: string): Promise<string> {
  const generator = new SitemapGenerator(baseUrl)

  // Add static pages
  generator.addStaticPages()

  // Add blog posts from local file system
  const blogIndexPath = join(PUBLIC_DIR, 'assets', 'blog', 'index.json')
  if (existsSync(blogIndexPath)) {
    try {
      const blogContent = readFileSync(blogIndexPath, 'utf-8')
      const posts = JSON.parse(blogContent) as Array<{
        id: string
        date: string
      }>

      posts.forEach((post) => {
        generator.addUrl({
          loc: `${baseUrl}/blog/${post.id}`,
          lastmod: post.date,
          changefreq: 'monthly',
          priority: 0.6
        })
      })
      console.log(`📝 Added ${posts.length} blog posts to sitemap`)
    } catch (error) {
      console.warn('⚠️ Failed to load blog posts:', error)
    }
  } else {
    console.log('📝 No blog index found, skipping blog posts')
  }

  return generator.generateXML()
}

async function generateSEOFiles() {
  console.log('🚀 Generating SEO files...')

  try {
    // Generate sitemap.xml
    console.log('📄 Generating sitemap.xml...')
    const sitemapContent = await generateSitemapWithLocalData(BASE_URL)
    const sitemapPath = join(PUBLIC_DIR, 'sitemap.xml')
    writeFileSync(sitemapPath, sitemapContent, 'utf-8')
    console.log(`✅ Generated sitemap.xml at ${sitemapPath}`)

    // Generate robots.txt
    console.log('🤖 Generating robots.txt...')
    const robotsContent = generateRobotsTxt(BASE_URL)
    const robotsPath = join(PUBLIC_DIR, 'robots.txt')
    writeFileSync(robotsPath, robotsContent, 'utf-8')
    console.log(`✅ Generated robots.txt at ${robotsPath}`)

    console.log('🎉 SEO files generated successfully!')
  } catch (error) {
    console.error('❌ Error generating SEO files:', error)
    process.exit(1)
  }
}

// Run the script
generateSEOFiles()
