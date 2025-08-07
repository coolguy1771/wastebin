import { ValidationRule } from '../hooks/useForm';

// Constants for validation
export const PASTE_CONSTRAINTS = {
  MIN_CONTENT_LENGTH: 1,
  MAX_CONTENT_LENGTH: 10 * 1024 * 1024, // 10MB in characters (roughly)
  MIN_EXPIRY_MINUTES: 1,
  MAX_EXPIRY_MINUTES: 365 * 24 * 60, // 1 year in minutes
  SUPPORTED_LANGUAGES: [
    'plaintext',
    'javascript',
    'typescript',
    'python',
    'java',
    'csharp',
    'cpp',
    'c',
    'go',
    'rust',
    'php',
    'ruby',
    'html',
    'css',
    'json',
    'xml',
    'yaml',
    'markdown',
    'sql',
    'bash',
    'powershell',
  ],
} as const;

// Custom validation functions
export const validatePasteContent = (content: unknown): string | null => {
  if (typeof content !== 'string') {
    return 'Invalid content type';
  }
  if (!content || content.trim().length === 0) {
    return 'Paste content cannot be empty';
  }

  if (content.length < PASTE_CONSTRAINTS.MIN_CONTENT_LENGTH) {
    return `Content must be at least ${PASTE_CONSTRAINTS.MIN_CONTENT_LENGTH} character`;
  }

  if (content.length > PASTE_CONSTRAINTS.MAX_CONTENT_LENGTH) {
    return `Content must be no more than ${Math.floor(PASTE_CONSTRAINTS.MAX_CONTENT_LENGTH / 1024 / 1024)}MB`;
  }

  return null;
};

export const validateLanguage = (language: unknown): string | null => {
  if (typeof language !== 'string') {
    return 'Invalid language type';
  }
  if (!language) {
    return 'Please select a language';
  }

  if (
    !PASTE_CONSTRAINTS.SUPPORTED_LANGUAGES.includes(
      language as (typeof PASTE_CONSTRAINTS.SUPPORTED_LANGUAGES)[number]
    )
  ) {
    return 'Please select a supported language';
  }

  return null;
};
export const validateExpiry = (expiryMinutes: unknown): string | null => {
  if (typeof expiryMinutes !== 'string') {
    return 'Invalid expiry time type';
  }
  if (!expiryMinutes) {
    return 'Please select an expiry time';
  }

  const minutes = parseInt(expiryMinutes, 10);

  if (isNaN(minutes)) {
    return 'Invalid expiry time';
  }

  if (minutes === 0) {
    return null; // Never expire is valid
  }

  if (minutes < PASTE_CONSTRAINTS.MIN_EXPIRY_MINUTES) {
    return `Expiry time must be at least ${PASTE_CONSTRAINTS.MIN_EXPIRY_MINUTES} minute`;
  }

  if (minutes > PASTE_CONSTRAINTS.MAX_EXPIRY_MINUTES) {
    return `Expiry time must be no more than ${Math.floor(PASTE_CONSTRAINTS.MAX_EXPIRY_MINUTES / 60 / 24)} days`;
  }

  return null;
};

// Validation rules for paste creation form
export interface PasteFormData {
  content: string;
  language: string;
  expires: string;
  burn: boolean;
  [key: string]: unknown;
}

export const pasteValidationRules: Record<keyof PasteFormData, ValidationRule> = {
  content: {
    required: true,
    custom: validatePasteContent as (value: unknown) => string | null,
  },
  language: {
    required: true,
    custom: validateLanguage as (value: unknown) => string | null,
  },
  expires: {
    required: true,
    custom: validateExpiry as (value: unknown) => string | null,
  },
  burn: {
    // No validation needed for boolean
  },
};

// Utility functions for form data transformation
export const formatExpiryTime = (expiresInMinutes: string): string | null => {
  if (expiresInMinutes === '0') {
    return null; // No expiration
  }
  const expiryDate = new Date(Date.now() + parseInt(expiresInMinutes) * 60000);
  return expiryDate.toISOString(); // RFC 3339 format
};

export const getLanguageDisplayName = (language: string): string => {
  const displayNames: Record<string, string> = {
    plaintext: 'Plain Text',
    javascript: 'JavaScript',
    typescript: 'TypeScript',
    python: 'Python',
    java: 'Java',
    csharp: 'C#',
    cpp: 'C++',
    c: 'C',
    go: 'Go',
    rust: 'Rust',
    php: 'PHP',
    ruby: 'Ruby',
    html: 'HTML',
    css: 'CSS',
    json: 'JSON',
    xml: 'XML',
    yaml: 'YAML',
    markdown: 'Markdown',
    sql: 'SQL',
    bash: 'Bash',
    powershell: 'PowerShell',
  };

  return displayNames[language] || language.charAt(0).toUpperCase() + language.slice(1);
};

export const getExpiryOptions = () => [
  { value: '0', label: 'Never' },
  { value: '10', label: '10 minutes' },
  { value: '60', label: '1 hour' },
  { value: '720', label: '12 hours' },
  { value: '1440', label: '1 day' },
  { value: '10080', label: '1 week' },
  { value: '43200', label: '1 month' },
  { value: '525600', label: '1 year' },
];

// Security validation
export const containsPotentiallyDangerousContent = (content: string): boolean => {
  // Check for potentially dangerous patterns
  const dangerousPatterns = [
    /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi,
    /javascript:/gi,
    /data:text\/html/gi,
    /vbscript:/gi,
    /on\w+\s*=/gi, // Event handlers like onclick, onload, etc.
  ];

  return dangerousPatterns.some(pattern => pattern.test(content));
};

export const sanitizeInput = (input: string): string => {
  // Basic HTML entity encoding
  return input
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;');
};
