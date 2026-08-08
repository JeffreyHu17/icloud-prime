import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vite";

export default defineConfig({
  base: "./",
  plugins: [vue()],
  build: {
    outDir: "../internal/server/static/dist",
    emptyOutDir: true,
  },
});
