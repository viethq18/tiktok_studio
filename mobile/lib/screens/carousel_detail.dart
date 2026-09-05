import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';
import 'package:url_launcher/url_launcher.dart';

import '../api/generated.dart';
import '../api/session.dart';
import '../i18n/strings.dart';
import '../widgets/common.dart';
import '../state/design_editor.dart';
import '../widgets/slide_preview.dart';
import '../widgets/text_editor_sheet.dart';
import 'slide_images.dart';

/// Review, adjust and publish.
///
/// There is deliberately no canvas editor here. Moving text by a few pixels on
/// a phone is a poor experience, and the phone is where a creator *posts* —
/// so the app covers review, image swaps, caption and export, and hands off to
/// the web for fine editing.
class CarouselDetailScreen extends StatefulWidget {
  const CarouselDetailScreen({
    super.key,
    required this.api,
    required this.carouselId,
  });

  final Api api;
  final String carouselId;

  @override
  State<CarouselDetailScreen> createState() => _CarouselDetailScreenState();
}

class _CarouselDetailScreenState extends State<CarouselDetailScreen> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) {
    final s = Strings(Localizations.localeOf(context).languageCode);
    return Scaffold(
      appBar: AppBar(
        title: Text(_tab == 0 ? s.slides : s.caption),
        actions: [
          IconButton(
            tooltip: s.openOnWeb,
            icon: const Icon(Icons.open_in_new, size: 20),
            onPressed: () => launchUrl(
              Uri.parse(
                '${Api.baseUrl.replaceFirst(':8080', ':3000')}/carousels/${widget.carouselId}',
              ),
              mode: LaunchMode.externalApplication,
            ),
          ),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _tab,
        onDestinationSelected: (i) => setState(() => _tab = i),
        destinations: [
          NavigationDestination(
            icon: const Icon(Icons.view_carousel_outlined),
            label: s.slides,
          ),
          NavigationDestination(icon: const Icon(Icons.tag), label: s.caption),
        ],
      ),
      body: SafeArea(
        child: _tab == 0
            ? _SlidesTab(
                api: widget.api,
                carouselId: widget.carouselId,
                strings: s,
              )
            : _CaptionTab(
                api: widget.api,
                carouselId: widget.carouselId,
                strings: s,
              ),
      ),
    );
  }
}

class _SlidesTab extends StatefulWidget {
  const _SlidesTab({
    required this.api,
    required this.carouselId,
    required this.strings,
  });

  final Api api;
  final String carouselId;
  final Strings strings;

  @override
  State<_SlidesTab> createState() => _SlidesTabState();
}

class _SlidesTabState extends State<_SlidesTab> {
  late final DesignEditor _editor = DesignEditor(
    api: widget.api,
    carouselId: widget.carouselId,
  );
  final _pages = PageController(viewportFraction: 0.82);
  bool _editing = false;
  bool _loading = true;
  int _page = 0;

  @override
  void initState() {
    super.initState();
    _editor.addListener(_onEditorChanged);
    _editor.load().whenComplete(() {
      if (mounted) setState(() => _loading = false);
    });
  }

  void _onEditorChanged() => setState(() {});

  @override
  void dispose() {
    _editor.removeListener(_onEditorChanged);
    // Never leave an edit unsaved because the screen closed.
    _editor.save();
    _editor.dispose();
    _pages.dispose();
    super.dispose();
  }

  Future<void> _editText(DesignSlide slide, DesignElement element) async {
    final updated = await showModalBottomSheet<DesignElement>(
      context: context,
      isScrollControlled: true,
      builder: (_) =>
          TextEditorSheet(element: element, strings: widget.strings),
    );
    if (updated != null) _editor.updateElement(slide.id, updated);
  }

  @override
  Widget build(BuildContext context) {
    final s = widget.strings;
    if (_loading) return const Center(child: CircularProgressIndicator());

    final design = _editor.design;
    if (design == null) return EmptyState(title: s.somethingWentWrong);
    final slides = design.slides;

    return Column(
      children: [
        _EditBar(
          editing: _editing,
          editor: _editor,
          strings: s,
          onToggle: () => setState(() {
            _editing = !_editing;
            if (!_editing) _editor.select(null);
          }),
        ),
        Expanded(
          // The left inset keeps the screen edge clear of the PageView so iOS's
          // swipe-from-edge back gesture still reaches the route; a full-width
          // horizontal scroller swallows it. Mirrored on the right to keep the
          // carousel centred.
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            child: PageView.builder(
              controller: _pages,
              // While editing, horizontal drags move text rather than turn pages.
              physics: _editing
                  ? const NeverScrollableScrollPhysics()
                  : const PageScrollPhysics(),
              onPageChanged: (i) => setState(() => _page = i),
              itemCount: slides.length,
              itemBuilder: (context, i) => Padding(
                padding: const EdgeInsets.fromLTRB(8, 12, 8, 8),
                child: SlidePreview(
                  slide: slides[i],
                  canvas: design.canvas,
                  index: i + 1,
                  total: slides.length,
                  editing: _editing,
                  selectedElementId: _editor.selectedElementId,
                  onSelect: _editor.select,
                  onMove: (element, dx, dy) =>
                      _editor.moveBy(slides[i].id, element, dx, dy),
                  onEditText: (element) => _editText(slides[i], element),
                  onChangeImage: () async {
                    await Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => SlideImagesScreen(
                          api: widget.api,
                          carouselId: widget.carouselId,
                          slide: slides[i],
                        ),
                      ),
                    );
                    await _editor.load();
                  },
                  strings: s,
                ),
              ),
            ),
          ),
        ),
        if (_editing && slides.length > 1)
          Padding(
            padding: const EdgeInsets.only(bottom: 4),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                IconButton(
                  onPressed: _page == 0
                      ? null
                      : () => _pages.previousPage(
                          duration: const Duration(milliseconds: 200),
                          curve: Curves.easeOut,
                        ),
                  icon: const Icon(Icons.chevron_left),
                ),
                Text(
                  '${_page + 1} / ${slides.length}',
                  style: const TextStyle(color: AppColors.body),
                ),
                IconButton(
                  onPressed: _page >= slides.length - 1
                      ? null
                      : () => _pages.nextPage(
                          duration: const Duration(milliseconds: 200),
                          curve: Curves.easeOut,
                        ),
                  icon: const Icon(Icons.chevron_right),
                ),
              ],
            ),
          ),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
          child: _ExportButton(
            api: widget.api,
            carouselId: widget.carouselId,
            strings: s,
          ),
        ),
      ],
    );
  }
}

/// The edit toggle plus the save indicator.
class _EditBar extends StatelessWidget {
  const _EditBar({
    required this.editing,
    required this.editor,
    required this.strings,
    required this.onToggle,
  });

  final bool editing;
  final DesignEditor editor;
  final Strings strings;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context) {
    final status = editor.saving
        ? strings.saving
        : editor.conflictReloaded
        ? strings.reloaded
        : '';
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 8, 0),
      child: Row(
        children: [
          Expanded(
            child: Text(
              editing ? strings.tapToEdit : status,
              style: const TextStyle(color: AppColors.muted, fontSize: 12),
            ),
          ),
          if (!editing && status.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: Text(
                status,
                style: const TextStyle(color: AppColors.muted, fontSize: 12),
              ),
            ),
          TextButton.icon(
            onPressed: onToggle,
            icon: Icon(editing ? Icons.check : Icons.edit_outlined, size: 18),
            label: Text(editing ? strings.done : strings.edit),
          ),
        ],
      ),
    );
  }
}

class _ExportButton extends StatefulWidget {
  const _ExportButton({
    required this.api,
    required this.carouselId,
    required this.strings,
  });

  final Api api;
  final String carouselId;
  final Strings strings;

  @override
  State<_ExportButton> createState() => _ExportButtonState();
}

class _ExportButtonState extends State<_ExportButton> {
  bool _busy = false;
  Timer? _poll;
  String? _error;

  @override
  void dispose() {
    _poll?.cancel();
    super.dispose();
  }

  Future<void> _export() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final job = await widget.api.client.createExport(widget.carouselId);
      _poll = Timer.periodic(const Duration(milliseconds: 1200), (timer) async {
        final current = await widget.api.client.getExport(job.id);
        if (!mounted) return;
        if (current.status == 'completed') {
          timer.cancel();
          setState(() => _busy = false);
          final url = current.downloadUrl;
          if (url != null && url.isNotEmpty) {
            // The ZIP is shared as a link: the download is a time-limited
            // signed URL, so the sheet hands it to Files, TikTok or anything
            // else the creator wants to use.
            await SharePlus.instance.share(ShareParams(uri: Uri.parse(url)));
          }
        } else if (current.status == 'failed') {
          timer.cancel();
          setState(() {
            _busy = false;
            _error = widget.strings.somethingWentWrong;
          });
        }
      });
    } on ApiException catch (e) {
      setState(() {
        _busy = false;
        _error = widget.strings.error(e.code, e.message);
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        FilledButton.icon(
          onPressed: _busy ? null : _export,
          icon: _busy
              ? const SizedBox(
                  height: 16,
                  width: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.ios_share, size: 18),
          label: Text(_busy ? widget.strings.exporting : widget.strings.export),
        ),
        ErrorNote(_error),
      ],
    );
  }
}

class _CaptionTab extends StatefulWidget {
  const _CaptionTab({
    required this.api,
    required this.carouselId,
    required this.strings,
  });

  final Api api;
  final String carouselId;
  final Strings strings;

  @override
  State<_CaptionTab> createState() => _CaptionTabState();
}

class _CaptionTabState extends State<_CaptionTab> {
  late Future<CarouselCaption> _future;
  final _controller = TextEditingController();
  List<String> _hashtags = [];
  String? _loadedAt;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _future = widget.api.client.getCaption(widget.carouselId);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    try {
      await widget.api.client.updateCaption(
        widget.carouselId,
        UpdateCaptionRequest(caption: _controller.text, hashtags: _hashtags),
      );
    } on ApiException {
      // Surfaced on the next load; a failed autosave should not block typing.
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  Future<void> _regenerate() async {
    setState(
      () => _future = widget.api.client.regenerateCaption(widget.carouselId),
    );
    _loadedAt = null;
  }

  @override
  Widget build(BuildContext context) {
    final s = widget.strings;
    return FutureBuilder<CarouselCaption>(
      future: _future,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }
        if (!snapshot.hasData) return EmptyState(title: s.somethingWentWrong);

        final caption = snapshot.data!;
        final stamp = caption.updatedAt.toIso8601String();
        if (_loadedAt != stamp) {
          _loadedAt = stamp;
          _controller.text = caption.caption;
          _hashtags = List.of(caption.hashtags);
        }

        return ListView(
          padding: const EdgeInsets.all(16),
          children: [
            TextField(
              controller: _controller,
              maxLines: 8,
              decoration: const InputDecoration(),
              onEditingComplete: _save,
              onTapOutside: (_) => _save(),
            ),
            const SizedBox(height: 16),
            Wrap(
              spacing: 6,
              runSpacing: 6,
              children: [
                for (final tag in _hashtags)
                  Chip(
                    label: Text(tag, style: const TextStyle(fontSize: 12)),
                    onDeleted: () {
                      setState(() => _hashtags.remove(tag));
                      _save();
                    },
                  ),
              ],
            ),
            const SizedBox(height: 20),
            FilledButton.icon(
              onPressed: () async {
                await Clipboard.setData(
                  ClipboardData(
                    text: '${_controller.text}\n\n${_hashtags.join(' ')}',
                  ),
                );
                if (!context.mounted) return;
                ScaffoldMessenger.of(context)
                    .showSnackBar(SnackBar(content: Text(s.copied)));
              },
              icon: const Icon(Icons.copy, size: 18),
              label: Text(s.copyCaption),
            ),
            const SizedBox(height: 10),
            OutlinedButton.icon(
              onPressed: _regenerate,
              icon: const Icon(Icons.refresh, size: 18),
              label: Text(s.regenerateCaption),
            ),
            if (_saving)
              const Padding(
                padding: EdgeInsets.only(top: 12),
                child: Text('…', style: TextStyle(color: AppColors.muted)),
              ),
          ],
        );
      },
    );
  }
}
