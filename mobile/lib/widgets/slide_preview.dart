import 'dart:math' as math;

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../api/generated.dart';
import '../i18n/strings.dart';
import 'common.dart';

/// Renders a slide from Design JSON.
///
/// This is a faithful preview, not an editor: the same absolute coordinates the
/// exporter uses, scaled to the widget. It intentionally cannot be dragged.
class SlidePreview extends StatelessWidget {
  const SlidePreview({
    super.key,
    required this.slide,
    required this.canvas,
    required this.index,
    required this.total,
    required this.onChangeImage,
    required this.strings,
    this.editing = false,
    this.selectedElementId,
    this.onSelect,
    this.onMove,
    this.onEditText,
  });

  final DesignSlide slide;
  final DesignCanvas canvas;
  final int index;
  final int total;
  final VoidCallback onChangeImage;
  final Strings strings;

  /// When true, text elements can be selected, dragged and opened for editing.
  final bool editing;
  final String? selectedElementId;
  final void Function(String? elementId)? onSelect;

  /// Deltas are reported in logical canvas units, not screen pixels.
  final void Function(DesignElement element, double dx, double dy)? onMove;
  final void Function(DesignElement element)? onEditText;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Expanded(
          child: LayoutBuilder(
            builder: (context, constraints) {
              // Coordinates are stored against the logical canvas (§99), so the
              // preview scales the whole slide rather than laying anything out.
              // The scale has to respect BOTH axes: taking height alone made the
              // slide wider than its box on a phone, and the overflow was
              // clipped off-centre.
              final scale = math.min(
                constraints.maxWidth / canvas.width,
                constraints.maxHeight / canvas.height,
              );
              return Center(
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: SizedBox(
                    width: canvas.width * scale,
                    height: canvas.height * scale,
                    child: Stack(
                      children: [
                        Positioned.fill(
                          child: ColoredBox(color: hex(slide.background, Colors.black)),
                        ),
                        for (final element in slide.elements)
                          _element(element, scale),
                        if (editing)
                          // Tapping empty canvas clears the selection.
                          Positioned.fill(
                            child: GestureDetector(
                              behavior: HitTestBehavior.translucent,
                              onTap: () => onSelect?.call(null),
                            ),
                          ),
                        if (editing)
                          for (final element in slide.elements)
                            if (element.type == 'text') _handle(element, scale),
                      ],
                    ),
                  ),
                ),
              );
            },
          ),
        ),
        const SizedBox(height: 12),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text('$index / $total',
                style: const TextStyle(color: AppColors.muted, fontSize: 12)),
            const SizedBox(width: 12),
            OutlinedButton.icon(
              style: OutlinedButton.styleFrom(minimumSize: const Size(0, 36)),
              onPressed: onChangeImage,
              icon: const Icon(Icons.image_outlined, size: 18),
              label: Text(strings.images),
            ),
          ],
        ),
      ],
    );
  }

  Widget _element(DesignElement e, double scale) {
    switch (e.type) {
      case 'image':
        final url = e.url;
        return Positioned(
          left: e.x * scale,
          top: e.y * scale,
          width: e.width * scale,
          height: e.height * scale,
          child: Stack(fit: StackFit.expand, children: [
            if (url != null && url.isNotEmpty)
              CachedNetworkImage(
                imageUrl: url,
                fit: e.fit == 'contain' ? BoxFit.contain : BoxFit.cover,
              ),
            if (e.overlay != null)
              ColoredBox(
                color: hex(e.overlay!.color, Colors.black)
                    .withValues(alpha: e.overlay!.opacity),
              ),
          ]),
        );
      case 'shape':
        return Positioned(
          left: e.x * scale,
          top: e.y * scale,
          width: e.width * scale,
          height: e.height * scale,
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: hex(e.fill, Colors.white),
              borderRadius: BorderRadius.circular((e.radius ?? 0) * scale),
            ),
          ),
        );
      case 'text':
        final style = e.style;
        if (style == null) return const SizedBox.shrink();
        return Positioned(
          left: e.x * scale,
          top: e.y * scale,
          width: e.width * scale,
          child: Text(
            e.content ?? '',
            textAlign: switch (style.textAlign) {
              'center' => TextAlign.center,
              'right' => TextAlign.right,
              _ => TextAlign.left,
            },
            style: TextStyle(
              fontSize: style.fontSize * scale,
              height: style.lineHeight,
              fontWeight:
                  style.fontWeight >= 700 ? FontWeight.w700 : FontWeight.w400,
              color: hex(style.color, Colors.white),
            ),
          ),
        );
      default:
        return const SizedBox.shrink();
    }
  }

  /// Identifies the drag handle for a given element, so tests can target the
  /// interaction layer rather than guessing at the widget tree.
  static Key handleKey(String elementId) => ValueKey('slide-handle-$elementId');

  /// An invisible, draggable hit area sitting above one text element.
  ///
  /// Interaction is a separate layer so the rendering path stays identical to
  /// the read-only preview — what you drag is exactly what gets exported.
  Widget _handle(DesignElement element, double scale) {
    final selected = selectedElementId == element.id;
    return Positioned(
      left: element.x * scale,
      top: element.y * scale,
      width: element.width * scale,
      height: element.height * scale,
      child: GestureDetector(
        key: handleKey(element.id),
        behavior: HitTestBehavior.opaque,
        onTap: () => selected
            ? onEditText?.call(element)
            : onSelect?.call(element.id),
        onPanStart: (_) => onSelect?.call(element.id),
        onPanUpdate: (details) => onMove?.call(
          element,
          details.delta.dx / scale,
          details.delta.dy / scale,
        ),
        child: DecoratedBox(
          decoration: BoxDecoration(
            border: Border.all(
              color: selected ? const Color(0xFF3B82F6) : Colors.white24,
              width: selected ? 2 : 1,
            ),
            borderRadius: BorderRadius.circular(4),
          ),
        ),
      ),
    );
  }

  static Color hex(String? value, Color fallback) {
    if (value == null) return fallback;
    var hex = value.replaceFirst('#', '');
    if (hex.length == 3) {
      hex = hex.split('').map((c) => '$c$c').join();
    }
    if (hex.length == 6) hex = 'FF$hex';
    final parsed = int.tryParse(hex, radix: 16);
    return parsed == null ? fallback : Color(parsed);
  }
}

