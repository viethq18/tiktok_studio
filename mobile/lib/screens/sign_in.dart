import 'package:flutter/material.dart';

import '../api/generated.dart';
import '../api/session.dart';
import '../i18n/strings.dart';
import '../widgets/common.dart';

/// Dev-login only for now.
///
/// Google sign-in on the web is a browser redirect; on iOS it needs
/// ASWebAuthenticationSession plus a deep link back into the app, which in turn
/// needs the backend to accept a mobile redirect URI. Until that exists,
/// offering a Google button that cannot complete would be worse than omitting it.
class SignInScreen extends StatefulWidget {
  const SignInScreen({super.key, required this.api, required this.onSignedIn});

  final Api api;
  final VoidCallback onSignedIn;

  @override
  State<SignInScreen> createState() => _SignInScreenState();
}

class _SignInScreenState extends State<SignInScreen> {
  final _email = TextEditingController();
  bool _pending = false;
  String? _error;

  @override
  void dispose() {
    _email.dispose();
    super.dispose();
  }

  Future<void> _submit(Strings s) async {
    setState(() {
      _pending = true;
      _error = null;
    });
    try {
      await widget.api.client.devLogin(DevLoginRequest(email: _email.text.trim()));
      await widget.api.persistSession();
      widget.onSignedIn();
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
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: AppCard(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(s.appName,
                      style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w700)),
                  const SizedBox(height: 6),
                  Text(s.signInSubtitle, style: const TextStyle(color: AppColors.body)),
                  const SizedBox(height: 24),
                  Text(s.email,
                      style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
                  const SizedBox(height: 8),
                  TextField(
                    controller: _email,
                    keyboardType: TextInputType.emailAddress,
                    autocorrect: false,
                    decoration: InputDecoration(hintText: s.emailHint),
                    onSubmitted: (_) => _submit(s),
                  ),
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: _pending ? null : () => _submit(s),
                    child: Text(_pending ? s.signingIn : s.signInWithEmail),
                  ),
                  ErrorNote(_error),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
