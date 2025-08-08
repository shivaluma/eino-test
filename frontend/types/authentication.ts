import { z } from "zod";
import { envBooleanSchema } from "@/lib/utils";

export const SocialAuthenticationProviderSchema = z.enum([
  "github",
  "google",
  "microsoft",
]);

export type SocialAuthenticationProvider = z.infer<
  typeof SocialAuthenticationProviderSchema
>;

export const GitHubConfigSchema = z.object({
  enabled: z.boolean().default(false),
});

export const GoogleConfigSchema = z.object({
  enabled: z.boolean().default(false),
});

export const MicrosoftConfigSchema = z.object({
  enabled: z.boolean().default(false),
});

export const SocialAuthenticationConfigSchema = z.object({
  github: GitHubConfigSchema.optional(),
  google: GoogleConfigSchema.optional(),
  microsoft: MicrosoftConfigSchema.optional(),
});

export const AuthConfigSchema = z.object({
  emailAndPasswordEnabled: envBooleanSchema.default(true),
  signUpEnabled: envBooleanSchema.default(true),
  socialAuthenticationProviders: SocialAuthenticationConfigSchema,
});

export type GitHubConfig = z.infer<typeof GitHubConfigSchema>;
export type GoogleConfig = z.infer<typeof GoogleConfigSchema>;
export type MicrosoftConfig = z.infer<typeof MicrosoftConfigSchema>;
export type SocialAuthenticationConfig = z.infer<
  typeof SocialAuthenticationConfigSchema
>;

export type AuthConfig = z.infer<typeof AuthConfigSchema>;
