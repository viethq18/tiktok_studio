import 'dart:async';

import 'package:flutter/foundation.dart';

import '../api/generated.dart';
import '../api/session.dart';

/// Holds the design being edited and pushes changes back to the server.
///
/// The phone edits the same Design JSON the web editor and the exporter use, so
/// everything here works in logical canvas coordinates (§99). Saving carries the
/// version the client holds; a 409 means another device won, and the newest
/// design is adopted rather than overwritten (§80).
class DesignEditor extends ChangeNotifier {
  DesignEditor({required this.api, required this.carouselId});

  static const _debounce = Duration(milliseconds: 1200);

  final Api api;
  final String carouselId;

  Design? design;
  int version = 0;
  String? selectedElementId;
  bool saving = false;
  bool conflictReloaded = false;
  String? error;

  Timer? _timer;
  bool _inFlight = false;
  bool _disposed = false;

  /// Leaving the screen while a save is in flight used to notify listeners
  /// after disposal. Every notification goes through here instead.
  void _notify() {
    if (!_disposed) notifyListeners();
  }

  Future<void> load() async {
    try {
      final view = await api.client.getDesign(carouselId);
      design = view.design;
      version = view.version;
      error = null;
    } on ApiException catch (e) {
      error = e.code;
    } catch (_) {
      // Offline, DNS failure, anything else: the screen shows an error state
      // rather than letting the future fail silently.
      error = 'network';
    }
    _notify();
  }

  DesignSlide? slideAt(int index) {
    final slides = design?.slides;
    if (slides == null || index < 0 || index >= slides.length) return null;
    return slides[index];
  }

  void select(String? elementId) {
    if (selectedElementId == elementId) return;
    selectedElementId = elementId;
    _notify();
  }

  /// Replaces one element and schedules a save.
  void updateElement(String slideId, DesignElement element) {
    final current = design;
    if (current == null) return;
    design = current.copyWith(
      slides: [
        for (final slide in current.slides)
          if (slide.id != slideId)
            slide
          else
            slide.copyWith(elements: [
              for (final existing in slide.elements)
                existing.id == element.id ? element : existing,
            ]),
      ],
    );
    _notify();
    _scheduleSave();
  }

  /// Moves a text element by a delta in logical canvas units, keeping it inside
  /// the safe area — the same box the backend validator enforces (§123).
  void moveBy(String slideId, DesignElement element, double dx, double dy) {
    final canvas = design?.canvas;
    if (canvas == null) return;
    final safe = canvas.safeArea;
    final maxX = (safe.x + safe.width - element.width).toDouble();
    final maxY = (safe.y + safe.height - element.height).toDouble();
    updateElement(
      slideId,
      element.copyWith(
        x: (element.x + dx).clamp(safe.x.toDouble(), maxX < safe.x ? safe.x.toDouble() : maxX),
        y: (element.y + dy).clamp(safe.y.toDouble(), maxY < safe.y ? safe.y.toDouble() : maxY),
      ),
    );
  }

  void _scheduleSave() {
    _timer?.cancel();
    _timer = Timer(_debounce, save);
  }

  Future<void> save() async {
    final current = design;
    if (current == null || _inFlight) return;
    _inFlight = true;
    saving = true;
    error = null;
    _notify();
    try {
      final saved = await api.client.saveDesign(
        carouselId,
        SaveDesignRequest(version: version, design: current),
      );
      version = saved.version;
      design = saved.design;
      conflictReloaded = false;
    } on ApiException catch (e) {
      if (e.statusCode == 409) {
        // Someone else saved first. Take their version rather than clobbering it.
        final fresh = await api.client.getDesign(carouselId);
        design = fresh.design;
        version = fresh.version;
        conflictReloaded = true;
      } else {
        error = e.code;
      }
    } finally {
      _inFlight = false;
      saving = false;
      _notify();
    }
  }

  /// Flushes any pending edit, then tears down. Callers just call dispose():
  /// asking them to save first invited a save that outlived the object.
  @override
  void dispose() {
    _timer?.cancel();
    final pending = design;
    _disposed = true;
    if (pending != null) unawaited(_flush(pending));
    super.dispose();
  }

  /// A last save with no state updates, since nothing is listening any more.
  Future<void> _flush(Design pending) async {
    try {
      await api.client.saveDesign(
        carouselId,
        SaveDesignRequest(version: version, design: pending),
      );
    } catch (_) {
      // The screen is gone; there is nobody to tell.
    }
  }
}
