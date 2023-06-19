import { perfTestStore } from "@rilldata/web-common/features/models/workspace/perf-test-store";
import { Batcher } from "@rilldata/web-common/runtime-client/batcher/Batcher";
import type { FetchWrapperOptions } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { get } from "svelte/store";
import { HttpRequestQueue } from "./http-request-queue/HttpRequestQueue";
import { runtime } from "./runtime-store";

/**
 * Runtime base URL
 *  Local
 *    In dev & prod: http://localhost:9009
 *  Cloud
 *    In dev: http://localhost:9009
 *    In prod: https://{region}.runtime.rilldata.com
 */

export const httpRequestQueue = new HttpRequestQueue();
export const batcher = new Batcher();

export const httpClient = async <T>(
  config: FetchWrapperOptions
): Promise<T> => {
  // naive request interceptors

  // set host
  const host = get(runtime).host;
  const interceptedConfig = { ...config, baseUrl: host };

  // set jwt
  const jwt = get(runtime).jwt;
  if (jwt) {
    interceptedConfig.headers = {
      ...interceptedConfig.headers,
      Authorization: `Bearer ${jwt}`,
    };
  }

  if (get(perfTestStore).batch) {
    return (await batcher.add(interceptedConfig)) as Promise<T>;
  } else {
    return (await httpRequestQueue.add(interceptedConfig)) as Promise<T>;
  }
};

export default httpClient;
