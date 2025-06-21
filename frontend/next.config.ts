import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  compiler: {
    removeConsole: {
      exclude: ["error", "warn", "info"],
    },
  },
};

export default nextConfig;
