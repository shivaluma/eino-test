'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader } from 'lucide-react';

import { validateJWT, storeTokensSecurely } from '@/lib/auth/token-utils';

export default function OAuthCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(true);

  useEffect(() => {
    const handleCallback = async () => {
      try {
        // Check for OAuth provider errors first
        const errorParam = searchParams.get('error');
        if (errorParam) {
          const errorDescription = searchParams.get('error_description') || 'OAuth authentication failed';
          console.error('OAuth provider error:', errorParam, errorDescription);
          setError(errorDescription);
          setIsProcessing(false);
          
          // Redirect to sign-in with error context after delay
          setTimeout(() => {
            router.push(`/sign-in?error=${encodeURIComponent(errorDescription)}`);
          }, 3000);
          return;
        }

        // Validate required tokens
        const accessToken = searchParams.get('access_token');
        const refreshToken = searchParams.get('refresh_token');

        if (!accessToken) {
          throw new Error('Missing access token');
        }

        if (!refreshToken) {
          throw new Error('Missing refresh token');  
        }

        // Validate token format using secure utilities
        const accessTokenValidation = validateJWT(accessToken);
        const refreshTokenValidation = validateJWT(refreshToken);
        
        if (!accessTokenValidation.isValid) {
          throw new Error(`Invalid access token: ${accessTokenValidation.error}`);
        }
        
        if (!refreshTokenValidation.isValid) {
          throw new Error(`Invalid refresh token: ${refreshTokenValidation.error}`);
        }

        // Store tokens securely
        storeTokensSecurely(accessToken, refreshToken);
        
        // Redirect to home page on success
        router.replace('/');
        
      } catch (err) {
        console.error('OAuth callback error:', err);
        const errorMessage = err instanceof Error ? err.message : 'Authentication failed';
        setError(errorMessage);
        setIsProcessing(false);
        
        // Redirect to sign-in after delay
        setTimeout(() => {
          router.push(`/sign-in?error=${encodeURIComponent(errorMessage)}`);
        }, 3000);
      }
    };

    handleCallback();
  }, [searchParams, router]);

  if (error) {
    return (
      <div className="w-full h-screen flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="text-destructive flex items-center gap-2">
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
              Authentication Failed
            </CardTitle>
            <CardDescription className="text-sm">
              {error}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-4">
              You will be redirected to the sign-in page in a few seconds.
            </p>
            <button 
              onClick={() => router.push('/sign-in')}
              className="w-full px-4 py-2 text-sm font-medium text-white bg-primary rounded-md hover:bg-primary/90 transition-colors"
            >
              Return to Sign In
            </button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="w-full h-screen flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <svg className="w-5 h-5 text-green-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
            Authentication Successful
          </CardTitle>
          <CardDescription>
            {isProcessing 
              ? "Please wait while we complete your authentication..."
              : "Redirecting you to the application..."}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex justify-center">
          <Loader className="size-8 animate-spin text-primary" />
        </CardContent>
      </Card>
    </div>
  );
}