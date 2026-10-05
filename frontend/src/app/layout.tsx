import type { Metadata, Viewport } from "next";
import { Bebas_Neue, Geist, Geist_Mono } from "next/font/google";

import { gameBrand } from "@/lib/brand/game-brand";

import { Providers } from "./providers";

import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const bebasNeue = Bebas_Neue({
  variable: "--font-bebas-neue",
  weight: "400",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  description: gameBrand.tagline,
  title: {
    default: gameBrand.title,
    template: `%s · ${gameBrand.title}`,
  },
};

export const viewport: Viewport = {
  themeColor: "#0a0807",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en" className={`${geistSans.variable} ${geistMono.variable} ${bebasNeue.variable} h-full antialiased`}>
      <body className="flex min-h-full flex-col">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
