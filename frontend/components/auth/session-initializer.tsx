'use client';

import { useEffect } from 'react';
import { useAuth } from '@/lib/auth/context';
import type { SessionUser } from '@/lib/auth/server';

interface SessionInitializerProps {
  initialSession?: { session: { user: SessionUser }; user: SessionUser } | null;
}

export function SessionInitializer({ initialSession }: SessionInitializerProps) {
  const { login } = useAuth();

  useEffect(() => {
    // If we have initial session data from server, set it in the auth context
    if (initialSession?.user) {
      login(initialSession.user);
    }
  }, [initialSession, login]);

  // This component doesn't render anything
  return null;
}