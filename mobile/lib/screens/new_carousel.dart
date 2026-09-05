import 'dart:async';

import 'package:flutter/material.dart';

import '../api/generated.dart';
import '../api/session.dart';
import '../i18n/strings.dart';
import '../widgets/common.dart';
import '../widgets/dynamic_form.dart';
import 'carousel_detail.dart';

const _ratios = [
  ('4:5', '1080 × 1350', true),
  ('1:1', '1080 × 1080', false),
  ('9:16', '1080 × 1920', false),
];

class NewCarouselScreen extends StatefulWidget {
  const NewCarouselScreen({super.key, required this.api, required this.project});

  final Api api;
  final Project project;

  @override
  State<NewCarouselScreen> createState() => _NewCarouselScreenState();
}

class _NewCarouselScreenState extends State<NewCarouselScreen> {
  late Future<SchemaResponse> _form;
  Map<String, Object?> _values = {};
  Map<String, String> _errors = {};
  String _ratio = '4:5';
  bool _seeded = false;

  String? _jobId;
  String? _carouselId;
  JobResponse? _job;
  Timer? _poll;
  String? _error;

  @override
  void initState() {
    super.initState();
    _form = widget.api.client.getCarouselForm(widget.project.id);
  }

  @override
  void dispose() {
    _poll?.cancel();
    super.dispose();
  }

  Future<void> _generate(Strings s, SchemaForm form) async {
    final errors = DynamicForm.validate(form, _values, s);
    if (errors.isNotEmpty) {
      setState(() => _errors = errors);
      return;
    }
    setState(() {
      _errors = {};
      _error = null;
    });
    try {
      final res = await widget.api.client.generateCarousel(
        widget.project.id,
        GenerateCarouselRequest(inputs: _values, ratio: _ratio),
      );
      setState(() {
        _jobId = res.jobId;
        _carouselId = res.carouselId;
      });
      _startPolling();
    } on ApiException catch (e) {
      setState(() => _error = s.error(e.code, e.message));
    }
  }

  /// The backend runs generation on a worker, so progress is polled (§13).
  void _startPolling() {
    _poll?.cancel();
    _poll = Timer.periodic(const Duration(milliseconds: 1500), (_) async {
      final jobId = _jobId;
      if (jobId == null) return;
      try {
        final job = await widget.api.client.getJob(jobId);
        if (!mounted) return;
        setState(() => _job = job);
        if (job.status == 'completed') {
          _poll?.cancel();
          final id = _carouselId;
          if (id == null) return;
          await Navigator.of(context).pushReplacement(
            MaterialPageRoute(
              builder: (_) => CarouselDetailScreen(api: widget.api, carouselId: id),
            ),
          );
        } else if (job.status == 'failed') {
          _poll?.cancel();
        }
      } on ApiException {
        // Transient failures are fine; the next tick retries.
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final s = Strings(Localizations.localeOf(context).languageCode);
    return Scaffold(
      appBar: AppBar(title: Text(s.newCarousel)),
      body: SafeArea(
        child: _jobId != null
            ? _Progress(job: _job, strings: s)
            : FutureBuilder<SchemaResponse>(
                future: _form,
                builder: (context, snapshot) {
                  if (snapshot.connectionState == ConnectionState.waiting) {
                    return const Center(child: CircularProgressIndicator());
                  }
                  if (!snapshot.hasData) {
                    return EmptyState(title: s.somethingWentWrong);
                  }
                  final form = snapshot.data!.form;
                  if (!_seeded) {
                    _values = DynamicForm.defaults(form);
                    _seeded = true;
                  }
                  return ListView(
                    padding: const EdgeInsets.all(16),
                    children: [
                      DynamicForm(
                        form: form,
                        values: _values,
                        errors: _errors,
                        onChanged: (name, value) => setState(() => _values[name] = value),
                      ),
                      Text(s.canvasFormat,
                          style: const TextStyle(fontWeight: FontWeight.w600)),
                      const SizedBox(height: 8),
                      RadioGroup<String>(
                        groupValue: _ratio,
                        onChanged: (v) => setState(() => _ratio = v ?? '4:5'),
                        child: Column(children: [
                          for (final (ratio, size, recommended) in _ratios)
                            RadioListTile<String>(
                              contentPadding: EdgeInsets.zero,
                              value: ratio,
                              activeColor: AppColors.ink,
                              title: Row(children: [
                                Text(ratio,
                                    style: const TextStyle(fontWeight: FontWeight.w600)),
                                const SizedBox(width: 8),
                                Text(size,
                                    style: const TextStyle(
                                        color: AppColors.body, fontSize: 13)),
                                if (recommended) ...[
                                  const Spacer(),
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                        horizontal: 8, vertical: 2),
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFECFDF5),
                                      borderRadius: BorderRadius.circular(999),
                                    ),
                                    child: Text(s.recommended,
                                        style: const TextStyle(
                                            color: Color(0xFF047857), fontSize: 11)),
                                  ),
                                ],
                              ]),
                            ),
                        ]),
                      ),
                      const SizedBox(height: 20),
                      FilledButton(
                        onPressed: () => _generate(s, form),
                        child: Text(s.generate),
                      ),
                      ErrorNote(_error),
                    ],
                  );
                },
              ),
      ),
    );
  }
}

/// Mirrors the web app's progress list: the backend owns the pipeline shape and
/// reports a stable status per step; the wording is chosen on the client.
class _Progress extends StatelessWidget {
  const _Progress({required this.job, required this.strings});

  final JobResponse? job;
  final Strings strings;

  @override
  Widget build(BuildContext context) {
    final steps = job?.steps ?? const <Map<String, dynamic>>[];
    final failed = job?.status == 'failed';
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text(failed ? strings.somethingWentWrong : strings.generating,
            style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w700)),
        const SizedBox(height: 16),
        ClipRRect(
          borderRadius: BorderRadius.circular(999),
          child: LinearProgressIndicator(
            value: (job?.progress ?? 0) / 100,
            minHeight: 6,
            backgroundColor: AppColors.line,
            color: failed ? const Color(0xFFB91C1C) : AppColors.ink,
          ),
        ),
        const SizedBox(height: 24),
        for (final step in steps)
          Padding(
            padding: const EdgeInsets.only(bottom: 14),
            child: Row(children: [
              SizedBox(
                width: 22,
                child: switch (step['state']) {
                  'done' => const Icon(Icons.check, size: 16, color: Color(0xFF047857)),
                  'active' => const SizedBox(
                      height: 14, width: 14, child: CircularProgressIndicator(strokeWidth: 2)),
                  _ => const Icon(Icons.circle, size: 7, color: AppColors.muted),
                },
              ),
              const SizedBox(width: 10),
              Text(
                strings.step('${step['status']}'),
                style: TextStyle(
                  color: step['state'] == 'pending' ? AppColors.muted : AppColors.ink,
                  fontWeight:
                      step['state'] == 'active' ? FontWeight.w600 : FontWeight.w400,
                ),
              ),
            ]),
          ),
        if (failed) ErrorNote(job?.errorMessage ?? strings.somethingWentWrong),
      ],
    );
  }
}
