import { useState, useCallback, useMemo } from 'react';
import { SelectChangeEvent } from '@mui/material';

export interface ValidationRule<T = unknown> {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: RegExp;
  custom?: (value: T) => string | null;
  message?: string;
}

export interface FormField<T = unknown> {
  value: T;
  error: string | null;
  touched: boolean;
  rules?: ValidationRule<T>;
}

export interface UseFormOptions<T> {
  initialValues: T;
  validationRules?: Partial<Record<keyof T, ValidationRule>>;
  onSubmit?: (values: T) => Promise<void> | void;
}

export interface UseFormReturn<T> {
  values: T;
  errors: Partial<Record<keyof T, string>>;
  touched: Partial<Record<keyof T, boolean>>;
  isValid: boolean;
  isSubmitting: boolean;
  setValue: (field: keyof T, value: T[keyof T]) => void;
  setFieldError: (field: keyof T, error: string | null) => void;
  setFieldTouched: (field: keyof T, touched?: boolean) => void;
  reset: () => void;
  handleSubmit: (e?: React.FormEvent) => Promise<void>;
  getFieldProps: (field: keyof T) => {
    value: T[keyof T];
    onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
    onBlur: () => void;
    error: boolean;
    helperText: string | undefined;
  };
  getSelectProps: (field: keyof T) => {
    value: T[keyof T];
    onChange: (e: SelectChangeEvent) => void;
    onBlur: () => void;
    error: boolean;
  };
}

/**
 * React hook for managing form state, validation, error handling, and Material-UI integration.
 *
 * Provides value management, field-level and form-level validation, error tracking, touched state, and helpers for integrating with Material-UI input and select components. Supports synchronous and asynchronous form submission and custom validation rules.
 *
 * @returns An object with form state, validation status, and methods for updating fields, handling submission, and integrating with Material-UI components.
 */
export function useForm<T extends Record<string, unknown>>({
  initialValues,
  validationRules = {},
  onSubmit,
}: UseFormOptions<T>): UseFormReturn<T> {
  const [values, setValues] = useState<T>(initialValues);
  const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({});
  const [touched, setTouched] = useState<Partial<Record<keyof T, boolean>>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Validation function
  const validateField = useCallback(
    (field: keyof T, value: T[keyof T]): string | null => {
      const rules = validationRules[field];
      if (!rules) return null;

      // Required validation
      if (rules.required && (!value || (typeof value === 'string' && value.trim() === ''))) {
        return rules.message || 'This field is required';
      }

      // Skip other validations if field is empty and not required
      if (!value || (typeof value === 'string' && value.trim() === '')) {
        return null;
      }

      // String validations
      if (typeof value === 'string') {
        if (rules.minLength && value.length < rules.minLength) {
          return rules.message || `Must be at least ${rules.minLength} characters`;
        }

        if (rules.maxLength && value.length > rules.maxLength) {
          return rules.message || `Must be no more than ${rules.maxLength} characters`;
        }

        if (rules.pattern && !rules.pattern.test(value)) {
          return rules.message || 'Invalid format';
        }
      }

      // Custom validation
      if (rules.custom) {
        const customError = rules.custom(value);
        if (customError) {
          return customError;
        }
      }

      return null;
    },
    [validationRules]
  );

  // Validate all fields
  const validateForm = useCallback((): boolean => {
    const newErrors: Partial<Record<keyof T, string>> = {};
    let hasErrors = false;

    Object.keys(values).forEach(key => {
      const field = key as keyof T;
      const error = validateField(field, values[field]);
      if (error) {
        newErrors[field] = error;
        hasErrors = true;
      }
    });

    setErrors(newErrors);
    return !hasErrors;
  }, [values, validateField]);

  // Set field value with validation
  const setValue = useCallback(
    (field: keyof T, value: T[keyof T]) => {
      setValues(prev => ({ ...prev, [field]: value }));

      // Validate field if it's been touched
      if (touched[field]) {
        const error = validateField(field, value);
        setErrors(prev => ({ ...prev, [field]: error }));
      }
    },
    [touched, validateField]
  );

  // Set field error manually
  const setFieldError = useCallback((field: keyof T, error: string | null) => {
    setErrors(prev => ({ ...prev, [field]: error }));
  }, []);

  // Set field touched state
  const setFieldTouched = useCallback(
    (field: keyof T, isTouched: boolean = true) => {
      setTouched(prev => ({ ...prev, [field]: isTouched }));

      // Validate field when touched
      if (isTouched) {
        const error = validateField(field, values[field]);
        setErrors(prev => ({ ...prev, [field]: error }));
      }
    },
    [values, validateField]
  );

  // Reset form
  const reset = useCallback(() => {
    setValues(initialValues);
    setErrors({});
    setTouched({});
    setIsSubmitting(false);
  }, [initialValues]);

  // Handle form submission
  const handleSubmit = useCallback(
    async (e?: React.FormEvent) => {
      if (e) {
        e.preventDefault();
      }

      // Mark all fields as touched
      const newTouched: Partial<Record<keyof T, boolean>> = {};
      Object.keys(values).forEach(key => {
        newTouched[key as keyof T] = true;
      });
      setTouched(newTouched);

      // Validate form
      const isValid = validateForm();
      if (!isValid || !onSubmit) {
        return;
      }

      setIsSubmitting(true);
      try {
        await onSubmit(values);
      } catch (error) {
        // Handle submission errors
        console.error('Form submission error:', error);
      } finally {
        setIsSubmitting(false);
      }
    },
    [values, validateForm, onSubmit]
  );

  // Get field props for easy integration with Material-UI
  const getFieldProps = useCallback(
    (field: keyof T) => ({
      value: values[field] || '',
      onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        setValue(field, e.target.value);
      },
      onBlur: () => {
        setFieldTouched(field, true);
      },
      error: !!(touched[field] && errors[field]),
      helperText: touched[field] ? errors[field] : undefined,
    }),
    [values, errors, touched, setValue, setFieldTouched]
  );

  // Get select props for Material-UI Select components
  const getSelectProps = useCallback(
    (field: keyof T) => ({
      value: values[field] || '',
      onChange: (e: SelectChangeEvent) => {
        setValue(field, e.target.value);
      },
      onBlur: () => {
        setFieldTouched(field, true);
      },
      error: !!(touched[field] && errors[field]),
    }),
    [values, errors, touched, setValue, setFieldTouched]
  );

  // Check if form is valid
  const isValid = useMemo(() => {
    return Object.keys(errors).length === 0;
  }, [errors]);

  return {
    values,
    errors,
    touched,
    isValid,
    isSubmitting,
    setValue,
    setFieldError,
    setFieldTouched,
    reset,
    handleSubmit,
    getFieldProps,
    getSelectProps,
  };
}
