import 'package:flutter/material.dart';

import '../api/generated.dart';
import '../i18n/strings.dart';
import 'common.dart';

/// Edits one text element: its content, size and alignment.
///
/// The web editor exposes more (font family, colour, weight). On a phone the
/// useful subset is the copy itself plus the two knobs that fix a bad line
/// break — anything finer belongs on a larger screen.
class TextEditorSheet extends StatefulWidget {
  const TextEditorSheet({super.key, required this.element, required this.strings});

  final DesignElement element;
  final Strings strings;

  @override
  State<TextEditorSheet> createState() => _TextEditorSheetState();
}

class _TextEditorSheetState extends State<TextEditorSheet> {
  late final TextEditingController _controller =
      TextEditingController(text: widget.element.content ?? '');
  late double _fontSize = widget.element.style?.fontSize ?? 36;
  late String _align = widget.element.style?.textAlign ?? 'left';

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  DesignElement _result() {
    final style = widget.element.style;
    return widget.element.copyWith(
      content: _controller.text,
      style: style?.copyWith(fontSize: _fontSize, textAlign: _align),
    );
  }

  @override
  Widget build(BuildContext context) {
    final s = widget.strings;
    final limit = const {'hook', 'headline', 'cta'}.contains(widget.element.role) ? 90 : 220;
    final length = _controller.text.characters.length;

    return Padding(
      padding: EdgeInsets.only(
        left: 16,
        right: 16,
        top: 16,
        bottom: MediaQuery.viewInsetsOf(context).bottom + 16,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(children: [
            Text(widget.element.role ?? 'text',
                style: const TextStyle(
                    fontWeight: FontWeight.w600, color: AppColors.body, fontSize: 12)),
            const Spacer(),
            Text('$length/$limit',
                style: TextStyle(
                    fontSize: 12,
                    color: length > limit ? const Color(0xFFB91C1C) : AppColors.muted)),
          ]),
          const SizedBox(height: 8),
          TextField(
            controller: _controller,
            autofocus: true,
            maxLines: 4,
            onChanged: (_) => setState(() {}),
          ),
          if (length > limit) ...[
            const SizedBox(height: 6),
            Text(s.tooLong, style: const TextStyle(color: Color(0xFFB45309), fontSize: 12)),
          ],
          const SizedBox(height: 16),
          Row(children: [
            Text(s.fontSize, style: const TextStyle(fontWeight: FontWeight.w600)),
            Expanded(
              child: Slider(
                value: _fontSize.clamp(18, 140),
                min: 18,
                max: 140,
                divisions: 61,
                activeColor: AppColors.ink,
                label: '${_fontSize.round()}',
                onChanged: (v) => setState(() => _fontSize = v),
              ),
            ),
            SizedBox(
                width: 34,
                child: Text('${_fontSize.round()}',
                    textAlign: TextAlign.end,
                    style: const TextStyle(fontWeight: FontWeight.w600))),
          ]),
          const SizedBox(height: 8),
          SegmentedButton<String>(
            segments: [
              ButtonSegment(value: 'left', icon: const Icon(Icons.format_align_left), label: Text(s.alignLeft)),
              ButtonSegment(value: 'center', icon: const Icon(Icons.format_align_center), label: Text(s.alignCenter)),
              ButtonSegment(value: 'right', icon: const Icon(Icons.format_align_right), label: Text(s.alignRight)),
            ],
            selected: {_align},
            showSelectedIcon: false,
            onSelectionChanged: (v) => setState(() => _align = v.first),
          ),
          const SizedBox(height: 20),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(_result()),
            child: Text(s.done),
          ),
        ],
      ),
    );
  }
}
