// Dynamic sitemap generation utilities for TibiaCores
// This can be used to generate more comprehensive sitemaps with dynamic content

interface SitemapUrl {
  loc: string
  lastmod?: string
  changefreq?: 'always' | 'hourly' | 'daily' | 'weekly' | 'monthly' | 'yearly' | 'never'
  priority?: number
}

export class SitemapGenerator {
  private baseUrl: string
  private urls: SitemapUrl[] = []

  constructor(baseUrl: string = 'https://tibiacores.com') {
    this.baseUrl = baseUrl
  }

  addUrl(url: SitemapUrl) {
    this.urls.push(url)
  }

  addStaticPages() {
    const staticPages = [
      { loc: '/', changefreq: 'daily' as const, priority: 1.0 },
      { loc: '/about', changefreq: 'weekly' as const, priority: 0.8 },
      { loc: '/blog', changefreq: 'weekly' as const, priority: 0.8 },
      { loc: '/highscores', changefreq: 'daily' as const, priority: 0.7 },
      { loc: '/sponsor', changefreq: 'monthly' as const, priority: 0.6 },
      { loc: '/privacy', changefreq: 'monthly' as const, priority: 0.5 },
    ]

    staticPages.forEach(page => {
      this.addUrl({
        loc: `${this.baseUrl}${page.loc}`,
        changefreq: page.changefreq,
        priority: page.priority,
        lastmod: new Date().toISOString().split('T')[0]
      })
    })
  }

  async addBlogPosts() {
    try {
      const response = await fetch('/assets/blog/index.json')
      if (!response.ok) return

      const posts = await response.json() as Array<{
        id: string
        date: string
      }>
      posts.forEach((post) => {
        this.addUrl({
          loc: `${this.baseUrl}/blog/${post.id}`,
          lastmod: post.date,
          changefreq: 'monthly',
          priority: 0.6
        })
      })
    } catch (error) {
      console.warn('Failed to load blog posts for sitemap:', error)
    }
  }



  generateXML(): string {
    const urlElements = this.urls.map(url => {
      let urlXml = `  <url>\n    <loc>${url.loc}</loc>`

      if (url.lastmod) {
        urlXml += `\n    <lastmod>${url.lastmod}</lastmod>`
      }

      if (url.changefreq) {
        urlXml += `\n    <changefreq>${url.changefreq}</changefreq>`
      }

      if (url.priority !== undefined) {
        urlXml += `\n    <priority>${url.priority}</priority>`
      }

      urlXml += '\n  </url>'
      return urlXml
    }).join('\n')

    return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urlElements}
</urlset>`
  }

  async generateComplete(): Promise<string> {
    this.urls = [] // Reset URLs
    this.addStaticPages()
    await this.addBlogPosts()
    return this.generateXML()
  }
}

// Utility function to generate a complete sitemap
export async function generateSitemap(baseUrl?: string): Promise<string> {
  const generator = new SitemapGenerator(baseUrl)
  return generator.generateComplete()
}

// Function to generate robots.txt content
export function generateRobotsTxt(baseUrl: string = 'https://tibiacores.com'): string {
  return `# Allow all crawlers
User-agent: *
Allow: /

# Sitemap location
Sitemap: ${baseUrl}/sitemap.xml

# Disallow admin and private routes
Disallow: /admin/
Disallow: /profile/
Disallow: /oauth/
Disallow: /signin
Disallow: /signup
Disallow: /lists/*/join
Disallow: /characters/claim
Disallow: /characters/details

# Crawl delay for politeness
Crawl-delay: 1

# Allow access to public character profiles and nested paths
Allow: /characters/public/*`
}
