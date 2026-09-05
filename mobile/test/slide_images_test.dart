import 'dart:convert';

import 'package:carousel_studio/api/generated.dart';
import 'package:carousel_studio/api/session.dart';
import 'package:carousel_studio/screens/slide_images.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

DesignSlide _slide() => DesignSlide(
      id: 'slide_1',
      index: 1,
      role: 'hook',
      background: '#000',
      elements: const [],
    );

String _candidates(int count, String prefix) => jsonEncode({
      'query': 'baby sleep',
      'page': 1,
      'candidates': [
        for (var i = 0; i < count; i++)
          {
            'id': '$prefix$i',
            'provider': 'picsum',
            'thumb_url': 'https://example.test/$prefix$i.jpg',
            'preview_url': 'https://example.test/$prefix$i-big.jpg',
            'width': 1080,
            'height': 1350,
            'photographer_name': 'Photographer $i',
          }
      ],
    });

void main() {
  late List<String> requests;

  Api api({int initial = 12, int perSearch = 12}) {
    requests = [];
    return Api.forTesting(ApiClient(
      baseUrl: 'https://api.test',
      httpClient: MockClient((request) async {
        requests.add('${request.method} ${request.url.path}');
        final search = request.url.path.endsWith('/search');
        return http.Response(
          _candidates(search ? perSearch : initial, search ? 's' : 'i'),
          200,
          headers: {'content-type': 'application/json; charset=utf-8'},
        );
      }),
    ));
  }

  Widget host(Api client) => MaterialApp(
        home: SlideImagesScreen(
          api: client,
          carouselId: 'c1',
          slide: _slide(),
        ),
      );

  testWidgets('the results grid scrolls when there are more than fit',
      (tester) async {
    await tester.pumpWidget(host(api(initial: 24)));
    await tester.pumpAndSettle();

    final grid = find.byType(GridView);
    expect(grid, findsOneWidget);

    // Read the scroll position, not a tile's coordinates: the grid recycles
    // tiles, so the first *built* tile sits near the top at any offset.
    // The last Scrollable is the grid; the first is the search field's
    // single-line EditableText.
    ScrollPosition position() =>
        tester.state<ScrollableState>(find.byType(Scrollable).last).position;

    expect(position().maxScrollExtent, greaterThan(0),
        reason: 'with 24 candidates there must be something to scroll');

    await tester.drag(grid, const Offset(0, -400));
    await tester.pumpAndSettle();

    expect(position().pixels, greaterThan(0),
        reason: 'dragging the grid must move it');
  });

  testWidgets('the search button issues a search request', (tester) async {
    final client = api();
    await tester.pumpWidget(host(client));
    await tester.pumpAndSettle();
    requests.clear();

    await tester.enterText(find.byType(TextField), 'sleeping baby crib');
    await tester.tap(find.byIcon(Icons.search));
    await tester.pumpAndSettle();

    expect(requests.where((r) => r.contains('/images/search')), isNotEmpty,
        reason: 'tapping search must reach the backend');
  });

  testWidgets('submitting the keyword field searches too', (tester) async {
    final client = api();
    await tester.pumpWidget(host(client));
    await tester.pumpAndSettle();
    requests.clear();

    await tester.enterText(find.byType(TextField), 'quiet nursery');
    await tester.testTextInput.receiveAction(TextInputAction.search);
    await tester.pumpAndSettle();

    expect(requests.where((r) => r.contains('/images/search')), isNotEmpty);
  });
}
