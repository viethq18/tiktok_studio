import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'generated.dart';

/// Wraps the generated client with the two things it cannot know about:
/// where the backend lives, and how to persist the session across launches.
///
/// The backend authenticates with an httpOnly cookie. A native app has no
/// browser cookie jar, so the value is captured from Set-Cookie by the
/// generated client and stored here.
class Api {
  Api._(this.client);

  static const _sessionKey = 'tks_session';

  /// Override at build time:
  /// `flutter run --dart-define=API_URL=https://api.example.com`
  static const baseUrl = String.fromEnvironment(
    'API_URL',
    defaultValue: 'http://localhost:8080',
  );

  final ApiClient client;

  /// Builds an Api around a client directly, without touching storage.
  @visibleForTesting
  factory Api.forTesting(ApiClient client) => Api._(client);

  static Future<Api> load() async {
    final prefs = await SharedPreferences.getInstance();
    final api = Api._(ApiClient(baseUrl: baseUrl));
    api.client.session = prefs.getString(_sessionKey);
    return api;
  }

  Future<void> persistSession() async {
    final prefs = await SharedPreferences.getInstance();
    final session = client.session;
    if (session == null) {
      await prefs.remove(_sessionKey);
    } else {
      await prefs.setString(_sessionKey, session);
    }
  }

  Future<void> signOut() async {
    try {
      await client.logout();
    } on ApiException {
      // A dead session is still a successful sign-out from the user's side.
    }
    client.session = null;
    await persistSession();
  }
}
