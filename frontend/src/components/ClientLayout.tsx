'use client';

import { ReactNode } from 'react';
import { PlayerProvider } from "@/contexts/PlayerContext";

interface ClientLayoutProps {
  children: ReactNode;
}

export default function ClientLayout({ children }: ClientLayoutProps) {
  return (
    <PlayerProvider>
      {children}
    </PlayerProvider>
  );
} 