import 'package:flutter/material.dart';

import '../api/generated.dart';
import '../i18n/strings.dart';
import 'common.dart';

/// Renders the AI-generated input form.
///
/// The backend decides which fields exist; this file decides how each one is
/// drawn. The registry below is closed — a component the backend did not
/// approve simply has no entry, so it cannot be rendered (blueprint §9.1).
class DynamicForm extends StatelessWidget {
  const DynamicForm({
    super.key,
    required this.form,
    required this.values,
    required this.errors,
    required this.onChanged,
  });

  final SchemaForm form;
  final Map<String, Object?> values;
  final Map<String, String> errors;
  final void Function(String name, Object? value) onChanged;

  static Map<String, Object?> defaults(SchemaForm form) {
    final out = <String, Object?>{};
    for (final p in form.properties) {
      if (p.defaultValue != null) {
        out[p.name] = p.defaultValue;
      } else if (p.type == 'array') {
        out[p.name] = <String>[];
      } else if (p.type == 'boolean') {
        out[p.name] = false;
      } else if (p.component == 'slider' && p.minimum != null) {
        out[p.name] = p.minimum!.toInt();
      }
    }
    return out;
  }

  /// Local validation is for feedback only; the backend validates again.
  static Map<String, String> validate(SchemaForm form, Map<String, Object?> values, Strings s) {
    final errors = <String, String>{};
    for (final p in form.properties) {
      final v = values[p.name];
      final empty = v == null || (v is String && v.trim().isEmpty) || (v is List && v.isEmpty);
      if (p.required && empty) errors[p.name] = '${p.title} ${s.required}';
    }
    return errors;
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (final property in form.properties) ...[
          Text(property.title, style: const TextStyle(fontWeight: FontWeight.w600)),
          if (property.description != null && property.description!.isNotEmpty) ...[
            const SizedBox(height: 3),
            Text(property.description!,
                style: const TextStyle(color: AppColors.body, fontSize: 12)),
          ],
          const SizedBox(height: 8),
          _field(property),
          if (errors[property.name] != null) ...[
            const SizedBox(height: 6),
            Text(errors[property.name]!,
                style: const TextStyle(color: Color(0xFFB91C1C), fontSize: 12)),
          ],
          const SizedBox(height: 20),
        ],
      ],
    );
  }

  Widget _field(SchemaProperty p) {
    switch (p.component) {
      case 'textarea':
        return _text(p, lines: 3);
      case 'number':
        return _number(p);
      case 'slider':
        return _slider(p);
      case 'radio':
      case 'select':
        return _choice(p);
      case 'multi_select':
        return _multi(p);
      case 'checkbox':
        return _checkbox(p);
      // 'text', 'url' and 'image' all take a single line of text on mobile.
      default:
        return _text(p, lines: 1);
    }
  }

  String _label(SchemaProperty p, String value) => p.enumLabels?[value] ?? value;

  Widget _text(SchemaProperty p, {required int lines}) {
    return TextFormField(
      initialValue: values[p.name] as String? ?? '',
      maxLines: lines,
      maxLength: p.maxLength,
      decoration: InputDecoration(
        hintText: p.placeholder,
        counterText: p.maxLength == null ? '' : null,
      ),
      onChanged: (v) => onChanged(p.name, v),
    );
  }

  Widget _number(SchemaProperty p) {
    return TextFormField(
      initialValue: '${values[p.name] ?? ''}',
      keyboardType: TextInputType.number,
      onChanged: (v) => onChanged(p.name, int.tryParse(v)),
    );
  }

  Widget _slider(SchemaProperty p) {
    final min = (p.minimum ?? 3).toDouble();
    final max = (p.maximum ?? 7).toDouble();
    final current = ((values[p.name] as num?) ?? min).toDouble().clamp(min, max);
    return Row(
      children: [
        Text('${min.toInt()}', style: const TextStyle(color: AppColors.muted)),
        Expanded(
          child: Slider(
            value: current,
            min: min,
            max: max,
            divisions: (max - min).toInt(),
            label: '${current.toInt()}',
            activeColor: AppColors.ink,
            onChanged: (v) => onChanged(p.name, v.toInt()),
          ),
        ),
        Text('${max.toInt()}', style: const TextStyle(color: AppColors.muted)),
        const SizedBox(width: 8),
        Text('${current.toInt()}', style: const TextStyle(fontWeight: FontWeight.w700)),
      ],
    );
  }

  Widget _choice(SchemaProperty p) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        for (final option in p.enumValue ?? const <String>[])
          ChoiceChip(
            label: Text(_label(p, option)),
            selected: values[p.name] == option,
            showCheckmark: false,
            onSelected: (_) => onChanged(p.name, option),
          ),
      ],
    );
  }

  Widget _multi(SchemaProperty p) {
    final selected = (values[p.name] as List?)?.cast<String>() ?? const <String>[];
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        for (final option in p.enumValue ?? const <String>[])
          FilterChip(
            label: Text(_label(p, option)),
            selected: selected.contains(option),
            showCheckmark: false,
            onSelected: (on) => onChanged(
              p.name,
              on
                  ? [...selected, option]
                  : selected.where((e) => e != option).toList(),
            ),
          ),
      ],
    );
  }

  Widget _checkbox(SchemaProperty p) {
    return CheckboxListTile(
      contentPadding: EdgeInsets.zero,
      controlAffinity: ListTileControlAffinity.leading,
      value: values[p.name] == true,
      title: Text(p.placeholder ?? p.title),
      onChanged: (v) => onChanged(p.name, v ?? false),
    );
  }
}
