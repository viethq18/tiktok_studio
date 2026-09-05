import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../api/generated.dart';
import '../api/session.dart';
import '../i18n/strings.dart';
import '../widgets/common.dart';

/// Picks the background image for one slide.
///
/// Selection is by candidate id, never by URL: the server resolves the id
/// against its own cached search results so the backend never fetches a URL the
/// client supplied (blueprint §84).
class SlideImagesScreen extends StatefulWidget {
  const SlideImagesScreen({
    super.key,
    required this.api,
    required this.carouselId,
    required this.slide,
  });

  final Api api;
  final String carouselId;
  final DesignSlide slide;

  @override
  State<SlideImagesScreen> createState() => _SlideImagesScreenState();
}

class _SlideImagesScreenState extends State<SlideImagesScreen> {
  final _keyword = TextEditingController();
  List<ImagePublicCandidate> _candidates = [];
  bool _loading = true;
  bool _applying = false;
  int _page = 1;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _keyword.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final res = await widget.api.client
          .getSlideImages(widget.carouselId, widget.slide.id);
      if (!mounted) return;
      setState(() {
        _candidates = res.candidates;
        _keyword.text = res.query;
        _loading = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = Strings(Localizations.localeOf(context).languageCode)
            .error(e.code, e.message);
      });
    }
  }

  Future<void> _search({bool append = false}) async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final res = await widget.api.client.searchSlideImages(
        widget.carouselId,
        widget.slide.id,
        SearchImagesRequest(query: _keyword.text.trim(), page: append ? _page + 1 : 1),
      );
      if (!mounted) return;
      setState(() {
        _page = res.page ?? 1;
        _candidates = append ? [..._candidates, ...res.candidates] : res.candidates;
        _loading = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = Strings(Localizations.localeOf(context).languageCode)
            .error(e.code, e.message);
      });
    }
  }

  Future<void> _select(ImagePublicCandidate candidate) async {
    setState(() => _applying = true);
    try {
      // version 0 means "whatever the server currently has": the phone has no
      // local edits to protect, so there is nothing to conflict with.
      await widget.api.client.selectSlideImage(
        widget.carouselId,
        widget.slide.id,
        SelectImageRequest(candidateId: candidate.id, version: 0),
      );
      if (mounted) Navigator.of(context).pop(true);
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _applying = false;
        _error = Strings(Localizations.localeOf(context).languageCode)
            .error(e.code, e.message);
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = Strings(Localizations.localeOf(context).languageCode);
    return Scaffold(
      appBar: AppBar(title: Text('${s.images} · ${widget.slide.index}')),
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(children: [
                Expanded(
                  child: TextField(
                    controller: _keyword,
                    textInputAction: TextInputAction.search,
                    decoration: InputDecoration(hintText: s.imageKeyword),
                    onSubmitted: (_) => _search(),
                  ),
                ),
                const SizedBox(width: 8),
                IconButton.filled(
                  style: IconButton.styleFrom(backgroundColor: AppColors.ink),
                  onPressed: _loading ? null : () => _search(),
                  icon: const Icon(Icons.search),
                ),
              ]),
            ),
            if (_error != null) Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: ErrorNote(_error),
            ),
            Expanded(
              child: _loading && _candidates.isEmpty
                  ? const Center(child: CircularProgressIndicator())
                  : GridView.builder(
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 2,
                        crossAxisSpacing: 12,
                        mainAxisSpacing: 12,
                        childAspectRatio: 0.8,
                      ),
                      itemCount: _candidates.length,
                      itemBuilder: (context, i) {
                        final candidate = _candidates[i];
                        return InkWell(
                          onTap: _applying ? null : () => _select(candidate),
                          borderRadius: BorderRadius.circular(10),
                          child: ClipRRect(
                            borderRadius: BorderRadius.circular(10),
                            child: Stack(fit: StackFit.expand, children: [
                              CachedNetworkImage(
                                imageUrl: candidate.thumbUrl,
                                fit: BoxFit.cover,
                              ),
                              if (candidate.photographerName != null)
                                Positioned(
                                  left: 0,
                                  right: 0,
                                  bottom: 0,
                                  child: Container(
                                    color: Colors.black54,
                                    padding: const EdgeInsets.symmetric(
                                        horizontal: 6, vertical: 3),
                                    child: Text(
                                      candidate.photographerName!,
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: const TextStyle(
                                          color: Colors.white, fontSize: 10),
                                    ),
                                  ),
                                ),
                            ]),
                          ),
                        );
                      },
                    ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
              child: OutlinedButton(
                onPressed: _loading ? null : () => _search(append: true),
                child: Text(s.searchImages),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
