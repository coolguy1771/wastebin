import { useState, useCallback } from 'react';
import { getErrorMessage, isRetryableError } from '../services/api';

export interface UseApiState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  isRetryable: boolean;
}

export interface UseApiActions<T> {
  retry: () => Promise<void>;
  reset: () => void;
  execute: (operation: () => Promise<T>) => Promise<T | null>;
}

export type UseApiReturn<T> = UseApiState<T> & UseApiActions<T>;

/**
 * Custom hook for handling API operations with loading, error states, and retry functionality
 */
export function useApi<T = unknown>(): UseApiReturn<T> {
  const [state, setState] = useState<UseApiState<T>>({
    data: null,
    loading: false,
    error: null,
    isRetryable: false,
  });

  const [lastOperation, setLastOperation] = useState<(() => Promise<T>) | null>(null);

  const execute = useCallback(async (operation: () => Promise<T>): Promise<T | null> => {
    setState(prev => ({
      ...prev,
      loading: true,
      error: null,
    }));

    setLastOperation(() => operation);

    try {
      const result = await operation();
      setState(prev => ({
        ...prev,
        data: result,
        loading: false,
        error: null,
        isRetryable: false,
      }));
      return result;
    } catch (error) {
      const errorMessage = getErrorMessage(error);
      const retryable = isRetryableError(error);

      setState(prev => ({
        ...prev,
        loading: false,
        error: errorMessage,
        isRetryable: retryable,
      }));

      return null;
    }
  }, []);

  const retry = useCallback(async (): Promise<void> => {
    if (lastOperation) {
      await execute(lastOperation);
    }
  }, [execute, lastOperation]);

  const reset = useCallback(() => {
    setState({
      data: null,
      loading: false,
      error: null,
      isRetryable: false,
    });
    setLastOperation(null);
  }, []);

  return {
    ...state,
    retry,
    reset,
    execute,
  };
}

/**
 * Specialized hook for API operations that return void (like delete operations)
 */
export function useApiMutation(): UseApiReturn<void> & {
  execute: (operation: () => Promise<void>) => Promise<void>;
} {
  const api = useApi<void>();

  const execute = useCallback(
    async (operation: () => Promise<void>): Promise<void> => {
      await api.execute(operation);
    },
    [api.execute]
  );

  return {
    ...api,
    execute,
  };
}
