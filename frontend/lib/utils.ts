import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { z } from "zod";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const createEmitter = () => {
  const listeners = new Set<(value: string) => void>();
  return {
    on: (listener: (value: string) => void) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    off: (listener: (value: string) => void) => {
      listeners.delete(listener);
    },
    emit: (value: string) => {
      listeners.forEach((listener) => listener(value));
    },
  };
};

export const envBooleanSchema = z
  .union([z.string(), z.boolean()])
  .optional()
  .transform((val) => {
    if (typeof val === "boolean") return val;
    if (typeof val === "string") {
      const lowerVal = val.toLowerCase();
      return lowerVal === "true" || lowerVal === "1" || lowerVal === "y";
    }
    return false;
  });

export function errorToString(error: unknown) {
  if (error == null) {
    return "unknown error";
  }

  if (typeof error === "string") {
    return error;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return JSON.stringify(error);
}

export function parseEnvBoolean(value: string | boolean | undefined): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    const lowerVal = value.toLowerCase();
    return lowerVal === "true" || lowerVal === "1" || lowerVal === "y";
  }
  return false;
}


export function capitalizeFirstLetter(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export function generateUUID(): string {
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

export const createDebounce = () => {
  let timeout: ReturnType<typeof setTimeout>;

  const debounce = (func: (...args: any[]) => any, waitFor = 200) => {
    clearTimeout(timeout!);
    timeout = setTimeout(() => func(), waitFor);
    return timeout;
  };

  debounce.clear = () => {
    clearTimeout(timeout!);
  };
  return debounce;
};

export const createThrottle = () => {
  let lastCall = 0;
  return (func: (...args: any[]) => any, waitFor = 200) => {
    const now = Date.now();
    if (now - lastCall >= waitFor) {
      lastCall = now;
      func();
    }
  };
};