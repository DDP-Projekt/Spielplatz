import type { RequestHandler } from "@sveltejs/kit";

export const prerender = false;

import { PUBLIC_BACKEND_HOST } from '$env/static/public';

const proxy: RequestHandler = async ({ params, url, request, fetch }) => {
  const target = new URL(`${url.protocol}${PUBLIC_BACKEND_HOST}/api/${params.path ?? ''}${url.search}`);

  return fetch(new Request(target, request));
};

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
export const OPTIONS = proxy;