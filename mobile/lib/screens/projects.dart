import 'package:flutter/material.dart';

import '../api/generated.dart';
import '../api/session.dart';
import '../i18n/strings.dart';
import '../widgets/common.dart';
import 'carousels.dart';
import 'new_project.dart';

class ProjectsScreen extends StatefulWidget {
  const ProjectsScreen({super.key, required this.api, required this.onSignedOut});

  final Api api;
  final VoidCallback onSignedOut;

  @override
  State<ProjectsScreen> createState() => _ProjectsScreenState();
}

class _ProjectsScreenState extends State<ProjectsScreen> {
  late Future<List<Project>> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<Project>> _load() async =>
      (await widget.api.client.listProjects()).projects;

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
        title: Text(s.projects,
            style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 18)),
        actions: [
          IconButton(
            tooltip: s.signOut,
            icon: const Icon(Icons.logout, size: 20),
            onPressed: () async {
              await widget.api.signOut();
              widget.onSignedOut();
            },
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        backgroundColor: AppColors.ink,
        foregroundColor: Colors.white,
        onPressed: () async {
          final created = await Navigator.of(context).push<bool>(
            MaterialPageRoute(builder: (_) => NewProjectScreen(api: widget.api)),
          );
          if (created ?? false) await _refresh();
        },
        icon: const Icon(Icons.add),
        label: Text(s.newProject),
      ),
      body: FutureBuilder<List<Project>>(
        future: _future,
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snapshot.hasError) {
            return EmptyState(
              title: s.somethingWentWrong,
              action: OutlinedButton(onPressed: _refresh, child: Text(s.retry)),
            );
          }
          final projects = snapshot.data ?? const [];
          if (projects.isEmpty) {
            return EmptyState(title: s.noProjects, body: s.noProjectsBody);
          }
          return RefreshIndicator(
            onRefresh: _refresh,
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 96),
              itemCount: projects.length,
              separatorBuilder: (_, _) => const SizedBox(height: 12),
              itemBuilder: (context, i) {
                final p = projects[i];
                return AppCard(
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (_) => CarouselsScreen(api: widget.api, project: p),
                    ),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(p.name,
                          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                      const SizedBox(height: 4),
                      Text(p.niche,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(color: AppColors.body, fontSize: 13)),
                      const SizedBox(height: 10),
                      Row(
                        children: [
                          Text(s.carouselCount(p.carouselCount),
                              style: const TextStyle(color: AppColors.muted, fontSize: 12)),
                          const SizedBox(width: 8),
                          Text('· ${p.language.toUpperCase()}',
                              style: const TextStyle(color: AppColors.muted, fontSize: 12)),
                        ],
                      ),
                    ],
                  ),
                );
              },
            ),
          );
        },
      ),
    );
  }
}
