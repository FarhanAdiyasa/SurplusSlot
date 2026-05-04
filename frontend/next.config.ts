import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/proxy/:path*',
        destination: 'http://43.128.71.31:8090/:path*'
      }
    ]
  }
};

export default nextConfig;
