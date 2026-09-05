import 'package:carousel_studio/main.dart' as app;
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

/// Runs on a real device or simulator against a live backend.
///
/// Widget tests kept passing while the image picker misbehaved on a phone, so
/// this drives the actual screens the way a person does.
///
///   flutter test integration_test -d `<device>`
void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  void dump(String label) {
    final texts = find
        .byType(Text)
        .evaluate()
        .map((e) => (e.widget as Text).data)
        .whereType<String>()
        .where((t) => t.trim().isNotEmpty)
        .take(14)
        .join(' | ');
    debugPrint('STEP[$label] texts: $texts');
  }

  Future<void> settle(WidgetTester tester, [int seconds = 6]) async {
    final deadline = DateTime.now().add(Duration(seconds: seconds));
    while (DateTime.now().isBefore(deadline)) {
      await tester.pump(const Duration(milliseconds: 200));
    }
  }

  testWidgets('search and scroll inside the slide image picker',
      (tester) async {
    app.main();
    await settle(tester);

    // Sign in if the stored session is gone.
    if (find.byType(TextField).evaluate().isNotEmpty &&
        find.text('Projects').evaluate().isEmpty) {
      await tester.enterText(find.byType(TextField).first, 'mobile@example.com');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await settle(tester);
    }

    dump('after sign-in');

    // Open the first project, then its first carousel.
    await tester.tap(find.byType(InkWell).first);
    await settle(tester);
    dump('after project tap');

    await tester.tap(find.byType(InkWell).first);
    await settle(tester, 8);
    dump('after carousel tap');

    // Into the image picker for the current slide.
    final imagesButton = find.textContaining(RegExp('Hình ảnh|Images'));
    expect(imagesButton, findsWidgets, reason: 'the Images action must be reachable');
    // The PageView keeps neighbouring pages built, so several Images buttons
    // exist; only the centred one is actually on screen.
    await tester.tap(imagesButton.first);
    await settle(tester, 8);
    dump('after images tap');
    debugPrint('TEXTFIELDS=${find.byType(TextField).evaluate().length}');

    // --- search ---
    final keyword = find.byType(TextField).first;
    await tester.enterText(keyword, 'sleeping baby crib');
    await tester.pump();
    await tester.tap(find.byIcon(Icons.search));
    await settle(tester, 8);

    expect(find.byType(GridView), findsOneWidget,
        reason: 'searching must leave a grid of results on screen');

    // --- append more, then scroll ---
    // Count by scroll extent, not by built widgets: a GridView only builds the
    // tiles it can show, so counting InkWells saturates and hides growth.
    ScrollPosition gridPosition() => tester
        .state<ScrollableState>(find.descendant(
            of: find.byType(GridView), matching: find.byType(Scrollable)))
        .position;
    double extent() => gridPosition().maxScrollExtent + gridPosition().viewportDimension;

    final afterSearch = extent();
    expect(afterSearch, greaterThan(0), reason: 'the search returned nothing');

    // Press "find more" repeatedly. This used to grow the grid once and then do
    // nothing, because a single failed call left the loading flag stuck and
    // disabled the button for good.
    final more = find.textContaining(RegExp('Tìm thêm ảnh|Find more images'));
    expect(more, findsWidgets, reason: 'the append action must be labelled distinctly');

    await tester.tap(more.first);
    await settle(tester, 6);
    final afterFirstMore = extent();
    expect(afterFirstMore, greaterThan(afterSearch),
        reason: '"find more" must add results');

    await tester.tap(more.first);
    await settle(tester, 6);
    expect(extent(), greaterThan(afterFirstMore),
        reason: 'the second "find more" must work as well as the first');

    expect(gridPosition().maxScrollExtent, greaterThan(0),
        reason: 'after loading several pages the grid must be scrollable');

    await tester.drag(find.byType(GridView), const Offset(0, -400));
    await settle(tester, 2);
    expect(gridPosition().pixels, greaterThan(0),
        reason: 'dragging the grid must move it');
  });
}
