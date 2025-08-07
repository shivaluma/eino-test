"use client";

import { createAuthClient } from "better-auth/react";
import { toast } from "sonner";
import { handleErrorWithToast } from "@/components/ui/shared-toast";

export const authClient = createAuthClient({
  fetchOptions: {
    onError(e) {
      if (e.error.status === 429) {
        toast.error("Too many requests. Please try again later.");
        return;
      }
      handleErrorWithToast(e.error);
    },
  },
});
