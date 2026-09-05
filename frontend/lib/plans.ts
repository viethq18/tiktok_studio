/**
 * Plan definitions for the pricing page.
 *
 * Billing is not implemented: nothing here is enforced by the backend yet, and
 * the page says so. Limits are expressed against features that actually exist
 * so the ladder stays honest rather than promising capabilities we do not have.
 */
import type { TranslationKey } from '@/lib/i18n';

export type PlanId = 'free' | 'pro' | 'max';

export type PlanFeature = { key: TranslationKey; vars?: Record<string, string | number> };

export type Plan = {
  id: PlanId;
  nameKey: TranslationKey;
  taglineKey: TranslationKey;
  /** USD per month when billed monthly. */
  monthly: number;
  popular?: boolean;
  features: PlanFeature[];
};

/** Yearly billing is 20% off the monthly rate. */
export const YEARLY_DISCOUNT = 0.2;

export const PLANS: Plan[] = [
  {
    id: 'free',
    nameKey: 'pricing.free.name',
    taglineKey: 'pricing.free.tagline',
    monthly: 0,
    features: [
      { key: 'plan.projects', vars: { n: 1 } },
      { key: 'plan.carousels', vars: { n: 5 } },
      { key: 'plan.ratios' },
      { key: 'plan.export' },
      { key: 'plan.caption' },
      { key: 'plan.languages' },
    ],
  },
  {
    id: 'pro',
    nameKey: 'pricing.pro.name',
    taglineKey: 'pricing.pro.tagline',
    monthly: 20,
    popular: true,
    features: [
      { key: 'plan.projects', vars: { n: 10 } },
      { key: 'plan.carousels', vars: { n: 100 } },
      { key: 'plan.brand' },
      { key: 'plan.images' },
      { key: 'plan.regenerate' },
      { key: 'plan.editor' },
    ],
  },
  {
    id: 'max',
    nameKey: 'pricing.max.name',
    taglineKey: 'pricing.max.tagline',
    monthly: 100,
    features: [
      { key: 'plan.projectsUnlimited' },
      { key: 'plan.carousels', vars: { n: 500 } },
      { key: 'plan.priority' },
      { key: 'plan.history' },
      { key: 'plan.earlyAccess' },
      { key: 'plan.support' },
    ],
  },
];

/** Monthly-equivalent price for the chosen billing period. */
export function monthlyPrice(plan: Plan, yearly: boolean): number {
  return yearly ? Math.round(plan.monthly * (1 - YEARLY_DISCOUNT)) : plan.monthly;
}

export function yearlyTotal(plan: Plan): number {
  return monthlyPrice(plan, true) * 12;
}

export function formatUSD(amount: number): string {
  return `$${amount.toLocaleString('en-US')}`;
}

/** Rows of the comparison table: a feature and how each plan answers it. */
export type ComparisonRow = { key: TranslationKey; values: Record<PlanId, string | boolean> };

export const COMPARISON: ComparisonRow[] = [
  { key: 'compare.projects', values: { free: '1', pro: '10', max: '∞' } },
  { key: 'compare.carousels', values: { free: '5', pro: '100', max: '500' } },
  { key: 'compare.languages', values: { free: true, pro: true, max: true } },
  { key: 'compare.ratios', values: { free: true, pro: true, max: true } },
  { key: 'compare.export', values: { free: true, pro: true, max: true } },
  { key: 'compare.caption', values: { free: true, pro: true, max: true } },
  { key: 'compare.upload', values: { free: false, pro: true, max: true } },
  { key: 'compare.brand', values: { free: false, pro: true, max: true } },
  { key: 'compare.regenerate', values: { free: false, pro: true, max: true } },
  { key: 'compare.history', values: { free: false, pro: false, max: true } },
  { key: 'compare.priority', values: { free: false, pro: false, max: true } },
  { key: 'compare.support', values: { free: false, pro: false, max: true } },
];
