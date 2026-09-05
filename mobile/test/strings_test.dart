import 'package:carousel_studio/i18n/strings.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('pipeline step labels come from the client, keyed by backend status', () {
    // The backend sends a stable status; the wording must follow the UI
    // language, not the server's.
    expect(const Strings('vi').step('searching_images'), 'Đang tìm hình ảnh');
    expect(const Strings('en').step('searching_images'), 'Finding images');
  });

  test('an unknown step falls back to the raw status rather than blank text', () {
    expect(const Strings('en').step('some_new_step'), 'some_new_step');
  });

  test('errors localise by code and fall back to the server message', () {
    expect(const Strings('vi').error('rate_limited', 'ignored'),
        'Bạn đã đạt giới hạn hôm nay.');
    expect(const Strings('en').error('unknown_code', 'Server said this'),
        'Server said this');
    expect(const Strings('en').error('unknown_code', ''),
        const Strings('en').somethingWentWrong);
  });
}
