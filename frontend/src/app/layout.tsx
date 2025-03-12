import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import ClientLayout from "@/components/ClientLayout";
import Navbar from "@/components/layout/Navbar";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "SoulPit Manager - Tibia Soul Pit Tracking Tool",
  description: "Track your progress in Tibia's Soul Pit dungeon",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${inter.className} min-h-screen bg-amber-50 dark:bg-gray-900`}>
        <ClientLayout>
          {/* Header with Navbar component */}
          <Navbar />

          {/* Main Content */}
          <main className="container mx-auto p-4 sm:p-6 lg:p-8">
            {children}
          </main>

          {/* Footer */}
          <footer className="border-t border-amber-200 dark:border-amber-800 p-4 text-center text-sm text-amber-700 dark:text-amber-400">
            <p>SoulPit Manager - Tibia Soul Pit Tracking Tool &copy; {new Date().getFullYear()}</p>
          </footer>
        </ClientLayout>
      </body>
    </html>
  );
}
