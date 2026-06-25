/** @type {import('next').NextConfig} */
const nextConfig = {
  // Build minimal autonome : le démon Go lance `next start` (ou `node server.js`).
  output: "standalone",
  reactStrictMode: true,
};

export default nextConfig;
