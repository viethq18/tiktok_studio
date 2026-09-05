'use client';

/**
 * Brand kit editor (§39).
 *
 * The placement defaults encode a recommendation so a creator does not have to
 * reason about layout: the handle rides every slide because a viewer can enter
 * the carousel anywhere and it is what converts them into a follower; the logo
 * sits on the first and last slide only because it carries more visual weight;
 * the website appears beside the call to action, where the viewer has already
 * decided to act. Each is overridable.
 */
import { Upload } from 'lucide-react';
import * as React from 'react';

import { Button, ErrorNote, Input, Label, Spinner } from '@/components/ui/primitives';
import { api, ApiError } from '@/lib/api/client';
import { useT } from '@/lib/i18n';
import type { Brand, BrandPosition, BrandScope, Placement } from '@/types';
import { DEFAULT_BRAND_DISPLAY } from '@/types';

const SCOPES: BrandScope[] = ['all', 'first_last', 'first', 'last', 'off'];
const POSITIONS: BrandPosition[] = [
  'top_left', 'top_center', 'top_right',
  'bottom_left', 'bottom_center', 'bottom_right',
];

export function BrandPanel({
  projectId,
  brand,
  onChange,
}: {
  projectId: string;
  brand: Brand;
  onChange: (next: Brand) => void;
}) {
  const t = useT();
  const fileInput = React.useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = React.useState(false);
  const [uploadError, setUploadError] = React.useState('');

  const display = brand.display ?? DEFAULT_BRAND_DISPLAY;
  const setDisplay = (key: keyof typeof display, patch: Partial<Placement>) =>
    onChange({ ...brand, display: { ...display, [key]: { ...display[key], ...patch } } });

  async function uploadLogo(file: File) {
    setUploading(true);
    setUploadError('');
    try {
      const asset = await api.uploadAsset(file, projectId);
      onChange({ ...brand, logo_asset_id: asset.id, logo_url: asset.url });
    } catch (err) {
      setUploadError(err instanceof ApiError ? err.message : t('brand.logoUploadFailed'));
    } finally {
      setUploading(false);
    }
  }

  return (
    <div className="space-y-8">
      <section className="space-y-5">
        <p className="text-sm text-neutral-500">{t('brand.colorsHint')}</p>
        {(
          [
            ['primary_color', t('brand.primary')],
            ['secondary_color', t('brand.secondary')],
            ['accent_color', t('brand.accent')],
          ] as const
        ).map(([key, label]) => (
          <div key={key} className="space-y-2">
            <Label>{label}</Label>
            <div className="flex items-center gap-3">
              <input
                type="color"
                aria-label={label}
                value={brand[key] || '#111111'}
                onChange={(e) => onChange({ ...brand, [key]: e.target.value.toUpperCase() })}
                className="h-10 w-14 cursor-pointer rounded border border-neutral-200"
              />
              <Input
                value={brand[key] ?? ''}
                placeholder="#111111"
                onChange={(e) => onChange({ ...brand, [key]: e.target.value })}
              />
              {brand[key] ? (
                <Button variant="ghost" size="sm" onClick={() => onChange({ ...brand, [key]: '' })}>
                  {t('common.clear')}
                </Button>
              ) : null}
            </div>
          </div>
        ))}
      </section>

      <section className="space-y-5 border-t border-neutral-100 pt-6">
        <div>
          <h3 className="text-sm font-medium">{t('brand.identityTitle')}</h3>
          <p className="mt-1 text-sm text-neutral-500">{t('brand.identityHint')}</p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="handle">{t('brand.handle')}</Label>
          <Input
            id="handle"
            value={brand.handle ?? ''}
            placeholder="@babycare.vn"
            onChange={(e) => onChange({ ...brand, handle: e.target.value })}
          />
          <PlacementRow
            placement={display.handle}
            onChange={(patch) => setDisplay('handle', patch)}
            recommendation={t('brand.handleRecommendation')}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="website">{t('brand.website')}</Label>
          <Input
            id="website"
            value={brand.website ?? ''}
            placeholder="babycare.vn"
            onChange={(e) => onChange({ ...brand, website: e.target.value })}
          />
          <PlacementRow
            placement={display.website}
            onChange={(patch) => setDisplay('website', patch)}
            recommendation={t('brand.websiteRecommendation')}
          />
        </div>

        <div className="space-y-2">
          <Label>{t('brand.logo')}</Label>
          <div className="flex items-center gap-3">
            {brand.logo_url ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={brand.logo_url}
                alt=""
                className="h-16 w-16 rounded-lg border border-neutral-200 bg-neutral-50 object-contain p-1"
              />
            ) : (
              <div className="flex h-16 w-16 items-center justify-center rounded-lg border border-dashed border-neutral-300 text-xs text-neutral-400">
                {t('brand.noLogo')}
              </div>
            )}
            <input
              ref={fileInput}
              type="file"
              accept="image/png,image/jpeg,image/webp"
              className="hidden"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) void uploadLogo(file);
                e.target.value = '';
              }}
            />
            <Button variant="secondary" size="sm" disabled={uploading} onClick={() => fileInput.current?.click()}>
              {uploading ? <Spinner className="h-4 w-4" /> : <Upload className="h-4 w-4" />}
              {brand.logo_asset_id ? t('brand.replaceLogo') : t('brand.uploadLogo')}
            </Button>
            {brand.logo_asset_id ? (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onChange({ ...brand, logo_asset_id: '', logo_url: '' })}
              >
                {t('common.remove')}
              </Button>
            ) : null}
          </div>
          <p className="text-xs text-neutral-400">{t('brand.logoHint')}</p>
          <PlacementRow
            placement={display.logo}
            onChange={(patch) => setDisplay('logo', patch)}
            recommendation={t('brand.logoRecommendation')}
          />
        </div>
        <ErrorNote>{uploadError}</ErrorNote>
      </section>
    </div>
  );
}

function PlacementRow({
  placement,
  onChange,
  recommendation,
}: {
  placement: Placement;
  onChange: (patch: Partial<Placement>) => void;
  recommendation: string;
}) {
  const t = useT();
  return (
    <div className="rounded-lg bg-neutral-50 p-3">
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-2">
          <span className="text-xs text-neutral-500">{t('brand.showOn')}</span>
          <select
            className="h-8 rounded border border-neutral-200 bg-white px-2 text-xs"
            value={placement.scope}
            onChange={(e) => onChange({ scope: e.target.value as BrandScope })}
          >
            {SCOPES.map((s) => (
              <option key={s} value={s}>
                {t(`brand.scope.${s}`)}
              </option>
            ))}
          </select>
        </div>
        {placement.scope !== 'off' ? (
          <div className="flex items-center gap-2">
            <span className="text-xs text-neutral-500">{t('brand.position')}</span>
            <select
              className="h-8 rounded border border-neutral-200 bg-white px-2 text-xs"
              value={placement.position}
              onChange={(e) => onChange({ position: e.target.value as BrandPosition })}
            >
              {POSITIONS.map((p) => (
                <option key={p} value={p}>
                  {t(`brand.position.${p}`)}
                </option>
              ))}
            </select>
          </div>
        ) : null}
      </div>
      <p className="mt-2 text-xs text-neutral-400">{recommendation}</p>
    </div>
  );
}
