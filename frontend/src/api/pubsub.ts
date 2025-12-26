import http from './http'
import type { PublishRequest, PublishResponse } from '@/types/pubsub'

export const pubsubApi = {
  /**
   * Publish a message to a Redis channel
   * Requirements: 2.4
   */
  publish(request: PublishRequest) {
    return http.post<PublishResponse>('/pubsub/publish', request)
  }
}

export default pubsubApi
