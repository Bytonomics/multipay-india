/**
 * Script loading utility with deduplication
 * Ensures each script URL is loaded only once per session
 */

/**
 * Registry of in-flight script loading promises keyed by script URL
 * Prevents duplicate requests for the same script
 *
 * ALLOWED-EXCEPTION: This is the ONLY Map permitted in the library.
 * A generic URL→promise memoization cache (unbounded keys) is correctly
 * implemented as a Map. This pattern must NOT be copied elsewhere; use
 * named-field interfaces for all other domain-keyed data.
 */
const scriptRegistry = new Map<string, Promise<void>>();

/**
 * Load an external JavaScript script dynamically
 * Deduplicates concurrent requests for the same script URL
 *
 * @param src - The URL of the script to load
 * @returns Promise that resolves when script is loaded, rejects on load failure
 *
 * @example
 * ```ts
 * await loadScript('https://checkout.cashfree.com/script.js');
 * await loadScript('https://checkout.razorpay.com/v1/checkout.js');
 * ```
 */
export async function loadScript(src: string): Promise<void> {
  // Return existing promise if script is already being loaded
  const existingPromise = scriptRegistry.get(src);
  if (existingPromise) {
    return existingPromise;
  }

  // Check if script is already loaded in DOM (by URL)
  if (isScriptLoaded(src)) {
    return Promise.resolve();
  }

  // Create and cache the loading promise
  const loadingPromise = createScriptElement(src);
  scriptRegistry.set(src, loadingPromise);

  try {
    await loadingPromise;
  } catch (error) {
    // Remove failed promise from registry so it can be retried
    scriptRegistry.delete(src);
    throw error;
  }
}

/**
 * Create a script element and append it to document head
 * @param src - Script URL to load
 * @returns Promise that resolves on load, rejects on error
 */
function createScriptElement(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = src;
    script.async = true;
    script.defer = true;

    script.onload = () => resolve();
    script.onerror = () => reject(new Error(`Failed to load script: ${src}`));

    document.head.appendChild(script);
  });
}

/**
 * Check if a script with the given src is already loaded in the DOM
 * @param src - Script URL to check
 * @returns true if script element with matching src exists
 */
function isScriptLoaded(src: string): boolean {
  const scripts = document.getElementsByTagName("script");
  for (let i = 0; i < scripts.length; i++) {
    const script = scripts[i];
    if (script && (script.src === src || script.getAttribute("src") === src)) {
      return true;
    }
  }
  return false;
}

/**
 * Clear a script from the registry (for testing or cleanup)
 * @param src - Script URL to remove from registry
 */
export function clearScriptRegistry(src?: string): void {
  if (src) {
    scriptRegistry.delete(src);
  } else {
    scriptRegistry.clear();
  }
}

/**
 * Get all script URLs currently in the loading registry
 * @returns Array of script URLs being loaded
 */
export function getRegisteredScripts(): string[] {
  return Array.from(scriptRegistry.keys());
}
