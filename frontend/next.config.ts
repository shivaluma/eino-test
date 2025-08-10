import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

const nextConfig: NextConfig = {
  output: "standalone",
  // Disable static exports to prevent client reference manifest errors
  trailingSlash: false,
  devIndicators: {
    position: "bottom-right",
  }
};

const withNextIntl = createNextIntlPlugin();
export default withNextIntl(nextConfig);
