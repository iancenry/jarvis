import {
  DEFAULT_AUTH_REDIRECT,
  getAuthRouteUrl,
  getPostAuthRedirect,
  parseAuthSearch,
  sanitizeRedirectTo,
} from "./-auth";
import { describe, expect, it } from "vitest";

describe("sanitizeRedirectTo", () => {
  it("accepts safe internal redirects", () => {
    expect(sanitizeRedirectTo("/todos?page=2#filters")).toBe(
      "/todos?page=2#filters",
    );
  });

  it("rejects external, blank, and auth-route redirects", () => {
    expect(sanitizeRedirectTo("")).toBeUndefined();
    expect(sanitizeRedirectTo("https://example.com")).toBeUndefined();
    expect(sanitizeRedirectTo("//example.com")).toBeUndefined();
    expect(sanitizeRedirectTo("/sign-in")).toBeUndefined();
    expect(sanitizeRedirectTo("/sign-up?redirect=/dashboard")).toBeUndefined();
  });
});

describe("auth redirect helpers", () => {
  it("parses and sanitizes auth search params", () => {
    expect(parseAuthSearch({ redirect: "/dashboard" })).toEqual({
      redirect: "/dashboard",
    });
    expect(parseAuthSearch({ redirect: "https://example.com" })).toEqual({
      redirect: undefined,
    });
  });

  it("falls back to the default redirect when needed", () => {
    expect(getPostAuthRedirect(undefined)).toBe(DEFAULT_AUTH_REDIRECT);
    expect(getPostAuthRedirect("/todos")).toBe("/todos");
  });

  it("builds auth URLs only with safe redirects", () => {
    expect(getAuthRouteUrl("/sign-in", "/todos?page=2")).toBe(
      "/sign-in?redirect=%2Ftodos%3Fpage%3D2",
    );
    expect(getAuthRouteUrl("/sign-up", "https://example.com")).toBe("/sign-up");
  });
});
