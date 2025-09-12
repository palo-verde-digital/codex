import { defineConfig } from "vite";

export default defineConfig({
    build: {
        outDir: 'dist',
        minify: false,
        rollupOptions: {
            input: './main.ts',
            output: {
                entryFileNames: 'main.js',
                format: 'es'
            }
        },
        lib: false
    },
    logLevel: 'info'
})