"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

import { useObjectState } from "@/hooks/use-object-state";

import { Loader } from "lucide-react";

import { useLoginMutation } from "@/lib/mutations/auth";

import { GithubIcon } from "@/components/icons/github-icon";
import { GoogleIcon } from "@/components/icons/google-icon";
import { useTranslations } from "next-intl";
import { SocialAuthenticationProvider } from "@/types/authentication";
import { initiateOAuth } from "@/lib/auth/oauth";
import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";

export default function SignIn({
  emailAndPasswordEnabled,
  signUpEnabled,
  socialAuthenticationProviders,
}: {
  emailAndPasswordEnabled: boolean;
  signUpEnabled: boolean;
  socialAuthenticationProviders: SocialAuthenticationProvider[];
}) {
  const t = useTranslations("Auth.SignIn");

  const loginMutation = useLoginMutation();
  const [isOAuthLoading, setIsOAuthLoading] = useState(false);
  const [oauthError, setOauthError] = useState<string | null>(null);
  const searchParams = useSearchParams();

  const [formData, setFormData] = useObjectState({
    email: "",
    password: "",
  });

  // Check for OAuth errors in URL parameters
  useEffect(() => {
    const error = searchParams.get('error');
    if (error) {
      setOauthError(decodeURIComponent(error));
      // Clear the error from URL after a delay
      setTimeout(() => {
        setOauthError(null);
        // Clear URL without refresh
        window.history.replaceState({}, document.title, window.location.pathname);
      }, 5000);
    }
  }, [searchParams]);

  const emailAndPasswordSignIn = () => {
    loginMutation.mutate({
      email: formData.email,
      password: formData.password,
    });
  };

  const handleSocialSignIn = async (provider: SocialAuthenticationProvider) => {
    try {
      setIsOAuthLoading(true);
      setOauthError(null); // Clear any previous errors
      
      // Validate provider before initiating OAuth
      if (!['github', 'google'].includes(provider)) {
        throw new Error(`Unsupported OAuth provider: ${provider}`);
      }
      
      // Initiate OAuth flow - this will redirect to the provider
      initiateOAuth(provider as 'github' | 'google');
      
    } catch (error) {
      console.error('OAuth initiation failed:', error);
      setOauthError(error instanceof Error ? error.message : 'OAuth initialization failed');
      setIsOAuthLoading(false);
    }
  };
  return (
    <div className="w-full h-full flex flex-col p-4 md:p-8 justify-center">
      <Card className="w-full md:max-w-md bg-background border-none mx-auto shadow-none animate-in fade-in duration-1000">
        <CardHeader className="my-4">
          <CardTitle className="text-2xl text-center my-1">
            {t("title")}
          </CardTitle>
          <CardDescription className="text-center text-muted-foreground">
            {t("description")}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col">
          {/* OAuth Error Display */}
          {oauthError && (
            <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md">
              <div className="flex items-start gap-2">
                <svg className="w-4 h-4 text-destructive mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
                <div className="text-sm">
                  <p className="font-medium text-destructive">Authentication Failed</p>
                  <p className="text-destructive/80">{oauthError}</p>
                </div>
              </div>
            </div>
          )}
          
          {/* Login Error Display */}
          {loginMutation.isError && (
            <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md">
              <div className="flex items-start gap-2">
                <svg className="w-4 h-4 text-destructive mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
                <div className="text-sm">
                  <p className="font-medium text-destructive">Login Failed</p>
                  <p className="text-destructive/80">{loginMutation.error?.message || 'Please check your credentials and try again.'}</p>
                </div>
              </div>
            </div>
          )}
          
          {emailAndPasswordEnabled && (
            <div className="flex flex-col gap-6">
              <div className="grid gap-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  autoFocus
                  disabled={loginMutation.isPending}
                  value={formData.email}
                  onChange={(e) => setFormData({ email: e.target.value })}
                  type="email"
                  placeholder="user@example.com"
                  required
                />
              </div>
              <div className="grid gap-2">
                <div className="flex items-center">
                  <Label htmlFor="password">Password</Label>
                </div>
                <Input
                  id="password"
                  disabled={loginMutation.isPending}
                  value={formData.password}
                  placeholder="********"
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      emailAndPasswordSignIn();
                    }
                  }}
                  onChange={(e) => setFormData({ password: e.target.value })}
                  type="password"
                  required
                />
              </div>
              <Button
                className="w-full"
                onClick={emailAndPasswordSignIn}
                disabled={loginMutation.isPending}
              >
                {loginMutation.isPending ? (
                  <Loader className="size-4 animate-spin ml-1" />
                ) : (
                  t("signIn")
                )}
              </Button>
            </div>
          )}
          {socialAuthenticationProviders.length > 0 && (
            <>
              {emailAndPasswordEnabled && (
                <div className="flex items-center my-4">
                  <div className="flex-1 h-px bg-accent"></div>
                  <span className="px-4 text-sm text-muted-foreground">
                    {t("orContinueWith")}
                  </span>
                  <div className="flex-1 h-px bg-accent"></div>
                </div>
              )}
              <div className="flex flex-col gap-2 w-full">
                {socialAuthenticationProviders.includes("google") && (
                  <Button
                    variant="outline"
                    onClick={() => handleSocialSignIn("google")}
                    className="flex-1 w-full"
                    disabled={isOAuthLoading || loginMutation.isPending}
                  >
                    {isOAuthLoading ? (
                      <Loader className="size-4 animate-spin" />
                    ) : (
                      <>
                        <GoogleIcon className="size-4 fill-foreground" />
                        Google
                      </>
                    )}
                  </Button>
                )}
                {socialAuthenticationProviders.includes("github") && (
                  <Button
                    variant="outline"
                    onClick={() => handleSocialSignIn("github")}
                    className="flex-1 w-full"
                    disabled={isOAuthLoading || loginMutation.isPending}
                  >
                    {isOAuthLoading ? (
                      <Loader className="size-4 animate-spin" />
                    ) : (
                      <>
                        <GithubIcon className="size-4 fill-foreground" />
                        GitHub
                      </>
                    )}
                  </Button>
                )}
              </div>
            </>
          )}
          {signUpEnabled && (
            <div className="my-8 text-center text-sm text-muted-foreground">
              {t("noAccount")}
              <Link href="/sign-up" className="underline-offset-4 text-primary">
                {t("signUp")}
              </Link>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
