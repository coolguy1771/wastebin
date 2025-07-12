/**
 * Centralized API service for Wastebin
 * Handles all API communication with proper error handling, retries, and typing
 */

// Types
export interface PasteData {
  content: string;
  language: string;
  expiry_time: string | null;
  burn: boolean;
}

export interface PasteResponse {
  uuid: string;
  message: string;
}

export interface PasteDetails {
  uuid: string;
  content: string;
  language: string;
  burn: boolean;
  expiry_timestamp: string;
  created_at: string;
}

export interface ApiError {
  error: string;
  code?: string;
  details?: string;
}

import { config } from '../config/env';

// Configuration
const API_BASE_URL = config.apiBaseUrl;
const API_TIMEOUT = config.apiTimeout;
const MAX_RETRIES = 3;

// Custom error classes
export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string,
    public details?: string
  ) {
    super(message);
    this.name = 'APIError';
  }
}

export class NetworkError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'NetworkError';
  }
}

// Utility functions
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

const fetchWithTimeout = async (url: string, options: RequestInit = {}): Promise<Response> => {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), API_TIMEOUT);

  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal,
    });
    clearTimeout(timeoutId);
    return response;
  } catch (error) {
    clearTimeout(timeoutId);
    if (error instanceof Error && error.name === 'AbortError') {
      throw new NetworkError('Request timeout');
    }
    throw error;
  }
};

const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    let errorData: ApiError;
    
    try {
      errorData = await response.json();
    } catch {
      // If JSON parsing fails, create a generic error
      errorData = {
        error: response.statusText || 'Unknown error occurred',
      };
    }

    throw new APIError(
      errorData.error,
      response.status,
      errorData.code,
      errorData.details
    );
  }

  // Handle empty responses (like for raw content)
  const contentType = response.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return response.json();
  } else {
    return response.text() as unknown as T;
  }
};

const apiRequest = async <T>(
  endpoint: string,
  options: RequestInit = {},
  retries = 0
): Promise<T> => {
  try {
    const url = `${API_BASE_URL}${endpoint}`;
    const response = await fetchWithTimeout(url, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    });

    return handleResponse<T>(response);
  } catch (error) {
    // Retry logic for network errors
    if (error instanceof NetworkError && retries < MAX_RETRIES) {
      await delay(Math.pow(2, retries) * 1000); // Exponential backoff
      return apiRequest<T>(endpoint, options, retries + 1);
    }
    throw error;
  }
};

// API methods
export const pasteAPI = {
  /**
   * Create a new paste
   */
  create: async (data: PasteData): Promise<PasteResponse> => {
    return apiRequest<PasteResponse>('/api/v1/paste', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  /**
   * Get paste details by UUID
   */
  get: async (uuid: string): Promise<PasteDetails> => {
    return apiRequest<PasteDetails>(`/api/v1/paste/${uuid}`);
  },

  /**
   * Get raw paste content by UUID
   */
  getRaw: async (uuid: string): Promise<string> => {
    return apiRequest<string>(`/paste/${uuid}/raw`, {
      headers: {
        'Accept': 'text/plain',
      },
    });
  },

  /**
   * Delete paste by UUID
   */
  delete: async (uuid: string): Promise<{ message: string }> => {
    return apiRequest<{ message: string }>(`/api/v1/paste/${uuid}`, {
      method: 'DELETE',
    });
  },
};

// Health check
export const healthAPI = {
  check: async (): Promise<{ status: string }> => {
    return apiRequest<{ status: string }>('/healthz');
  },
};

// Error helper functions
export const getErrorMessage = (error: unknown): string => {
  if (error instanceof APIError) {
    return error.message;
  }
  if (error instanceof NetworkError) {
    return 'Network connection failed. Please check your internet connection and try again.';
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'An unexpected error occurred. Please try again.';
};

export const isRetryableError = (error: unknown): boolean => {
  if (error instanceof NetworkError) {
    return true;
  }
  if (error instanceof APIError) {
    // Retry on server errors but not client errors
    return error.status >= 500;
  }
  return false;
};