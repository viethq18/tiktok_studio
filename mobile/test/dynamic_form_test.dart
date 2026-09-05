import 'package:carousel_studio/api/generated.dart';
import 'package:carousel_studio/i18n/strings.dart';
import 'package:carousel_studio/widgets/dynamic_form.dart';
import 'package:flutter_test/flutter_test.dart';

SchemaProperty _property(
  String name, {
  String type = 'string',
  String component = 'text',
  bool required = false,
  double? minimum,
  Object? defaultValue,
}) =>
    SchemaProperty(
      name: name,
      title: name,
      type: type,
      component: component,
      required: required,
      minimum: minimum,
      defaultValue: defaultValue,
    );

void main() {
  test('defaults honour the backend default, then the slider minimum', () {
    final form = SchemaForm(version: 1, properties: [
      _property('slide_count', type: 'integer', component: 'slider', minimum: 3),
      _property('cta', defaultValue: 'save'),
      _property('tags', type: 'array', component: 'multi_select'),
    ]);

    final values = DynamicForm.defaults(form);

    expect(values['slide_count'], 3);
    expect(values['cta'], 'save');
    expect(values['tags'], isEmpty);
  });

  test('validation reports required fields that are blank', () {
    final form = SchemaForm(version: 1, properties: [
      _property('topic', required: true),
      _property('angle'),
    ]);
    final strings = const Strings('en');

    expect(DynamicForm.validate(form, {'topic': '   '}, strings), contains('topic'));
    expect(DynamicForm.validate(form, {'topic': 'sleep signs'}, strings), isEmpty);
  });
}
