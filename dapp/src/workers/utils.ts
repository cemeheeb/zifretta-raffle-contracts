type RetryOptions = {
  retries?: number;
  delayMs?: number;
  shouldRetry?: (error: any) => boolean;
};

async function withRetry<T>(
  delegate: () => Promise<T>,
  options?: RetryOptions
): Promise<T> {
  const { retries = 3, delayMs = 1000, shouldRetry = () => true } = options || {};
  let attempts = 0;

  while (attempts <= retries) {
    try {
      return await delegate();
    } catch (error) {
      if (attempts === retries || !shouldRetry(error)) {
        throw error;
      }
      attempts++;
      console.warn(`tonapi call failed. retrying in ${delayMs}ms (attempt ${attempts}/${retries})...`);
      await new Promise((resolve) => setTimeout(resolve, delayMs));
    }
  }

  throw new Error("Maximum retries exceeded.");
}

export async function infinityRetry<T>(delegate: () => Promise<T>) {
  return await withRetry<T>(delegate, {
    retries: 999,
    shouldRetry: ({error}) => error === "rate limit: free tier",
  });
}
