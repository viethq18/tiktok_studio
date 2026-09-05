'use client';

/**
 * Create carousel (§108, §109): dynamic form on the left, generation progress
 * on the right. The form is whatever the AI decided this niche needs.
 */
import { useMutation } from '@tanstack/react-query';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import * as React from 'react';

import { GenerationProgress } from '@/components/generation/GenerationProgress';
import { AppShell } from '@/components/project/AppShell';
import { DynamicForm, defaultValues, validate, type FormValues } from '@/components/schema-form/DynamicForm';
import { Button, Card, ErrorNote, Spinner } from '@/components/ui/primitives';
import { api } from '@/lib/api/client';
import { useErrorMessage, useT } from '@/lib/i18n';
import { useJob, useProject, useSchema } from '@/lib/api/hooks';
import { cn, uuid } from '@/lib/utils';

const RATIOS = [
  { ratio: '4:5', size: '1080 × 1350', recommended: true },
  { ratio: '1:1', size: '1080 × 1080', recommended: false },
  { ratio: '9:16', size: '1080 × 1920', recommended: false },
];

export default function NewCarouselPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const router = useRouter();
  const t = useT();
  const errorMessage = useErrorMessage();
  const { data: project } = useProject(projectId);
  const { data: schema, isLoading: schemaLoading, error: schemaError } = useSchema(projectId);

  const [values, setValues] = React.useState<FormValues>({});
  const [seededVersion, setSeededVersion] = React.useState<number | null>(null);
  const [errors, setErrors] = React.useState<Record<string, string>>({});
  const [ratio, setRatio] = React.useState('4:5');
  const [jobId, setJobId] = React.useState<string | null>(null);
  const [carouselId, setCarouselId] = React.useState<string | null>(null);

  // One idempotency key per visit: a double click cannot start two generations (§159).
  const idempotencyKey = React.useRef(uuid());

  // Seed the form once per schema version, without an effect round-trip.
  if (schema?.form && schema.version !== seededVersion) {
    setSeededVersion(schema.version);
    setValues(defaultValues(schema.form));
  }

  const { data: job } = useJob(jobId);

  React.useEffect(() => {
    if (job?.status === 'completed' && carouselId) {
      router.replace(`/carousels/${carouselId}`);
    }
  }, [job?.status, carouselId, router]);

  const generate = useMutation({
    mutationFn: () => api.generateCarousel(projectId, values, ratio, idempotencyKey.current),
    onSuccess: (res) => {
      setJobId(res.job_id);
      setCarouselId(res.carousel_id);
    },
  });

  function submit() {
    if (!schema?.form) return;
    const found = validate(schema.form, values, t);
    setErrors(found);
    if (Object.keys(found).length > 0) return;
    generate.mutate();
  }

  return (
    <AppShell
      breadcrumb={
        <Link href={`/projects/${projectId}`} className="hover:text-neutral-900">
          {project?.name ?? 'Project'}
        </Link>
      }
    >
      <div className="grid gap-6 lg:grid-cols-[minmax(0,380px)_1fr]">
        <Card className="p-6">
          <h1 className="text-lg font-semibold tracking-tight">{t('create.title')}</h1>
          <p className="mt-1 text-sm text-neutral-500">{t('create.subtitle')}</p>

          {schemaLoading ? (
            <div className="mt-10 flex flex-col items-center gap-3 text-sm text-neutral-500">
              <Spinner />
              {t('create.preparingForm')}
            </div>
          ) : schemaError ? (
            <div className="mt-6">
              <ErrorNote>{t('create.formFailed')}</ErrorNote>
            </div>
          ) : schema?.form ? (
            <div className="mt-6">
              <DynamicForm
                form={schema.form}
                values={values}
                errors={errors}
                onChange={(name, value) => setValues((v) => ({ ...v, [name]: value }))}
                disabled={Boolean(jobId)}
              />

              <div className="mt-6 space-y-2">
                <p className="text-sm font-medium">{t('create.ratio')}</p>
                {RATIOS.map((r) => (
                  <label
                    key={r.ratio}
                    className={cn(
                      'flex cursor-pointer items-center gap-3 rounded-lg border px-3 py-2.5 text-sm transition',
                      ratio === r.ratio ? 'border-neutral-900 bg-neutral-50' : 'border-neutral-200 hover:bg-neutral-50',
                    )}
                  >
                    <input
                      type="radio"
                      name="ratio"
                      checked={ratio === r.ratio}
                      onChange={() => setRatio(r.ratio)}
                      className="h-4 w-4 accent-neutral-900"
                    />
                    <span className="font-medium">{r.ratio}</span>
                    <span className="text-neutral-500">{r.size}</span>
                    {r.recommended ? (
                      <span className="ml-auto rounded-full bg-emerald-50 px-2 py-0.5 text-xs text-emerald-700">
                        {t('create.recommended')}
                      </span>
                    ) : null}
                  </label>
                ))}
              </div>

              <Button
                size="lg"
                className="mt-6 w-full"
                disabled={generate.isPending || Boolean(jobId)}
                onClick={submit}
              >
                {jobId ? t('create.submitting') : t('create.submit')}
              </Button>
              <div className="mt-3">
                <ErrorNote>
                  {generate.error ? errorMessage(generate.error) : ''}
                </ErrorNote>
              </div>
            </div>
          ) : null}
        </Card>

        <Card className="flex min-h-[420px] items-center justify-center p-8">
          {jobId ? (
            <div className="w-full max-w-sm">
              <GenerationProgress job={job} />
              {job?.status === 'failed' ? (
                <Button
                  variant="secondary"
                  className="mt-4 w-full"
                  onClick={() => {
                    idempotencyKey.current = uuid();
                    setJobId(null);
                    setCarouselId(null);
                  }}
                >
                  {t('common.retry')}
                </Button>
              ) : null}
            </div>
          ) : (
            <div className="max-w-sm text-center">
              <p className="text-sm font-medium text-neutral-900">{t('create.idleTitle')}</p>
              <p className="mt-2 text-sm text-neutral-500">{t('create.idleBody')}</p>
            </div>
          )}
        </Card>
      </div>
    </AppShell>
  );
}
