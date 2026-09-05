import 'package:carousel_studio/api/generated.dart';
import 'package:carousel_studio/api/session.dart';
import 'package:carousel_studio/state/design_editor.dart';
import 'package:flutter_test/flutter_test.dart';

/// [width] lets a test choose between a full-width element (which cannot move
/// sideways at all) and a narrower one (which can).
Design _design({double width = 920}) => Design(
      version: '1.0',
      canvas: DesignCanvas(
        width: 1080,
        height: 1350,
        ratio: '4:5',
        safeArea: DesignSafeArea(x: 80, y: 80, width: 920, height: 1190),
      ),
      palette: DesignPalette(
          primary: '#111', secondary: '#222', accent: '#333',
          background: '#444', text: '#555'),
      slides: [
        DesignSlide(
          id: 'slide_1',
          index: 1,
          role: 'hook',
          background: '#000',
          elements: [
            DesignElement(
              id: 'hook',
              type: 'text',
              z: 0,
              x: 80,
              y: 400,
              width: width,
              height: 200,
              opacity: 1,
              content: 'hello',
            ),
          ],
        ),
      ],
    );

DesignEditor _editor({double width = 920}) {
  final editor = DesignEditor(
    api: Api.forTesting(ApiClient(baseUrl: 'https://api.test')),
    carouselId: 'c1',
  );
  editor.design = _design(width: width);
  return editor;
}

DesignElement _hook(DesignEditor editor) =>
    editor.design!.slides.first.elements.first;

void main() {
  group('moving text', () {
    test('applies a delta in logical canvas units', () {
      final editor = _editor(width: 400);
      editor.moveBy('slide_1', _hook(editor), 40, -60);
      expect(_hook(editor).x, 120);
      expect(_hook(editor).y, 340);
    });

    test('cannot be dragged above or left of the safe area', () {
      // The backend clamps to the safe area too (§123); doing it here as well
      // means the preview never shows a position the server would move.
      final editor = _editor(width: 400);
      editor.moveBy('slide_1', _hook(editor), -500, -500);
      expect(_hook(editor).x, 80);
      expect(_hook(editor).y, 80);
    });

    test('cannot be dragged past the bottom of the safe area', () {
      final editor = _editor(width: 400);
      editor.moveBy('slide_1', _hook(editor), 0, 5000);
      // safe bottom (80 + 1190) minus the element height (200).
      expect(_hook(editor).y, 1070);
    });

    test('a full-width element cannot be nudged sideways out of the safe area', () {
      final editor = _editor();
      editor.moveBy('slide_1', _hook(editor), 5000, 0);
      expect(_hook(editor).x, 80);
    });
  });

  group('editing content', () {
    test('replaces only the element that changed', () {
      final editor = _editor();
      editor.updateElement(
          'slide_1', _hook(editor).copyWith(content: 'changed'));
      expect(_hook(editor).content, 'changed');
      expect(editor.design!.slides.first.elements, hasLength(1));
      expect(_hook(editor).x, 80, reason: 'geometry must survive a text edit');
    });

    test('selection is tracked so the canvas can highlight it', () {
      final editor = _editor();
      expect(editor.selectedElementId, isNull);
      editor.select('hook');
      expect(editor.selectedElementId, 'hook');
      editor.select(null);
      expect(editor.selectedElementId, isNull);
    });
  });
}
