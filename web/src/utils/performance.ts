/**
 * Performance monitoring and optimization utilities
 */

// Performance timing interface
interface PerformanceTiming {
  name: string;
  startTime: number;
  endTime?: number;
  duration?: number;
}

// Performance monitor class
class PerformanceMonitor {
  private timings: Map<string, PerformanceTiming> = new Map();
  private isEnabled: boolean;

  constructor(enabled: boolean = true) {
    this.isEnabled = enabled;
  }

  // Start timing
  start(name: string): void {
    if (!this.isEnabled) return;

    this.timings.set(name, {
      name,
      startTime: performance.now(),
    });
  }

  // End timing
  end(name: string): number | null {
    if (!this.isEnabled) return null;

    const timing = this.timings.get(name);
    if (!timing) {
      console.warn(`Performance timing '${name}' not found`);
      return null;
    }

    const endTime = performance.now();
    const duration = endTime - timing.startTime;

    timing.endTime = endTime;
    timing.duration = duration;

    if (process.env.NODE_ENV === 'development') {
      console.log(`⏱️ ${name}: ${duration.toFixed(2)}ms`);
    }

    return duration;
  }

  // Get timing result
  getTiming(name: string): PerformanceTiming | null {
    return this.timings.get(name) || null;
  }

  // Get all timings
  getAllTimings(): PerformanceTiming[] {
    return Array.from(this.timings.values());
  }

  // Clear all timings
  clear(): void {
    this.timings.clear();
  }

  // Measure function execution time
  measure<T>(name: string, fn: () => T): T {
    this.start(name);
    const result = fn();
    this.end(name);
    return result;
  }

  // Measure async function execution time
  async measureAsync<T>(name: string, fn: () => Promise<T>): Promise<T> {
    this.start(name);
    try {
      const result = await fn();
      this.end(name);
      return result;
    } catch (error) {
      this.end(name);
      throw error;
    }
  }
}

// Global performance monitor instance
export const performanceMonitor = new PerformanceMonitor(
  process.env.NODE_ENV === 'development'
);

// Web Vitals monitoring
export interface WebVitals {
  fcp?: number; // First Contentful Paint
  lcp?: number; // Largest Contentful Paint
  fid?: number; // First Input Delay
  cls?: number; // Cumulative Layout Shift
  ttfb?: number; // Time to First Byte
}

// Collect Web Vitals
export const collectWebVitals = (): Promise<WebVitals> => {
  return new Promise((resolve) => {
    const vitals: WebVitals = {};

    // First Contentful Paint
    new PerformanceObserver((list) => {
      const entries = list.getEntries();
      entries.forEach((entry) => {
        if (entry.name === 'first-contentful-paint') {
          vitals.fcp = entry.startTime;
        }
      });
    }).observe({ entryTypes: ['paint'] });

    // Largest Contentful Paint
    new PerformanceObserver((list) => {
      const entries = list.getEntries();
      const lastEntry = entries[entries.length - 1];
      vitals.lcp = lastEntry.startTime;
    }).observe({ entryTypes: ['largest-contentful-paint'] });

    // First Input Delay
    new PerformanceObserver((list) => {
      const entries = list.getEntries();
      entries.forEach((entry: any) => {
        vitals.fid = entry.processingStart - entry.startTime;
      });
    }).observe({ entryTypes: ['first-input'] });

    // Cumulative Layout Shift
    let clsValue = 0;
    new PerformanceObserver((list) => {
      const entries = list.getEntries();
      entries.forEach((entry: any) => {
        if (!entry.hadRecentInput) {
          clsValue += entry.value;
        }
      });
      vitals.cls = clsValue;
    }).observe({ entryTypes: ['layout-shift'] });

    // Time to First Byte
    const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
    if (navigation) {
      vitals.ttfb = navigation.responseStart - navigation.fetchStart;
    }

    // Resolve after a short delay to collect metrics
    setTimeout(() => resolve(vitals), 1000);
  });
};

// Memory usage monitoring
export const getMemoryUsage = (): any => {
  if ('memory' in performance) {
    return {
      used: Math.round((performance as any).memory.usedJSHeapSize / 1024 / 1024),
      total: Math.round((performance as any).memory.totalJSHeapSize / 1024 / 1024),
      limit: Math.round((performance as any).memory.jsHeapSizeLimit / 1024 / 1024),
    };
  }
  return null;
};

// Bundle size analysis (development only)
export const logBundleInfo = (): void => {
  if (process.env.NODE_ENV !== 'development') return;

  // Log module loading performance
  const modules = performance.getEntriesByType('module');
  if (modules.length > 0) {
    console.group('📦 Module Loading Performance');
    modules.forEach((module) => {
      console.log(`${module.name}: ${module.duration?.toFixed(2)}ms`);
    });
    console.groupEnd();
  }

  // Log resource loading
  const resources = performance.getEntriesByType('resource');
  const jsResources = resources.filter((r) => r.name.endsWith('.js'));
  const cssResources = resources.filter((r) => r.name.endsWith('.css'));

  console.group('🎯 Resource Loading Performance');
  console.log(`JavaScript files: ${jsResources.length}`);
  console.log(`CSS files: ${cssResources.length}`);
  
  const totalSize = resources.reduce((sum, resource: any) => {
    return sum + (resource.transferSize || 0);
  }, 0);
  
  console.log(`Total transfer size: ${(totalSize / 1024).toFixed(2)} KB`);
  console.groupEnd();
};

// Component render time tracking
export const useRenderTime = (componentName: string) => {
  React.useEffect(() => {
    const startTime = performance.now();
    
    return () => {
      const endTime = performance.now();
      const duration = endTime - startTime;
      
      if (process.env.NODE_ENV === 'development' && duration > 16) {
        console.warn(`🐌 Slow render: ${componentName} took ${duration.toFixed(2)}ms`);
      }
    };
  });
};

// Lazy loading utilities
export const createLazyComponent = <T extends React.ComponentType<any>>(
  importFn: () => Promise<{ default: T }>,
  fallback?: React.ComponentType
) => {
  const LazyComponent = React.lazy(importFn);
  
  return (props: React.ComponentProps<T>) =>
    React.createElement(
      React.Suspense,
      { fallback: fallback ? React.createElement(fallback) : React.createElement('div', {}, 'Loading...') },
      React.createElement(LazyComponent, props)
    );
};

// Image optimization utilities
export const optimizeImage = (src: string, _options: {
  width?: number;
  height?: number;
  quality?: number;
  format?: 'webp' | 'avif' | 'jpeg' | 'png';
} = {}): string => {
  // This would typically integrate with an image optimization service
  // For now, return the original source
  return src;
};

// Preload critical resources
export const preloadResource = (href: string, as: string): void => {
  const link = document.createElement('link');
  link.rel = 'preload';
  link.href = href;
  link.as = as;
  document.head.appendChild(link);
};

// Critical CSS injection
export const injectCriticalCSS = (css: string): void => {
  const style = document.createElement('style');
  style.textContent = css;
  document.head.appendChild(style);
};

// Service worker registration
export const registerServiceWorker = async (): Promise<boolean> => {
  if ('serviceWorker' in navigator && process.env.NODE_ENV === 'production') {
    try {
      const registration = await navigator.serviceWorker.register('/sw.js');
      console.log('✅ Service Worker registered:', registration);
      return true;
    } catch (error) {
      console.error('❌ Service Worker registration failed:', error);
      return false;
    }
  }
  return false;
};

// Export React import for useRenderTime hook
import React from 'react';