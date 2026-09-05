import type {
  Asset, Brand, Caption, Carousel, Design, DesignView, ExportJob, Form,
  ImageCandidate, Job, Project, ProjectContext, Registry, User,
} from '@/types';

const BASE = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
    readonly body?: unknown,
  ) {
    super(message);
  }
}

type RequestOptions = {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
  signal?: AbortSignal;
};

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const res = await fetch(`${BASE}/api/v1${path}`, {
    method: opts.method ?? 'GET',
    // The session lives in an httpOnly cookie, so every call is credentialed.
    credentials: 'include',
    headers: {
      ...(opts.body !== undefined ? { 'Content-Type': 'application/json' } : {}),
      ...opts.headers,
    },
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
    signal: opts.signal,
  });

  if (res.status === 204) return undefined as T;

  const text = await res.text();
  const parsed = text ? JSON.parse(text) : undefined;

  if (!res.ok) {
    const err = parsed?.error ?? {};
    // The message is a server-side fallback; the UI localises by `code` first.
    throw new ApiError(res.status, err.code ?? 'unknown', err.message ?? '', parsed);
  }
  return parsed as T;
}

export const api = {
  authConfig: () => request<{ google_enabled: boolean; dev_login_enabled: boolean }>('/auth/config'),
  me: () => request<User>('/auth/me'),
  devLogin: (email: string, name?: string) =>
    request<User>('/auth/dev-login', { method: 'POST', body: { email, name } }),
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
  googleLoginUrl: () => `${BASE}/api/v1/auth/google`,

  listProjects: () => request<{ projects: Project[] }>('/projects'),
  createProject: (niche: string, language: string) =>
    request<Project>('/projects', { method: 'POST', body: { niche, language } }),
  getProject: (id: string) => request<Project>(`/projects/${id}`),
  updateProject: (
    id: string,
    patch: { name?: string; niche?: string; description?: string; language?: string; brand?: Brand; context?: ProjectContext },
  ) => request<Project>(`/projects/${id}`, { method: 'PATCH', body: patch }),
  deleteProject: (id: string) => request<void>(`/projects/${id}`, { method: 'DELETE' }),

  getSchema: (projectId: string) => request<{ version: number; form: Form }>(`/projects/${projectId}/schema`),
  regenerateSchema: (projectId: string) =>
    request<{ version: number; form: Form }>(`/projects/${projectId}/generate-schema`, { method: 'POST' }),

  listCarousels: (projectId: string, params: { status?: string; q?: string } = {}) => {
    const qs = new URLSearchParams();
    if (params.status && params.status !== 'all') qs.set('status', params.status);
    if (params.q) qs.set('q', params.q);
    const suffix = qs.toString() ? `?${qs}` : '';
    return request<{ carousels: Carousel[] }>(`/projects/${projectId}/carousels${suffix}`);
  },
  generateCarousel: (
    projectId: string,
    inputs: Record<string, unknown>,
    ratio: string,
    idempotencyKey: string,
  ) =>
    request<{ carousel_id: string; carousel: Carousel; job_id: string; status: string }>(
      `/projects/${projectId}/carousels/generate`,
      { method: 'POST', body: { inputs, ratio }, headers: { 'Idempotency-Key': idempotencyKey } },
    ),
  getCarousel: (id: string) => request<{ carousel: Carousel; job?: Job }>(`/carousels/${id}`),
  updateCarousel: (id: string, patch: { title?: string; status?: string }) =>
    request<Carousel>(`/carousels/${id}`, { method: 'PATCH', body: patch }),
  deleteCarousel: (id: string) => request<void>(`/carousels/${id}`, { method: 'DELETE' }),
  duplicateCarousel: (id: string) => request<Carousel>(`/carousels/${id}/duplicate`, { method: 'POST' }),
  archiveCarousel: (id: string) => request<Carousel>(`/carousels/${id}/archive`, { method: 'POST' }),
  regenerateCarousel: (id: string) =>
    request<{ carousel_id: string; job_id: string }>(`/carousels/${id}/regenerate`, { method: 'POST' }),

  getJob: (id: string) => request<Job>(`/jobs/${id}`),

  getDesign: (carouselId: string) => request<DesignView>(`/carousels/${carouselId}/design`),
  saveDesign: (carouselId: string, version: number, design: Design) =>
    request<DesignView>(`/carousels/${carouselId}/design`, { method: 'PATCH', body: { version, design } }),

  slideImages: (carouselId: string, slideId: string) =>
    request<{ query: string; candidates: ImageCandidate[] }>(
      `/carousels/${carouselId}/slides/${slideId}/images`,
    ),
  searchMoreImages: (carouselId: string, slideId: string, query: string, page: number) =>
    request<{ query: string; page: number; candidates: ImageCandidate[] }>(
      `/carousels/${carouselId}/slides/${slideId}/images/search`,
      { method: 'POST', body: { query, page } },
    ),
  selectImage: (carouselId: string, slideId: string, candidateId: string, version: number) =>
    request<{ version: number; asset: Asset }>(
      `/carousels/${carouselId}/slides/${slideId}/images/select`,
      { method: 'POST', body: { candidate_id: candidateId, version } },
    ),

  uploadAsset: async (file: File, projectId?: string, carouselId?: string) => {
    const form = new FormData();
    form.append('file', file);
    if (projectId) form.append('project_id', projectId);
    if (carouselId) form.append('carousel_id', carouselId);
    const res = await fetch(`${BASE}/api/v1/assets/upload`, {
      method: 'POST',
      credentials: 'include',
      body: form,
    });
    const parsed = await res.json();
    if (!res.ok) {
      throw new ApiError(res.status, parsed?.error?.code ?? 'unknown', parsed?.error?.message ?? '');
    }
    return parsed as Asset;
  },

  getCaption: (carouselId: string) => request<Caption>(`/carousels/${carouselId}/caption`),
  updateCaption: (carouselId: string, caption: string, hashtags: string[]) =>
    request<Caption>(`/carousels/${carouselId}/caption`, { method: 'PATCH', body: { caption, hashtags } }),
  regenerateCaption: (carouselId: string) =>
    request<Caption>(`/carousels/${carouselId}/caption/generate`, { method: 'POST' }),

  applyBrand: (carouselId: string) =>
    request<DesignView>(`/carousels/${carouselId}/apply-brand`, { method: 'POST' }),

  createExport: (carouselId: string) =>
    request<ExportJob>(`/carousels/${carouselId}/export`, { method: 'POST' }),
  getExport: (id: string) => request<ExportJob>(`/exports/${id}`),

  registry: () => request<Registry>('/registry'),
};
