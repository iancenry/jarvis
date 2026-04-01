import { env } from "@/config/env";
import { useAuth } from "@clerk/clerk-react";
import { apiContract } from "@jarvis/openapi/contracts";
import { initClient } from "@ts-rest/core";

type Headers = Awaited<
  ReturnType<NonNullable<Parameters<typeof initClient>[1]["api"]>>
>["headers"];

export type TApiClient = ReturnType<typeof useApiClient>;

export const useApiClient = ({ isBlob = false }: { isBlob?: boolean } = {}) => {
  const { getToken } = useAuth();

  return initClient(apiContract, {
    baseUrl: "",
    baseHeaders: {
      "Content-Type": "application/json",
    },
    api: async ({ path, method, headers, body }) => {
      const token = await getToken();

      const makeRequest = async (
        retryCount = 0,
      ): Promise<{
        status: number;
        body: unknown;
        headers: Headers;
      }> => {
        try {
          const response = await fetch(`${env.VITE_API_URL}/api${path}`, {
            method,
            headers: {
              ...headers,
              ...(token ? { Authorization: `Bearer ${token}` } : {}),
            },
            body: body ? JSON.stringify(body) : undefined,
          });

          const responseHeaders = Object.fromEntries(
            response.headers.entries(),
          ) as unknown as Headers;

          // If unauthorized and we haven't retried yet, retry
          if (response.status === 401 && retryCount < 2) {
            return makeRequest(retryCount + 1);
          }

          const data = isBlob
            ? await response.blob()
            : await response.json().catch(() => null);

          return {
            status: response.status,
            body: data,
            headers: responseHeaders,
          };
        } catch (e) {
          return {
            status: 500,
            body: { message: "Internal server error" },
            headers: {} as Headers,
          };
        }
      };

      return makeRequest();
    },
  });
};
