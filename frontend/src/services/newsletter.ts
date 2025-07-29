import axios from 'axios'

interface NewsletterSubscribeRequest {
  email: string
}

interface NewsletterSubscribeResponse {
  message: string
}

class NewsletterService {
  async subscribe(email: string): Promise<string> {
    try {
      const response = await axios.post<NewsletterSubscribeResponse>('/newsletter/subscribe', {
        email
      } as NewsletterSubscribeRequest)
      
      return response.data.message
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.data?.message) {
        throw new Error(error.response.data.message)
      }
      throw new Error('Failed to subscribe to newsletter')
    }
  }
}

export const newsletterService = new NewsletterService()
