import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import ClientLayout from "@/components/ClientLayout";
import Link from "next/link";

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
          {/* Header */}
          <header className="bg-white dark:bg-gray-800 shadow">
            <div className="container mx-auto px-4 py-4 flex justify-between items-center">
              <Link href="/" className="text-xl font-bold text-gray-900 dark:text-white">
                SoulPit Manager
              </Link>
              <nav>
                <ul className="flex space-x-4">
                  <li>
                    <Link href="/profile" className="text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white">
                      Profile
                    </Link>
                  </li>
                  {/* Theme toggle button will go here */}
                </ul>
              </nav>
            </div>
          </header>

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
