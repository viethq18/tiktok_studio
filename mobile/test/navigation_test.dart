import 'package:carousel_studio/api/generated.dart';
import 'package:carousel_studio/api/session.dart';
import 'package:carousel_studio/screens/carousel_detail.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

/// Getting stuck on a screen is the worst kind of bug: nothing crashes, the
/// user just cannot leave. These tests pin the way back.
void main() {
  Api offlineApi() => Api.forTesting(ApiClient(
        baseUrl: 'https://api.test',
        // Every call fails, which is also the worst case for the UI: the screen
        // must still render its chrome, including the way back.
        httpClient: MockClient((_) async => http.Response('{}', 500)),
      ));

  testWidgets('the carousel screen offers a back button when pushed',
      (tester) async {
    await tester.pumpWidget(MaterialApp(
      home: Builder(
        builder: (context) => Scaffold(
          body: TextButton(
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => CarouselDetailScreen(
                    api: offlineApi(), carouselId: 'c1'),
              ),
            ),
            child: const Text('open'),
          ),
        ),
      ),
    ));

    await tester.tap(find.text('open'));
    await tester.pumpAndSettle();

    expect(find.byType(BackButton), findsOneWidget,
        reason: 'without this the screen is a dead end');
  });

  testWidgets('the back button actually pops back to the previous screen',
      (tester) async {
    await tester.pumpWidget(MaterialApp(
      home: Builder(
        builder: (context) => Scaffold(
          body: TextButton(
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => CarouselDetailScreen(
                    api: offlineApi(), carouselId: 'c1'),
              ),
            ),
            child: const Text('projects screen'),
          ),
        ),
      ),
    ));

    await tester.tap(find.text('projects screen'));
    await tester.pumpAndSettle();
    expect(find.text('projects screen'), findsNothing);

    await tester.tap(find.byType(BackButton));
    await tester.pumpAndSettle();

    expect(find.text('projects screen'), findsOneWidget);
  });

  testWidgets('a failed load still renders the app bar rather than hanging',
      (tester) async {
    await tester.pumpWidget(MaterialApp(
      home: CarouselDetailScreen(api: offlineApi(), carouselId: 'c1'),
    ));
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.byType(AppBar), findsOneWidget);
  });
}
