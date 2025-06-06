import { useI18n } from 'vue-i18n'

interface SEOData {
  title: string
  description: string
  keywords?: string
  ogTitle?: string
  ogDescription?: string
  ogImage?: string
  ogType?: string
  twitterCard?: string
  canonical?: string
  noindex?: boolean
}

export function useSEO() {
  const { locale } = useI18n()

  const updateMetaTags = (seoData: SEOData) => {
    // Update document title
    document.title = seoData.title

    // Helper function to update or create meta tags
    const updateMetaTag = (selector: string, content: string) => {
      let element = document.querySelector(selector) as HTMLMetaElement
      if (!element) {
        element = document.createElement('meta')
        if (selector.includes('property=')) {
          element.setAttribute('property', selector.match(/property="([^"]+)"/)?.[1] || '')
        } else if (selector.includes('name=')) {
          element.setAttribute('name', selector.match(/name="([^"]+)"/)?.[1] || '')
        }
        document.head.appendChild(element)
      }
      element.setAttribute('content', content)
    }

    // Update basic meta tags
    updateMetaTag('meta[name="description"]', seoData.description)
    if (seoData.keywords) {
      updateMetaTag('meta[name="keywords"]', seoData.keywords)
    }

    // Update Open Graph tags
    updateMetaTag('meta[property="og:title"]', seoData.ogTitle || seoData.title)
    updateMetaTag('meta[property="og:description"]', seoData.ogDescription || seoData.description)
    updateMetaTag('meta[property="og:type"]', seoData.ogType || 'website')

    if (seoData.ogImage) {
      updateMetaTag('meta[property="og:image"]', seoData.ogImage)
    }

    // Update Twitter Card tags
    updateMetaTag('meta[name="twitter:title"]', seoData.ogTitle || seoData.title)
    updateMetaTag('meta[name="twitter:description"]', seoData.ogDescription || seoData.description)
    updateMetaTag('meta[name="twitter:card"]', seoData.twitterCard || 'summary')

    // Update canonical URL
    if (seoData.canonical) {
      let canonicalLink = document.querySelector('link[rel="canonical"]') as HTMLLinkElement
      if (!canonicalLink) {
        canonicalLink = document.createElement('link')
        canonicalLink.rel = 'canonical'
        document.head.appendChild(canonicalLink)
      }
      canonicalLink.href = seoData.canonical
    }

    // Handle noindex
    if (seoData.noindex) {
      updateMetaTag('meta[name="robots"]', 'noindex, nofollow')
    } else {
      const robotsTag = document.querySelector('meta[name="robots"]')
      if (robotsTag) {
        robotsTag.remove()
      }
    }

    // Update language-specific Open Graph locales
    updateMetaTag('meta[property="og:locale"]', getOGLocale(locale.value))
  }

  const getOGLocale = (locale: string): string => {
    const localeMap: Record<string, string> = {
      en: 'en_US',
      pl: 'pl_PL',
      de: 'de_DE',
      es: 'es_ES',
      pt: 'pt_BR'
    }
    return localeMap[locale] || 'en_US'
  }

  const generateStructuredData = (type: string, data: Record<string, unknown>) => {
    const structuredData = {
      '@context': 'https://schema.org',
      '@type': type,
      ...data
    }

    let scriptTag = document.querySelector('script[type="application/ld+json"]') as HTMLScriptElement
    if (!scriptTag) {
      scriptTag = document.createElement('script')
      scriptTag.type = 'application/ld+json'
      document.head.appendChild(scriptTag)
    }
    scriptTag.textContent = JSON.stringify(structuredData)
  }

  const setPageSEO = (seoData: SEOData) => {
    updateMetaTags(seoData)
  }

  const setCharacterSEO = (characterName: string, world: string, coreCount: number) => {
    const title = `${characterName} (${world}) - TibiaCores`
    const description = `View ${characterName}'s Tibia soulcore collection from ${world}. Track ${coreCount} unlocked soul cores and progress.`

    setPageSEO({
      title,
      description,
      keywords: `Tibia, ${characterName}, ${world}, soulcore, character, profile`,
      ogType: 'profile',
      canonical: `${window.location.origin}/characters/public/${characterName}`
    })

    // Add structured data for character
    generateStructuredData('Person', {
      name: characterName,
      description: `Tibia character from ${world} with ${coreCount} soul cores`,
      url: `${window.location.origin}/characters/public/${characterName}`,
      sameAs: [`https://www.tibia.com/community/?name=${characterName}`]
    })
  }

  const setListSEO = (listName: string, world: string, memberCount: number) => {
    const title = `${listName} - Tibia Soulcore List | TibiaCores`
    const description = `Join the ${listName} soulcore hunting list for ${world}. Collaborate with ${memberCount} members to track creature collection progress.`

    setPageSEO({
      title,
      description,
      keywords: `Tibia, soulcore, hunting, list, ${world}, collaboration, ${listName}`,
      ogType: 'article',
      noindex: true // Private lists shouldn't be indexed
    })
  }

  const setBlogPostSEO = (post: {
    id: string
    title: string
    excerpt: string
    date: string
    author: string
    image?: string
    tags: string[]
  }) => {
    const title = `${post.title} | TibiaCores Blog`
    const description = post.excerpt

    setPageSEO({
      title,
      description,
      keywords: `TibiaCores, blog, ${post.tags.join(', ')}, Tibia`,
      ogType: 'article',
      ogImage: post.image || '/logo.png',
      canonical: `${window.location.origin}/blog/${post.id}`
    })

    // Add structured data for blog post
    generateStructuredData('BlogPosting', {
      headline: post.title,
      description: post.excerpt,
      author: {
        '@type': 'Person',
        name: post.author
      },
      datePublished: post.date,
      publisher: {
        '@type': 'Organization',
        name: 'TibiaCores',
        logo: {
          '@type': 'ImageObject',
          url: `${window.location.origin}/logo.png`
        }
      },
      mainEntityOfPage: {
        '@type': 'WebPage',
        '@id': `${window.location.origin}/blog/${post.id}`
      },
      image: post.image || `${window.location.origin}/logo.png`
    })
  }

  const setHighscoresSEO = () => {
    const title = `Tibia Soulcore Highscores | TibiaCores`
    const description = `Discover top Tibia players by soulcore collection. View leaderboards and compare your progress with other players.`

    setPageSEO({
      title,
      description,
      keywords: 'Tibia, soulcore, highscores, leaderboard, rankings, top players',
      canonical: `${window.location.origin}/highscores`
    })

    // Add structured data for highscores
    generateStructuredData('WebPage', {
      name: 'Tibia Soulcore Highscores',
      description,
      url: `${window.location.origin}/highscores`,
      breadcrumb: {
        '@type': 'BreadcrumbList',
        itemListElement: [
          {
            '@type': 'ListItem',
            position: 1,
            name: 'Home',
            item: window.location.origin
          },
          {
            '@type': 'ListItem',
            position: 2,
            name: 'Highscores',
            item: `${window.location.origin}/highscores`
          }
        ]
      }
    })
  }

  return {
    setPageSEO,
    setCharacterSEO,
    setListSEO,
    setBlogPostSEO,
    setHighscoresSEO,
    generateStructuredData
  }
}
