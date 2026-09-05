# Mobile — trạng thái và việc còn lại

Cập nhật: 2026-09-05 · Flutter 3.47 / Dart 3.13 · `mobile/`

---

## 1. Tóm tắt

App Flutter cho iOS và Android, dùng chung backend với web. Có sửa và di chuyển
chữ; không có editor canvas đầy đủ — xem mục 3.

| Hạng mục | Trạng thái |
|---|---|
| Màn hình | Sign in · Projects · New project · Carousels · New carousel + tiến trình · Carousel detail (slides / caption) · Slide images · Sửa chữ |
| API client | **Sinh tự động** từ struct Go — 37 method, 52 model |
| Test | 31 widget test + 1 integration test trên simulator · `flutter analyze` sạch |
| iOS | Build và chạy được trên simulator (đã kiểm chứng: auth, danh sách project với dữ liệu thật) |
| Android | Chưa chạy thử |

---

## 2. Contract sinh tự động — đọc phần này trước

Không sửa tay `mobile/lib/api/generated.dart`.

```
backend/internal/apispec/spec.go   ← khai báo endpoint bằng CHÍNH struct Go
                │  make apigen
                ├──► docs/openapi.yaml
                ├──► mobile/lib/api/generated.dart      (model + client Dart)
                └──► frontend/types/api.generated.ts    (type cho web)
```

Đổi struct Go mà quên chạy `make apigen` thì test
`TestGeneratedClientsAreUpToDate` fail. Đổi field mà web chưa cập nhật thì
`frontend/types/contract.test-d.ts` làm `tsc` fail. Hai chốt chặn này là lý do
duy nhất khiến ba client (Go, Dart, TypeScript) không trôi khỏi nhau.

Thêm endpoint mới: khai báo trong `Operations()` rồi `make apigen` — client Dart
tự có method mới.

---

## 3. Editor trên mobile làm được gì

Bật nút **Sửa** trên tab Slides:

- **Chạm** một khối chữ để chọn, **kéo** để di chuyển, **chạm lần nữa** để mở
  sheet sửa nội dung, cỡ chữ và căn lề.
- Kéo được clamp trong **safe area** — đúng vùng backend enforce (§123), nên
  preview không bao giờ hiện vị trí mà server sẽ dời đi.
- Autosave debounce 1,2 giây, mang theo `version`. Gặp 409 thì lấy bản mới của
  server thay vì ghi đè (§80).
- Khi đang sửa, `PageView` tắt vuốt ngang để kéo chữ không bị hiểu nhầm là lật
  slide; có nút mũi tên để chuyển slide.

**Chưa có:** đổi font family, đổi màu, resize khối chữ, di chuyển ảnh/shape,
undo/redo. Những thứ này ở web. Lý do không port cả editor: Fabric.js là thư
viện canvas của trình duyệt, và quan trọng hơn là nó kéo theo **bản thứ ba** của
thuật toán đo chữ (`fontkit` bên Go, `measure.ts` bên web) — ba bản là ba cơ hội
lệch nhau, đúng loại bug §157 cảnh báo. Di chuyển và sửa nội dung không cần đo
chữ nên an toàn; wrap và overflow vẫn do backend quyết.

`SlidePreview` render Design JSON **trung thực** theo đúng toạ độ tuyệt đối mà
exporter dùng, chỉ scale lại. Lớp kéo thả nằm **trên** lớp vẽ, tách biệt, nên
thứ bạn kéo đúng là thứ được export.

---

## 4. Vấn đề đã biết

### 4.1 Chặn việc dùng thật

| Vấn đề | Chi tiết |
|---|---|
| **Chỉ có dev login** | Google OAuth trên web là redirect trình duyệt. iOS cần `ASWebAuthenticationSession` + deep link, và backend phải nhận redirect URI của app. Hiện chỉ có đăng nhập bằng email — đủ để phát triển, không đủ để phát hành. |
| **`API_URL` mặc định là `localhost:8080`** | Chạy được trên simulator, **không** chạy trên máy thật. Build với `--dart-define=API_URL=https://…`. |
| **MinIO trả presigned URL `localhost:9100`** | Ảnh và file export sẽ không tải được trên máy thật cho tới khi backend deploy có domain thật. |
| **Chưa chạy thử Android** | Code không có gì phụ thuộc iOS, nhưng chưa build lần nào. |

### 4.2 Giới hạn đã biết

- **Session là cookie, không phải token.** Backend dùng httpOnly cookie; Dart
  không có cookie jar nên client tự bắt `Set-Cookie` và gửi lại. Có test cho
  đường này (`test/api_client_test.dart`) vì nó là toàn bộ cơ chế auth trên
  mobile. Nếu sau này backend đổi sang token thì đây là chỗ sửa.
- **Chọn ảnh gửi `version: 0`** — nghĩa là "lấy bản server đang có". App không
  giữ chỉnh sửa cục bộ nên không có gì để xung đột; web thì ngược lại, phải gửi
  version thật để optimistic concurrency hoạt động.
- **Không có offline / cache.** Mất mạng là màn hình trắng kèm nút thử lại.
- **UI chỉ vi/en**, theo locale máy. Ngôn ngữ nội dung vẫn theo project, độc lập.
- **Export = share link.** Backend trả signed URL; app mở share sheet với link
  đó thay vì tải file về. Đơn giản và đủ dùng, nhưng nếu muốn "lưu vào Ảnh" thì
  phải tải ZIP, giải nén và ghi vào photo library.
- **Chưa có màn hình project settings** (brand kit, chiến lược AI) — làm trên web.

### 4.3 Bug đã sửa — giữ lại để biết vì sao code viết như vậy

**Nút tìm ảnh chết sau lần đầu (đã sửa).** `SlideImagesScreen` chỉ bắt
`ApiException`; bất kỳ lỗi nào khác để `_loading` kẹt `true` **vĩnh viễn** →
cả nút search lẫn nút tìm thêm đều bị disable, không báo gì. Cùng lỗi ở
`_applying` khi chọn ảnh thất bại: mọi ô ảnh thành không bấm được. Giờ mọi lời
gọi mạng đi qua một helper có `finally`.

Kèm theo hai lỗi khiến nó *trông như* search hỏng: nút dưới cùng ghi "Tìm ảnh"
nhưng thực ra là *tải thêm trang sau* — gõ từ khoá mới rồi bấm nó sẽ nối kết quả
của truy vấn cũ vào cuối danh sách, ngoài tầm nhìn. Giờ ghi "Tìm thêm ảnh". Và
`_page` bị reset về 1 khi response thiếu trường `page`, nên "tìm thêm" lấy lại
đúng trang cũ.

**Dùng `DesignEditor` sau khi dispose (đã sửa).** Rời màn hình carousel trong lúc
autosave đang chạy thì `save()` gọi `notifyListeners()` trên object đã huỷ.
`dispose()` giờ tự flush bản chưa lưu và không notify nữa. Lỗi này do
integration test tìm ra, không phải widget test.

**Không back được từ màn hình slide (đã sửa).** `PageView` cuộn ngang phủ kín
chiều rộng nên nuốt luôn cử chỉ vuốt-từ-mép-trái mà iOS dùng để back. Sửa bằng
cách chừa 20pt hai bên để mép trái thuộc về route chứ không thuộc PageView.
`test/navigation_test.dart` kiểm tra nút back tồn tại **và** thực sự pop được,
kể cả khi tải design thất bại — màn hình cụt là loại bug không crash nên rất dễ
lọt.

**Slide viewer bị lệch (đã sửa).** `SlidePreview` tính `scale` chỉ theo chiều
cao (`maxHeight / canvas.height`). Trên khung hẹp của `PageView`, slide rộng hơn
chỗ được cấp nên bị cắt lệch và méo tỉ lệ (đo được 0.59 thay vì 0.8). Sửa bằng
cách lấy `min` của cả hai trục. Test `the slide fits inside a narrow viewport`
canh chừng: nó kiểm tra slide nằm gọn trong khung **và** giữ đúng tỉ lệ 4:5.

Đã đối chiếu trực tiếp: preview trên simulator ngắt dòng, đặt gạch đỏ và giữ tỉ
lệ giống hệt file PNG mà exporter Go sinh ra cho cùng Design JSON.

### 4.4 Chưa kiểm chứng

Đã chạy thật trên simulator: build, cài, đăng nhập bằng session thật, danh sách
project hiển thị đúng dữ liệu tiếng Việt từ backend.

**Chưa** bấm qua được các màn hình sâu hơn bằng tự động — điều khiển simulator
cần quyền Accessibility. Các phần đó được phủ bằng widget test thay thế
(`SlidePreview` với Design JSON thật, dynamic form, API client, strings).

---

## 5. Việc tiếp theo, theo thứ tự

1. **Deploy backend** có domain thật, đặt `MINIO_PUBLIC_ENDPOINT` cho đúng.
   Không có bước này thì app không chạy nổi ngoài simulator.
2. **Google sign-in trên mobile**: `ASWebAuthenticationSession` + deep link, và
   backend nhận redirect URI của app.
3. **Build và chạy thử Android.**
4. **Quyết định chuyện thanh toán trước khi viết thêm.** Bán subscription trong
   app iOS bắt buộc dùng In-App Purchase, Apple ăn 15–30%. Bán trên web rồi app
   chỉ đăng nhập thì né được, nhưng conversion kém hơn. Quyết định này đổi cả
   schema backend (xem `backend_progress.md` mục 4.2).
5. **Integration test** chạy trên simulator để phủ luồng bấm thật.
6. Pull-to-refresh cho carousel detail, trạng thái offline, lưu ZIP vào Photos.

---

## 6. Test trên máy thật

`integration_test/` chạy trên simulator/thiết bị với backend sống, bấm đúng như
người dùng. Widget test vẫn xanh trong khi màn hình chọn ảnh hỏng trên máy — nên
luồng nào quan trọng thì cần cả hai loại.

```bash
make test-mobile-device     # cần simulator + API đang chạy
```

Ba lần đo sai đáng nhớ khi viết test cho màn hình này:
`Scrollable.of(context)` trên chính GridView không tìm ra Scrollable; đếm
`InkWell` để đo số ảnh thì bão hoà vì grid tái sử dụng widget; và `tester.drag`
nuốt mất `kTouchSlop` nên delta đo được nhỏ hơn thực tế. Cách đúng: đọc
`ScrollPosition`, và dùng gesture thủ công khi cần đo chính xác.

## 6b. Chạy

```bash
make apigen                       # sinh lại client sau khi đổi struct Go
make mobile                       # flutter run
make test-mobile                  # 15 test

cd mobile && flutter run --dart-define=API_URL=https://api.example.com
```

Trên simulator iOS, `localhost` trỏ đúng vào máy host nên backend chạy local là
dùng được ngay.
