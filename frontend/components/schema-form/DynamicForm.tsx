'use client';

/**
 * Schema renderer (§95).
 *
 * The AI decides WHICH fields exist; this file decides HOW each one is drawn.
 * The registry below is closed — a component the backend did not approve simply
 * has no entry here, so it cannot be rendered (§9.1, §11).
 */
import * as React from 'react';

import { Input, Label, Textarea } from '@/components/ui/primitives';
import { useT } from '@/lib/i18n';
import { cn } from '@/lib/utils';
import type { FieldComponent, Form, FormProperty } from '@/types';

export type FormValues = Record<string, unknown>;

export function defaultValues(form: Form): FormValues {
  const out: FormValues = {};
  for (const p of form.properties) {
    if (p.default !== undefined && p.default !== null) out[p.name] = p.default;
    else if (p.type === 'array') out[p.name] = [];
    else if (p.type === 'boolean') out[p.name] = false;
    else if (p.component === 'slider' && p.minimum !== undefined) out[p.name] = p.minimum;
    else out[p.name] = '';
  }
  return out;
}

type Translate = (key: Parameters<ReturnType<typeof useT>>[0], vars?: Record<string, string | number>) => string;

/** Local validation is for feedback only; the backend re-validates (§96). */
export function validate(form: Form, values: FormValues, t: Translate): Record<string, string> {
  const errors: Record<string, string> = {};
  for (const p of form.properties) {
    const v = values[p.name];
    const empty = v === undefined || v === null || v === '' || (Array.isArray(v) && v.length === 0);
    if (p.required && empty) {
      errors[p.name] = `${p.title} ${t('create.required')}`;
      continue;
    }
    if (empty) continue;
    if (typeof v === 'string' && p.max_length && v.length > p.max_length) {
      errors[p.name] = t('create.maxLength', { n: p.max_length });
    }
    if (typeof v === 'number') {
      if (p.minimum !== undefined && v < p.minimum) errors[p.name] = t('create.min', { n: p.minimum });
      if (p.maximum !== undefined && v > p.maximum) errors[p.name] = t('create.max', { n: p.maximum });
    }
  }
  return errors;
}

type FieldProps = {
  property: FormProperty;
  value: unknown;
  error?: string;
  onChange: (value: unknown) => void;
};

const registry: Record<FieldComponent, React.ComponentType<FieldProps>> = {
  text: TextField,
  textarea: TextareaField,
  number: NumberField,
  select: SelectField,
  multi_select: MultiSelectField,
  radio: RadioField,
  checkbox: CheckboxField,
  slider: SliderField,
  image: ImageField,
  url: URLField,
};

function labelFor(p: FormProperty, value: string): string {
  return p.enum_labels?.[value] ?? value;
}

export function DynamicForm({
  form,
  values,
  errors,
  onChange,
  disabled,
}: {
  form: Form;
  values: FormValues;
  errors: Record<string, string>;
  onChange: (name: string, value: unknown) => void;
  disabled?: boolean;
}) {
  return (
    <div className={cn('space-y-6', disabled && 'pointer-events-none opacity-60')}>
      {form.properties.map((p) => {
        const Field = registry[p.component] ?? TextField;
        return (
          <div key={p.name} className="space-y-2">
            <Label htmlFor={p.name}>
              {p.title}
              {p.required ? <span className="ml-1 text-red-500">*</span> : null}
            </Label>
            {p.description ? <p className="text-xs text-neutral-500">{p.description}</p> : null}
            <Field
              property={p}
              value={values[p.name]}
              error={errors[p.name]}
              onChange={(v) => onChange(p.name, v)}
            />
            {p.help ? <p className="text-xs text-neutral-400">{p.help}</p> : null}
            {errors[p.name] ? <p className="text-xs text-red-600">{errors[p.name]}</p> : null}
          </div>
        );
      })}
    </div>
  );
}

function TextField({ property, value, onChange }: FieldProps) {
  return (
    <Input
      id={property.name}
      value={String(value ?? '')}
      placeholder={property.placeholder}
      maxLength={property.max_length}
      onChange={(e) => onChange(e.target.value)}
    />
  );
}

function URLField({ property, value, onChange }: FieldProps) {
  return (
    <Input
      id={property.name}
      type="url"
      value={String(value ?? '')}
      placeholder={property.placeholder ?? 'https://'}
      onChange={(e) => onChange(e.target.value)}
    />
  );
}

function TextareaField({ property, value, onChange }: FieldProps) {
  const text = String(value ?? '');
  return (
    <div>
      <Textarea
        id={property.name}
        rows={3}
        value={text}
        placeholder={property.placeholder}
        maxLength={property.max_length}
        onChange={(e) => onChange(e.target.value)}
      />
      {property.max_length ? (
        <p className="mt-1 text-right text-xs text-neutral-400">
          {text.length}/{property.max_length}
        </p>
      ) : null}
    </div>
  );
}

function NumberField({ property, value, onChange }: FieldProps) {
  return (
    <Input
      id={property.name}
      type="number"
      value={value === '' || value === undefined ? '' : Number(value)}
      min={property.minimum}
      max={property.maximum}
      onChange={(e) => onChange(e.target.value === '' ? '' : Number(e.target.value))}
    />
  );
}

function SelectField({ property, value, onChange }: FieldProps) {
  const t = useT();
  return (
    <select
      id={property.name}
      className="h-10 w-full rounded-lg border border-neutral-200 bg-white px-3 text-sm focus:border-neutral-900 focus:outline-none"
      value={String(value ?? '')}
      onChange={(e) => onChange(e.target.value)}
    >
      <option value="">{t('create.selectPlaceholder')}</option>
      {property.enum?.map((v) => (
        <option key={v} value={v}>
          {labelFor(property, v)}
        </option>
      ))}
    </select>
  );
}

function RadioField({ property, value, onChange }: FieldProps) {
  return (
    <div className="space-y-2">
      {property.enum?.map((v) => (
        <label
          key={v}
          className={cn(
            'flex cursor-pointer items-center gap-3 rounded-lg border px-3 py-2.5 text-sm transition',
            value === v ? 'border-neutral-900 bg-neutral-50' : 'border-neutral-200 hover:bg-neutral-50',
          )}
        >
          <input
            type="radio"
            name={property.name}
            value={v}
            checked={value === v}
            onChange={() => onChange(v)}
            className="h-4 w-4 accent-neutral-900"
          />
          <span>{labelFor(property, v)}</span>
        </label>
      ))}
    </div>
  );
}

function MultiSelectField({ property, value, onChange }: FieldProps) {
  const selected = Array.isArray(value) ? (value as string[]) : [];
  const toggle = (v: string) =>
    onChange(selected.includes(v) ? selected.filter((s) => s !== v) : [...selected, v]);
  return (
    <div className="flex flex-wrap gap-2">
      {property.enum?.map((v) => (
        <button
          key={v}
          type="button"
          onClick={() => toggle(v)}
          className={cn(
            'rounded-full border px-3 py-1.5 text-sm transition',
            selected.includes(v)
              ? 'border-neutral-900 bg-neutral-900 text-white'
              : 'border-neutral-200 text-neutral-700 hover:bg-neutral-50',
          )}
        >
          {labelFor(property, v)}
        </button>
      ))}
    </div>
  );
}

function CheckboxField({ property, value, onChange }: FieldProps) {
  return (
    <label className="flex cursor-pointer items-center gap-3 text-sm">
      <input
        type="checkbox"
        checked={Boolean(value)}
        onChange={(e) => onChange(e.target.checked)}
        className="h-4 w-4 accent-neutral-900"
      />
      <span>{property.placeholder ?? property.title}</span>
    </label>
  );
}

function SliderField({ property, value, onChange }: FieldProps) {
  const min = property.minimum ?? 0;
  const max = property.maximum ?? 10;
  const current = typeof value === 'number' ? value : min;
  return (
    <div>
      <input
        id={property.name}
        type="range"
        min={min}
        max={max}
        step={1}
        value={current}
        onChange={(e) => onChange(Number(e.target.value))}
        className="w-full accent-neutral-900"
      />
      <div className="flex justify-between text-xs text-neutral-500">
        <span>{min}</span>
        <span className="font-semibold text-neutral-900">{current}</span>
        <span>{max}</span>
      </div>
    </div>
  );
}

function ImageField({ property, value, onChange }: FieldProps) {
  return (
    <div className="space-y-2">
      <Input
        id={property.name}
        value={String(value ?? '')}
        placeholder={property.placeholder}
        onChange={(e) => onChange(e.target.value)}
      />
      {typeof value === 'string' && value ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img src={value} alt="" className="h-24 w-24 rounded-lg object-cover" />
      ) : null}
    </div>
  );
}


