import 'package:flutter/material.dart';

import '../api/generated.dart';
import '../api/session.dart';
import '../i18n/strings.dart';
import '../widgets/common.dart';

/// Onboarding. The content language is an explicit choice with no default —
/// a wrong default would write every carousel in the project in the wrong
/// language and only surface after the first generation.
class NewProjectScreen extends StatefulWidget {
  const NewProjectScreen({super.key, required this.api});

  final Api api;

  @override
  State<NewProjectScreen> createState() => _NewProjectScreenState();
}

class _NewProjectScreenState extends State<NewProjectScreen> {
  final _niche = TextEditingController();
  late Future<List<Language>> _languages;
  String? _language;
  bool _pending = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _languages = widget.api.client.getRegistry().then((r) => r.languages);
  }

  @override
  void dispose() {
    _niche.dispose();
    super.dispose();
  }

  Future<void> _create(Strings s) async {
    final language = _language;
    if (language == null) {
      setState(() => _error = s.pickLanguage);
      return;
    }
    setState(() {
      _pending = true;
      _error = null;
    });
    try {
      await widget.api.client.createProject(
        CreateProjectRequest(niche: _niche.text.trim(), language: language),
      );
      if (mounted) Navigator.of(context).pop(true);
    } on ApiException catch (e) {
      setState(() => _error = s.error(e.code, e.message));
    } catch (_) {
      setState(() => _error = s.somethingWentWrong);
    } finally {
      if (mounted) setState(() => _pending = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = Strings(Localizations.localeOf(context).languageCode);
    return Scaffold(
      appBar: AppBar(title: Text(s.newProject)),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(s.channelQuestion,
                style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w700)),
            const SizedBox(height: 16),
            TextField(
              controller: _niche,
              maxLines: 4,
              decoration: InputDecoration(hintText: s.channelHint),
              onChanged: (_) => setState(() {}),
            ),
            const SizedBox(height: 24),
            Text(s.contentLanguage,
                style: const TextStyle(fontWeight: FontWeight.w600)),
            const SizedBox(height: 4),
            Text(s.contentLanguageHint,
                style: const TextStyle(color: AppColors.body, fontSize: 13)),
            const SizedBox(height: 12),
            FutureBuilder<List<Language>>(
              future: _languages,
              builder: (context, snapshot) {
                final languages = snapshot.data ?? const <Language>[];
                if (languages.isEmpty) {
                  return const Padding(
                    padding: EdgeInsets.symmetric(vertical: 8),
                    child: SizedBox(
                      height: 24,
                      width: 24,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  );
                }
                return Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    for (final language in languages)
                      ChoiceChip(
                        label: Text(language.nativeName),
                        selected: _language == language.code,
                        showCheckmark: false,
                        onSelected: (_) => setState(() {
                          _language = language.code;
                          _error = null;
                        }),
                      ),
                  ],
                );
              },
            ),
            const SizedBox(height: 28),
            FilledButton(
              onPressed: _pending || _niche.text.trim().length < 8
                  ? null
                  : () => _create(s),
              child: Text(_pending ? s.analysing : s.create),
            ),
            ErrorNote(_error),
          ],
        ),
      ),
    );
  }
}
