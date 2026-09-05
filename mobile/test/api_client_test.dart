import 'dart:convert';

import 'package:carousel_studio/api/generated.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  group('session handling', () {
    // The backend authenticates with an httpOnly cookie. A native app has no
    // browser cookie jar, so the client has to capture Set-Cookie itself —
    // this is the whole auth mechanism on mobile and worth pinning down.
    test('captures the session cookie from a sign-in response', () async {
      final client = ApiClient(
        baseUrl: 'https://api.test',
        httpClient: MockClient((_) async => http.Response(
              jsonEncode({'id': 'u1', 'email': 'a@b.c', 'name': 'A', 'avatar_url': '', 'provider': 'dev', 'created_at': '2026-01-01T00:00:00Z'}),
              200,
              headers: {
                'content-type': 'application/json',
                'set-cookie': 'tks_session=abc123; Path=/; HttpOnly; SameSite=Lax',
              },
            )),
      );

      await client.devLogin(DevLoginRequest(email: 'a@b.c'));

      expect(client.session, 'tks_session=abc123');
    });

    test('sends the stored session on later requests', () async {
      String? sentCookie;
      final client = ApiClient(
        baseUrl: 'https://api.test',
        httpClient: MockClient((request) async {
          sentCookie = request.headers['Cookie'];
          return http.Response(jsonEncode({'projects': []}), 200,
              headers: {'content-type': 'application/json'});
        }),
      )..session = 'tks_session=abc123';

      await client.listProjects();

      expect(sentCookie, 'tks_session=abc123');
    });

    test('an empty session cookie clears the session rather than storing ""', () async {
      final client = ApiClient(
        baseUrl: 'https://api.test',
        httpClient: MockClient((_) async => http.Response('', 204,
            headers: {'set-cookie': 'tks_session=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT'})),
      )..session = 'tks_session=abc123';

      await client.logout();

      expect(client.session, isNull);
    });
  });

  group('error handling', () {
    test('surfaces the backend error code so the UI can localise it', () async {
      final client = ApiClient(
        baseUrl: 'https://api.test',
        httpClient: MockClient((_) async => http.Response(
              jsonEncode({
                'error': {'code': 'rate_limited', 'message': 'Bạn đã đạt giới hạn hôm nay.'}
              }),
              429,
              headers: {'content-type': 'application/json'},
            )),
      );

      expect(
        () => client.listProjects(),
        throwsA(isA<ApiException>()
            .having((e) => e.code, 'code', 'rate_limited')
            .having((e) => e.statusCode, 'statusCode', 429)),
      );
    });
  });

  group('generated models', () {
    test('decode a design payload without losing element data', () {
      final view = CarouselDesignView.fromJson({
        'version': 3,
        'design': {
          'version': '1.0',
          'canvas': {'width': 1080, 'height': 1350, 'ratio': '4:5', 'safe_area': {'x': 80, 'y': 80, 'width': 920, 'height': 1190}},
          'palette': {'primary': '#111', 'secondary': '#222', 'accent': '#333', 'background': '#444', 'text': '#555'},
          'slides': [
            {
              'id': 'slide_1',
              'index': 1,
              'role': 'hook',
              'background': '#0F172A',
              'elements': [
                {'id': 'bg', 'type': 'image', 'z': 0, 'x': 0, 'y': 0, 'width': 1080, 'height': 1350, 'opacity': 1, 'overlay': {'color': '#000000', 'opacity': 0.45}},
                {'id': 'hook', 'type': 'text', 'z': 1, 'x': 80, 'y': 400, 'width': 920, 'height': 200, 'opacity': 1, 'content': '5 dấu hiệu bé đang thiếu ngủ', 'role': 'hook', 'style': {'fontFamily': 'Inter', 'fontSize': 76, 'fontWeight': 700, 'color': '#FFFFFF', 'textAlign': 'center', 'lineHeight': 1.15}},
              ],
            }
          ],
        },
      });

      expect(view.version, 3);
      expect(view.design.slides, hasLength(1));
      final elements = view.design.slides.first.elements;
      expect(elements.first.overlay?.opacity, closeTo(0.45, 0.001));
      expect(elements.last.content, '5 dấu hiệu bé đang thiếu ngủ');
      expect(elements.last.style?.fontSize, 76);
    });

    test('tolerates fields the backend omits', () {
      final caption = CarouselCaption.fromJson({
        'caption': 'hi',
        'hashtags': ['#fyp'],
        'edited': false,
        'updated_at': '2026-01-01T00:00:00Z',
      });
      expect(caption.generatedAt, isNull);
      expect(caption.hashtags, ['#fyp']);
    });
  });
}
