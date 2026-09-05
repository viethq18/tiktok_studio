/// UI strings.
///
/// The web app resolves its locale from a cookie; the app follows the device
/// locale instead. Content language is a separate, per-project setting — the
/// two are deliberately independent (blueprint §164).
class Strings {
  const Strings(this.locale);

  final String locale;

  bool get _vi => locale.startsWith('vi');

  String _(String vi, String en) => _vi ? vi : en;

  String get appName => 'Carousel Studio';
  String get signIn => _('Đăng nhập', 'Sign in');
  String get signInSubtitle => _(
        'Nhập một ý tưởng, nhận về carousel gần như hoàn chỉnh.',
        'Bring one idea. Leave with a finished carousel.',
      );
  String get email => 'Email';
  String get emailHint => _('ban@example.com', 'you@example.com');
  String get signInWithEmail => _('Đăng nhập bằng email', 'Sign in with email');
  String get signOut => _('Đăng xuất', 'Sign out');
  String get signingIn => _('Đang đăng nhập…', 'Signing in…');

  String get projects => _('Projects', 'Projects');
  String get newProject => _('Project mới', 'New project');
  String get noProjects => _('Chưa có project nào', 'No projects yet');
  String get noProjectsBody => _(
        'Mỗi project là bộ não nội dung của một kênh TikTok.',
        'Each project is the content brain of one TikTok channel.',
      );
  String carouselCount(int n) => _('$n carousel', '$n carousels');

  String get channelQuestion =>
      _('Kênh của bạn nói về điều gì?', 'What is your channel about?');
  String get channelHint => _(
        'VD: Chăm sóc trẻ sơ sinh 0–3 tuổi',
        'e.g. Baby care for 0–3 year olds',
      );
  String get contentLanguage => _('Ngôn ngữ nội dung', 'Content language');
  String get contentLanguageHint => _(
        'AI sẽ viết toàn bộ carousel của project này bằng ngôn ngữ bạn chọn.',
        'The AI writes every carousel in this project in the language you pick.',
      );
  String get pickLanguage => _('Hãy chọn ngôn ngữ nội dung.', 'Pick a content language.');
  String get create => _('Tạo project', 'Create project');
  String get analysing =>
      _('AI đang tìm hiểu kênh của bạn…', 'Studying your channel…');

  String get newCarousel => _('Tạo carousel', 'New carousel');
  String get noCarousels => _('Chưa có carousel nào', 'No carousels yet');
  String get generate => _('Tạo carousel', 'Generate carousel');
  String get generating => _('Đang tạo…', 'Generating…');
  String get required => _('bắt buộc', 'required');
  String get canvasFormat => _('Khổ ảnh', 'Canvas format');
  String get recommended => _('Khuyên dùng', 'Recommended');

  String get slides => _('Slide', 'Slides');
  String get caption => _('Caption', 'Caption');
  String get images => _('Hình ảnh', 'Images');
  String get copyCaption =>
      _('Sao chép caption + hashtag', 'Copy caption + hashtags');
  String get copied => _('Đã sao chép', 'Copied');
  String get regenerateCaption => _('Viết lại bằng AI', 'Rewrite with AI');
  String get export => _('Export ZIP', 'Export ZIP');
  String get exporting => _('Đang export…', 'Exporting…');
  String get share => _('Chia sẻ', 'Share');
  String get openOnWeb => _('Mở trên web để chỉnh sâu', 'Open on the web to edit in detail');
  String get searchImages => _('Tìm ảnh', 'Search images');
  String get searchMore => _('Tìm thêm ảnh', 'Find more images');
  String get searching => _('Đang tìm…', 'Searching…');
  String get imageKeyword => _('Từ khoá tìm ảnh', 'Image keywords');
  String get retry => _('Thử lại', 'Try again');
  String get edit => _('Sửa', 'Edit');
  String get done => _('Xong', 'Done');
  String get fontSize => _('Cỡ chữ', 'Size');
  String get alignLeft => _('Trái', 'Left');
  String get alignCenter => _('Giữa', 'Centre');
  String get alignRight => _('Phải', 'Right');
  String get tapToEdit =>
      _('Chạm để chọn, kéo để di chuyển, chạm lần nữa để sửa chữ.',
        'Tap to select, drag to move, tap again to edit the text.');
  String get tooLong => _(
        'Dài hơn giới hạn — chữ có thể tràn khỏi vùng an toàn.',
        'Over the limit — this may overflow the safe area.',
      );
  String get saving => _('Đang lưu…', 'Saving…');
  String get saved => _('Đã lưu', 'Saved');
  String get reloaded =>
      _('Đã tải lại bản mới nhất', 'Reloaded the newest version');
  String get somethingWentWrong =>
      _('Đã có lỗi xảy ra. Vui lòng thử lại.', 'Something went wrong. Please try again.');

  /// Localised copy for the pipeline steps the backend reports.
  String step(String status) {
    switch (status) {
      case 'researching':
        return _('Đang nghiên cứu chủ đề', 'Researching the topic');
      case 'generating_content':
        return _('Đang viết nội dung', 'Writing the copy');
      case 'validating_content':
        return _('Đang kiểm tra nội dung', 'Checking the copy');
      case 'selecting_formula':
        return _('Đang chọn công thức', 'Choosing a formula');
      case 'generating_design':
        return _('Đang thiết kế slide', 'Designing the slides');
      case 'searching_images':
        return _('Đang tìm hình ảnh', 'Finding images');
      case 'finalizing':
        return _('Đang hoàn thiện', 'Finishing up');
      default:
        return status;
    }
  }

  /// Backend errors travel as stable codes; the wording is chosen here.
  String error(String code, String fallback) {
    switch (code) {
      case 'unauthorized':
        return _('Bạn cần đăng nhập.', 'Please sign in.');
      case 'rate_limited':
        return _('Bạn đã đạt giới hạn hôm nay.', 'You have hit today’s limit.');
      case 'ai_unavailable':
        return _('AI đang không phản hồi.', 'The AI is not responding.');
      case 'conflict':
        return _('Dữ liệu đã thay đổi ở nơi khác.', 'This changed somewhere else.');
      default:
        return fallback.isEmpty ? somethingWentWrong : fallback;
    }
  }
}
