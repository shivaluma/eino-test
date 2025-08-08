import {
  GitHubConfigSchema,
  GoogleConfigSchema,
  MicrosoftConfigSchema,
  GitHubConfig,
  GoogleConfig,
  MicrosoftConfig,
  AuthConfig,
  AuthConfigSchema,
} from "@/types/authentication";
import { parseEnvBoolean } from "@/lib/utils";

function parseSocialAuthConfigs() {
  const configs: {
    github?: GitHubConfig;
    google?: GoogleConfig;
    microsoft?: MicrosoftConfig;
  } = {};

  if (process.env.GITHUB_ENABLED === "true") {
    const githubResult = GitHubConfigSchema.safeParse({
      enabled: true,
    });
    if (githubResult.success) {
      configs.github = githubResult.data;
    }
  }

  if (process.env.GOOGLE_ENABLED === "true") {
    const googleResult = GoogleConfigSchema.safeParse({
      enabled: true,
    });

    if (googleResult.success) {
      configs.google = googleResult.data;
    }
  }

    if (process.env.MICROSOFT_ENABLED === "true") {
    const microsoftResult = MicrosoftConfigSchema.safeParse({
      enabled: true,
    });

    if (microsoftResult.success) {
      configs.microsoft = microsoftResult.data;
    }
  }

  return configs;
}

export function getAuthConfig(): AuthConfig {
  const rawConfig = {
    emailAndPasswordEnabled: process.env.DISABLE_EMAIL_SIGN_IN
      ? !parseEnvBoolean(process.env.DISABLE_EMAIL_SIGN_IN)
      : true,
    signUpEnabled: process.env.DISABLE_SIGN_UP
      ? !parseEnvBoolean(process.env.DISABLE_SIGN_UP)
      : true,
    socialAuthenticationProviders: parseSocialAuthConfigs(),
  };

  const result = AuthConfigSchema.safeParse(rawConfig);

  if (!result.success) {
    throw new Error(`Invalid auth configuration: ${result.error.message}`);
  }

  return result.data;
}
