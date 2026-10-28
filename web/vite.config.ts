import { defineConfig } from 'vitest/config';

// The app is served under /app/ on beetreecraft.com and built straight into
// the Hugo site's static directory so one sync publishes everything.
export default defineConfig({
  base: '/app/',
  build: {
    outDir: '../website/static/app',
    emptyOutDir: true,
    chunkSizeWarningLimit: 900,
  },
  test: {
    include: ['src/**/*.test.ts'],
  },
});
