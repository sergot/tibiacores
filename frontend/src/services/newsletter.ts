import axios from 'axios'

export interface NewsletterSubscriptionResponse {
  message: string
  status: 'subscribed' | 'already_subscribed' | 'resubscribed' | 'unsubscribed'
}

export interface NewsletterStats {
  total_subscribers: number
  active_subscribers: number
  pending_confirmation: number
  unsubscribed: number
}

class NewsletterService {
  async subscribe(email: string): Promise<NewsletterSubscriptionResponse> {
    const response = await axios.post<NewsletterSubscriptionResponse>('/newsletter/subscribe', {
      email,
    })
    return response.data
  }

  async unsubscribe(email: string): Promise<NewsletterSubscriptionResponse> {
    const response = await axios.post<NewsletterSubscriptionResponse>('/newsletter/unsubscribe', {
      email,
    })
    return response.data
  }

  async getStats(): Promise<NewsletterStats> {
    const response = await axios.get<NewsletterStats>('/newsletter/stats')
    return response.data
  }

  async checkSubscriptionStatus(email: string): Promise<{ subscribed: boolean }> {
    const response = await axios.get<{ subscribed: boolean }>('/newsletter/status', {
      params: { email }
    })
    return response.data
  }
}

export const newsletterService = new NewsletterService()
