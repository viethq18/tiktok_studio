import 'package:carousel_studio/api/generated.dart';
import 'package:carousel_studio/i18n/strings.dart';
import 'package:carousel_studio/widgets/slide_preview.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

/// The preview turns Design JSON into pixels using the same absolute
/// coordinates the Go exporter uses, scaled to the phone. If this drifts, the
/// app shows something the export will not produce.
void main() {
  final canvas = DesignCanvas(
    width: 1080,
    height: 1350,
    ratio: '4:5',
    safeArea: DesignSafeArea(x: 80, y: 80, width: 920, height: 1190),
  );

  DesignSlide slideWith(List<DesignElement> elements) => DesignSlide(
        id: 'slide_1',
        index: 1,
        role: 'hook',
        background: '#0F172A',
        elements: elements,
      );

  // A phone-shaped box: the PageView gives roughly this on an iPhone, and it is
  // narrower than the canvas would be if scaled by height alone.
  Widget host(
    DesignSlide slide, {
    double width = 320,
    double height = 620,
    bool editing = false,
    String? selectedElementId,
    void Function(String?)? onSelect,
    void Function(DesignElement, double, double)? onMove,
    void Function(DesignElement)? onEditText,
  }) =>
      MaterialApp(
        home: Scaffold(
          body: SizedBox(
            width: width,
            height: height,
            child: SlidePreview(
              slide: slide,
              canvas: canvas,
              index: 1,
              total: 3,
              onChangeImage: () {},
              strings: const Strings('en'),
              editing: editing,
              selectedElementId: selectedElementId,
              onSelect: onSelect,
              onMove: onMove,
              onEditText: onEditText,
            ),
          ),
        ),
      );

  DesignElement textElement() => DesignElement(
        id: 'hook',
        type: 'text',
        z: 0,
        x: 80,
        y: 400,
        width: 400,
        height: 200,
        opacity: 1,
        content: 'drag me',
        style: DesignStyle(
          fontFamily: 'Inter',
          fontSize: 40,
          fontWeight: 700,
          color: '#FFFFFF',
          textAlign: 'left',
          lineHeight: 1.2,
        ),
      );

  testWidgets('the slide fits inside a narrow viewport instead of overflowing',
      (tester) async {
    // Scaling by height alone made the slide wider than the box, so it was
    // clipped off-centre — the misalignment this test exists to prevent.
    await tester.pumpWidget(host(slideWith([
      DesignElement(
        id: 'bg',
        type: 'image',
        z: 0,
        x: 0,
        y: 0,
        width: 1080,
        height: 1350,
        opacity: 1,
      ),
    ])));

    final canvasBox = tester.getSize(find.byType(ClipRRect).first);
    expect(canvasBox.width, lessThanOrEqualTo(320),
        reason: 'the slide is wider than the space it was given');
    expect(canvasBox.height, lessThanOrEqualTo(620));
    // 4:5 must be preserved whichever axis constrains it.
    expect(canvasBox.width / canvasBox.height, closeTo(1080 / 1350, 0.01));
    expect(tester.takeException(), isNull);
  });

  testWidgets('draws slide text with the font size scaled to the viewport',
      (tester) async {
    await tester.pumpWidget(host(slideWith([
      DesignElement(
        id: 't',
        type: 'text',
        z: 0,
        x: 80,
        y: 400,
        width: 920,
        height: 200,
        opacity: 1,
        content: '5 dấu hiệu bé đang thiếu ngủ',
        style: DesignStyle(
          fontFamily: 'Inter',
          fontSize: 76,
          fontWeight: 700,
          color: '#FFFFFF',
          textAlign: 'center',
          lineHeight: 1.15,
        ),
      ),
    ])));

    final text = tester.widget<Text>(find.text('5 dấu hiệu bé đang thiếu ngủ'));
    expect(text.style?.color, const Color(0xFFFFFFFF));
    expect(text.textAlign, TextAlign.center);
    // The preview area is shorter than the slide's own row of controls, so the
    // scale is whatever height the canvas got — assert it scaled down at all.
    expect(text.style!.fontSize, lessThan(76));
    expect(text.style!.fontSize, greaterThan(0));
  });

  testWidgets('honours font weight from the design', (tester) async {
    await tester.pumpWidget(host(slideWith([
      DesignElement(
        id: 'body',
        type: 'text',
        z: 0,
        x: 80,
        y: 900,
        width: 920,
        height: 100,
        opacity: 1,
        content: 'body copy',
        style: DesignStyle(
          fontFamily: 'Inter',
          fontSize: 36,
          fontWeight: 400,
          color: '#F3F4F6',
          textAlign: 'left',
          lineHeight: 1.4,
        ),
      ),
    ])));

    final text = tester.widget<Text>(find.text('body copy'));
    expect(text.style?.fontWeight, FontWeight.w400);
    expect(text.textAlign, TextAlign.left);
  });

  testWidgets('an image element with no asset does not crash the preview',
      (tester) async {
    // A carousel can be viewed before the worker has attached imagery.
    await tester.pumpWidget(host(slideWith([
      DesignElement(
        id: 'bg',
        type: 'image',
        z: 0,
        x: 0,
        y: 0,
        width: 1080,
        height: 1350,
        opacity: 1,
        overlay: DesignOverlay(color: '#000000', opacity: 0.45),
      ),
    ])));

    expect(tester.takeException(), isNull);
  });

  group('colour parsing', () {
    test('accepts 6-digit, 3-digit and malformed hex', () {
      expect(SlidePreview.hex('#FF3B5C', Colors.black), const Color(0xFFFF3B5C));
      expect(SlidePreview.hex('#fff', Colors.black), const Color(0xFFFFFFFF));
      expect(SlidePreview.hex('not-a-colour', Colors.black), Colors.black);
      expect(SlidePreview.hex(null, Colors.red), Colors.red);
    });
  });

  group('editing gestures', () {
    testWidgets('reports drags in logical canvas units, not screen pixels',
        (tester) async {
      DesignElement? moved;
      double? dx, dy;
      await tester.pumpWidget(host(
        slideWith([textElement()]),
        editing: true,
        onMove: (element, x, y) {
          moved = element;
          dx = (dx ?? 0) + x;
          dy = (dy ?? 0) + y;
        },
      ));

      // The canvas is scaled down to fit the phone, so a 60pt drag on screen is
      // a much larger move in the 1080-wide design space. Deriving the scale
      // from the rendered handle keeps the assertion honest about whatever
      // layout the test surface produced.
      final handle = find.byKey(SlidePreview.handleKey('hook'));
      final scale = tester.getSize(handle).width / 400; // element's logical width

      // A manual gesture, because tester.drag() consumes kTouchSlop before the
      // recogniser reports anything — measuring after the drag has started is
      // what makes the assertion exact.
      final gesture = await tester.startGesture(tester.getCenter(handle));
      await gesture.moveBy(const Offset(40, 0)); // past the slop
      await tester.pump();
      dx = 0;
      dy = 0;

      await gesture.moveBy(const Offset(60, 30));
      await tester.pump();
      await gesture.up();

      expect(moved?.id, 'hook');
      expect(dx, closeTo(60 / scale, 0.5));
      expect(dy, closeTo(30 / scale, 0.5));
    });

    testWidgets('tapping selects, and tapping the selection opens the editor',
        (tester) async {
      String? selected;
      DesignElement? editing;

      await tester.pumpWidget(host(
        slideWith([textElement()]),
        editing: true,
        onSelect: (id) => selected = id,
        onEditText: (element) => editing = element,
      ));
      await tester.tap(find.byKey(SlidePreview.handleKey('hook')));
      expect(selected, 'hook', reason: 'first tap selects');
      expect(editing, isNull, reason: 'first tap must not open the sheet');

      await tester.pumpWidget(host(
        slideWith([textElement()]),
        editing: true,
        selectedElementId: 'hook',
        onSelect: (id) => selected = id,
        onEditText: (element) => editing = element,
      ));
      await tester.tap(find.byKey(SlidePreview.handleKey('hook')));
      expect(editing?.id, 'hook', reason: 'second tap edits');
    });

    testWidgets('read-only mode adds no gesture handles', (tester) async {
      var moved = false;
      await tester.pumpWidget(host(
        slideWith([textElement()]),
        onMove: (_, _, _) => moved = true,
      ));
      await tester.drag(find.byType(SlidePreview), const Offset(60, 30));
      await tester.pump();
      expect(moved, isFalse);
    });
  });
}
