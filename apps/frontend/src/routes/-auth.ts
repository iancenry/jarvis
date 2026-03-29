export type AuthSearch = {
  redirect?: string;
};

export const DEFAULT_AUTH_REDIRECT = "/dashboard";

const AUTH_ROUTE_PREFIXES = ["/sign-in", "/sign-up"] as const;

export function sanitizeRedirectTo(value: unknown): string | undefined {
  if (typeof value !== "string") {
    return undefined;
  }

  const normalized = value.trim();

  if (
    normalized.length === 0 ||
    !normalized.startsWith("/") ||
    normalized.startsWith("//")
  ) {
    return undefined;
  }

  if (
    AUTH_ROUTE_PREFIXES.some(
      (prefix) =>
        normalized === prefix ||
        normalized.startsWith(`${prefix}?`) ||
        normalized.startsWith(`${prefix}#`),
    )
  ) {
    return undefined;
  }

  return normalized;
}

export function parseAuthSearch(search: Record<string, unknown>): AuthSearch {
  return {
    redirect: sanitizeRedirectTo(search.redirect),
  };
}

export function getPostAuthRedirect(redirectTo: unknown): string {
  return sanitizeRedirectTo(redirectTo) ?? DEFAULT_AUTH_REDIRECT;
}

export function getAuthRouteUrl(
  path: "/sign-in" | "/sign-up",
  redirectTo: unknown,
): string {
  const safeRedirect = sanitizeRedirectTo(redirectTo);

  if (!safeRedirect) {
    return path;
  }

  return `${path}?redirect=${encodeURIComponent(safeRedirect)}`;
}
