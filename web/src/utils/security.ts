/**
 * Security utilities for client-side protection
 */

// CSP (Content Security Policy) utilities
export const CSP_DIRECTIVES = {
  defaultSrc: ["'self'"],
  scriptSrc: ["'self'", "'unsafe-inline'", "'unsafe-eval'"], // Note: unsafe-* should be removed in production
  styleSrc: ["'self'", "'unsafe-inline'", 'https://fonts.googleapis.com'],
  fontSrc: ["'self'", 'https://fonts.gstatic.com'],
  imgSrc: ["'self'", 'data:', 'https:'],
  connectSrc: ["'self'"],
  objectSrc: ["'none'"],
  mediaSrc: ["'self'"],
  frameSrc: ["'none'"],
  upgradeInsecureRequests: true,
} as const;

// HTML sanitization
export const sanitizeHTML = (input: string): string => {
  // Create a temporary element to parse HTML
  const temp = document.createElement('div');
  temp.textContent = input;
  return temp.innerHTML;
};

// Advanced HTML sanitization with allowlist
export const sanitizeHTMLAdvanced = (
  input: string,
  allowedTags: string[] = [],
  allowedAttributes: string[] = []
): string => {
  const temp = document.createElement('div');
  temp.innerHTML = input;

  // Remove all scripts
  const scripts = temp.querySelectorAll('script');
  scripts.forEach(script => script.remove());

  // Remove dangerous attributes
  const dangerousAttributes = [
    'onclick',
    'onload',
    'onerror',
    'onmouseover',
    'onmouseout',
    'onfocus',
    'onblur',
    'onchange',
    'onsubmit',
    'onreset',
    'onselect',
    'onunload',
    'onresize',
    'onscroll',
    'href',
    'src',
    'action',
    'formaction',
    'background',
    'lowsrc',
    'ping',
    'poster',
    'xlink:href',
    'xml:base',
  ];

  const allElements = temp.querySelectorAll('*');
  allElements.forEach(element => {
    // Remove dangerous attributes
    dangerousAttributes.forEach(attr => {
      if (element.hasAttribute(attr) && !allowedAttributes.includes(attr)) {
        element.removeAttribute(attr);
      }
    });

    // Remove disallowed tags
    if (!allowedTags.includes(element.tagName.toLowerCase())) {
      // Replace with text content
      const textNode = document.createTextNode(element.textContent || '');
      element.parentNode?.replaceChild(textNode, element);
    }
  });

  return temp.innerHTML;
};

// URL validation
export const isValidURL = (url: string): boolean => {
  try {
    const parsedURL = new URL(url);
    return ['http:', 'https:'].includes(parsedURL.protocol);
  } catch {
    return false;
  }
};

// Safe URL creation (prevents javascript: and data: URLs)
export const createSafeURL = (url: string): string | null => {
  if (!isValidURL(url)) {
    return null;
  }

  const parsedURL = new URL(url);

  // Block dangerous protocols
  const dangerousProtocols = ['javascript:', 'data:', 'vbscript:', 'file:'];
  if (dangerousProtocols.includes(parsedURL.protocol)) {
    return null;
  }

  return parsedURL.toString();
};

// Input validation patterns
export const VALIDATION_PATTERNS = {
  uuid: /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i,
  email: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  alphanumeric: /^[a-zA-Z0-9]+$/,
  alphanumericWithSpaces: /^[a-zA-Z0-9\s]+$/,
  noScript: /^(?!.*<script).*$/i,
  noHTML: /^[^<>]*$/,
} as const;

// Validate UUID
export const isValidUUID = (uuid: string): boolean => {
  return VALIDATION_PATTERNS.uuid.test(uuid);
};

// Validate email
export const isValidEmail = (email: string): boolean => {
  return VALIDATION_PATTERNS.email.test(email);
};

// Check for potential XSS patterns
export const containsXSS = (input: string): boolean => {
  const xssPatterns = [
    /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi,
    /javascript:/gi,
    /vbscript:/gi,
    /on\w+\s*=/gi,
    /<iframe/gi,
    /<object/gi,
    /<embed/gi,
    /<link/gi,
    /<meta/gi,
    /<style/gi,
    /expression\s*\(/gi,
    /url\s*\(/gi,
    /@import/gi,
    /xss/gi,
  ];

  return xssPatterns.some(pattern => pattern.test(input));
};

// Safe string encoding for URLs
export const encodeURIComponentSafe = (str: string): string => {
  return encodeURIComponent(str).replace(/[!'()*]/g, c => {
    return '%' + c.charCodeAt(0).toString(16);
  });
};

// Content Security Policy header generation
export const generateCSPHeader = (): string => {
  const directives = Object.entries(CSP_DIRECTIVES)
    .map(([key, value]) => {
      if (key === 'upgradeInsecureRequests') {
        return value ? 'upgrade-insecure-requests' : '';
      }
      const kebabKey = key.replace(/([A-Z])/g, '-$1').toLowerCase();
      const directiveValue = Array.isArray(value) ? value.join(' ') : value;
      return `${kebabKey} ${directiveValue}`;
    })
    .filter(Boolean)
    .join('; ');

  return directives;
};

// Rate limiting client-side check (simple implementation)
class RateLimiter {
  private attempts: Map<string, number[]> = new Map();
  private readonly maxAttempts: number;
  private readonly windowMs: number;

  constructor(maxAttempts: number = 5, windowMs: number = 60000) {
    this.maxAttempts = maxAttempts;
    this.windowMs = windowMs;
  }

  isAllowed(key: string): boolean {
    const now = Date.now();
    const attempts = this.attempts.get(key) || [];

    // Remove old attempts outside the window
    const validAttempts = attempts.filter(time => now - time < this.windowMs);

    if (validAttempts.length >= this.maxAttempts) {
      return false;
    }

    // Add current attempt
    validAttempts.push(now);
    this.attempts.set(key, validAttempts);

    return true;
  }

  reset(key: string): void {
    this.attempts.delete(key);
  }
}

// Export rate limiter instance
export const rateLimiter = new RateLimiter();

// Secure random string generation
export const generateSecureRandomString = (length: number = 32): string => {
  const array = new Uint8Array(length);
  crypto.getRandomValues(array);
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
};

// Check if running in secure context (HTTPS)
export const isSecureContext = (): boolean => {
  return window.isSecureContext;
};

// Validate file upload security
export const validateFile = (file: File): { valid: boolean; error?: string } => {
  // Check file size (10MB max)
  const maxSize = 10 * 1024 * 1024;
  if (file.size > maxSize) {
    return { valid: false, error: 'File size exceeds 10MB limit' };
  }

  // Check file type allowlist
  const allowedTypes = [
    'text/plain',
    'text/markdown',
    'text/html',
    'text/css',
    'text/javascript',
    'application/json',
    'application/xml',
  ];

  if (!allowedTypes.includes(file.type)) {
    return { valid: false, error: 'File type not allowed' };
  }

  // Check for dangerous file extensions
  const dangerousExtensions = ['.exe', '.bat', '.cmd', '.scr', '.pif', '.vbs', '.js'];
  const fileExtension = file.name.toLowerCase().substring(file.name.lastIndexOf('.'));

  if (dangerousExtensions.includes(fileExtension)) {
    return { valid: false, error: 'File extension not allowed' };
  }

  return { valid: true };
};

// Escape HTML entities
export const escapeHTML = (str: string): string => {
  const htmlEscapes: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#x27;',
    '/': '&#x2F;',
  };

  return str.replace(/[&<>"'/]/g, match => htmlEscapes[match]);
};

// Unescape HTML entities
export const unescapeHTML = (str: string): string => {
  const htmlUnescapes: Record<string, string> = {
    '&amp;': '&',
    '&lt;': '<',
    '&gt;': '>',
    '&quot;': '"',
    '&#x27;': "'",
    '&#x2F;': '/',
  };

  return str.replace(/&(?:amp|lt|gt|quot|#x27|#x2F);/g, match => htmlUnescapes[match]);
};
