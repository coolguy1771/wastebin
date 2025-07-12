/**
 * MSW (Mock Service Worker) handlers for testing API calls
 */

import { http, HttpResponse } from 'msw';
import { PasteResponse, PasteDetails } from '@services/api';

// Mock data
const mockPaste: PasteDetails = {
  uuid: '123e4567-e89b-12d3-a456-426614174000',
  content: 'console.log("Hello, World!");',
  language: 'javascript',
  burn: false,
  expiry_timestamp: new Date(Date.now() + 86400000).toISOString(), // 24 hours from now
  created_at: new Date().toISOString(),
};

const mockPasteResponse: PasteResponse = {
  uuid: mockPaste.uuid,
  message: 'Paste created successfully',
};

// API handlers
export const handlers = [
  // Create paste
  http.post('/api/v1/paste', async ({ request }) => {
    const body = await request.json();

    // Simulate validation errors
    if (!body || typeof body !== 'object') {
      return HttpResponse.json({ error: 'Invalid request body' }, { status: 400 });
    }

    if (!body.content) {
      return HttpResponse.json({ error: 'Content is required' }, { status: 400 });
    }

    if (body.content.length > 10 * 1024 * 1024) {
      return HttpResponse.json({ error: 'Content too large' }, { status: 413 });
    }

    // Simulate rate limiting
    if (body.content.includes('RATE_LIMIT_TEST')) {
      return HttpResponse.json({ error: 'Rate limit exceeded' }, { status: 429 });
    }

    // Simulate server error
    if (body.content.includes('SERVER_ERROR_TEST')) {
      return HttpResponse.json({ error: 'Internal server error' }, { status: 500 });
    }

    return HttpResponse.json(mockPasteResponse, { status: 201 });
  }),

  // Get paste
  http.get('/api/v1/paste/:uuid', ({ params }) => {
    const { uuid } = params;

    // Simulate not found
    if (uuid === 'not-found') {
      return HttpResponse.json({ error: 'Paste not found' }, { status: 404 });
    }

    // Simulate expired paste
    if (uuid === 'expired') {
      return HttpResponse.json({ error: 'Paste has expired' }, { status: 410 });
    }

    // Simulate burn after read
    if (uuid === 'burned') {
      return HttpResponse.json({ error: 'Paste has been consumed' }, { status: 410 });
    }

    // Simulate server error
    if (uuid === 'server-error') {
      return HttpResponse.json({ error: 'Internal server error' }, { status: 500 });
    }

    return HttpResponse.json(mockPaste);
  }),

  // Get raw paste
  http.get('/paste/:uuid/raw', ({ params }) => {
    const { uuid } = params;

    if (uuid === 'not-found') {
      return new HttpResponse('Paste not found', { status: 404 });
    }

    if (uuid === 'expired') {
      return new HttpResponse('Paste has expired', { status: 410 });
    }

    return new HttpResponse(mockPaste.content, {
      headers: {
        'Content-Type': 'text/plain',
      },
    });
  }),

  // Delete paste
  http.delete('/api/v1/paste/:uuid', ({ params }) => {
    const { uuid } = params;

    if (uuid === 'not-found') {
      return HttpResponse.json({ error: 'Paste not found' }, { status: 404 });
    }

    if (uuid === 'unauthorized') {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 403 });
    }

    return HttpResponse.json({ message: 'Paste deleted successfully' });
  }),

  // Health check
  http.get('/healthz', () => {
    return HttpResponse.json({ status: 'ok' });
  }),

  // Simulate network errors
  http.get('/api/v1/network-error', () => {
    return HttpResponse.error();
  }),
];

// Export mock data for use in tests
export { mockPaste, mockPasteResponse };
