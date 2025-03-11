import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',
  // Disable ESLint during production builds
  eslint: {
    // Only run ESLint in development, not in production
    ignoreDuringBuilds: process.env.NODE_ENV === 'production',
  },
  // Disable TypeScript type checking during production builds
  typescript: {
    // Only check types in development, not in production
    ignoreBuildErrors: process.env.NODE_ENV === 'production',
  },
  async rewrites() {
    // Only apply rewrites in development or other non-production environments
    if (process.env.NODE_ENV !== 'production') {
      return [
        {
          source: '/api/:path*',
          destination: 'http://backend:8080/api/:path*',
        },
      ];
    }
    
    // Return empty array in production
    return [];
  },
};

export default nextConfig;
