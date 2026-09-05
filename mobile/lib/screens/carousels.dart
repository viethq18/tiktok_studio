import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../api/generated.dart';
import '../api/session.dart';
import '../i18n/strings.dart';
import '../widgets/common.dart';
import 'carousel_detail.dart';
import 'new_carousel.dart';

class CarouselsScreen extends StatefulWidget {
  const CarouselsScreen({super.key, required this.api, required this.project});

  final Api api;
  final Project project;

  @override
  State<CarouselsScreen> createState() => _CarouselsScreenState();
}

class _CarouselsScreenState extends State<CarouselsScreen> {
  late Future<List<Carousel>> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<Carousel>> _load() async =>
      (await widget.api.client.listCarousels(widget.project.id)).carousels;

  Future<void> _refresh() async {
    final future = _load();
    setState(() => _future = future);
    await future;
  }

  @override
  Widget build(BuildContext context) {
    final s = Strings(Localizations.localeOf(context).languageCode);
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.project.name,
            style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 17)),
      ),
      floatingActionButton: FloatingActionButton.extended(
        backgroundColor: AppColors.ink,
        foregroundColor: Colors.white,
        onPressed: () async {
          final created = await Navigator.of(context).push<bool>(
            MaterialPageRoute(
              builder: (_) => NewCarouselScreen(api: widget.api, project: widget.project),
            ),
          );
          if (created ?? false) await _refresh();
        },
        icon: const Icon(Icons.add),
        label: Text(s.newCarousel),
      ),
      body: FutureBuilder<List<Carousel>>(
        future: _future,
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          }
          final carousels = snapshot.data ?? const <Carousel>[];
          if (carousels.isEmpty) {
            return RefreshIndicator(
              onRefresh: _refresh,
              child: ListView(children: [
                SizedBox(
                  height: MediaQuery.sizeOf(context).height * 0.6,
                  child: EmptyState(title: s.noCarousels),
                ),
              ]),
            );
          }
          return RefreshIndicator(
            onRefresh: _refresh,
            child: GridView.builder(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 96),
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                childAspectRatio: 0.62,
              ),
              itemCount: carousels.length,
              itemBuilder: (context, i) =>
                  _CarouselTile(api: widget.api, carousel: carousels[i], onChanged: _refresh),
            ),
          );
        },
      ),
    );
  }
}

class _CarouselTile extends StatelessWidget {
  const _CarouselTile({required this.api, required this.carousel, required this.onChanged});

  final Api api;
  final Carousel carousel;
  final Future<void> Function() onChanged;

  @override
  Widget build(BuildContext context) {
    final thumbnail = carousel.thumbnailUrl;
    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: () async {
        await Navigator.of(context).push(
          MaterialPageRoute(
            builder: (_) => CarouselDetailScreen(api: api, carouselId: carousel.id),
          ),
        );
        await onChanged();
      },
      child: Container(
        clipBehavior: Clip.antiAlias,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: AppColors.line),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Expanded(
              child: Container(
                color: const Color(0xFFF5F5F5),
                child: thumbnail == null || thumbnail.isEmpty
                    ? Center(
                        child: carousel.status == 'generating'
                            ? const SizedBox(
                                height: 20,
                                width: 20,
                                child: CircularProgressIndicator(strokeWidth: 2))
                            : const Icon(Icons.image_outlined, color: AppColors.muted),
                      )
                    : CachedNetworkImage(imageUrl: thumbnail, fit: BoxFit.cover),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(10),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(carousel.title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                  const SizedBox(height: 6),
                  StatusBadge(carousel.status),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
