# Frontend — trạng thái và việc còn lại

Cập nhật: 2026-09-05 · Next.js 16 (App Router) · React 19 · Tailwind 4 · ~6.200 dòng · `frontend/`

---

## 1. Tóm tắt

Đủ 9 màn hình cho luồng MVP, cộng landing và pricing công khai.
TypeScript + ESLint sạch, production build OK, **13 e2e test Playwright pass**.

| Hạng mục | Trạng thái |
|---|---|
| Trang | landing, pricing, login, onboarding, projects, project, settings, create carousel, editor |
| State | TanStack Query (server) + Zustand (editor) — tách bạch theo §94 |
| Canvas | Fabric.js v6 qua `FabricCanvasAdapter` |
| i18n | vi/en, resolve từ cookie **trên server** nên first paint đã đúng |
| Test | 13 e2e (5 editor · 3 ngôn ngữ · 5 pricing). **Không có unit test.** |

---

## 2. Bản đồ

```
app/
  page.tsx                    landing (SaaS flow đầy đủ)
  pricing/                    bảng giá, chưa nối payment
  login/ onboarding/
  projects/                   danh sách project
  projects/[id]/              dashboard carousel
  projects/[id]/settings/     General · AI Strategy · Brand
  projects/[id]/new/          form động + tiến trình generate
  carousels/[id]/             editor

components/
  landing/    Chrome (nav/footer/Section/Faq) · SlideStack · DeepDives · PricingTable
  editor/     CanvasWorkspace · EditorSidebar · SlideNavigator · ImagePicker · CaptionPanel · useAutosave
  schema-form/DynamicForm     registry component đóng cho form AI sinh
  project/    AppShell · BrandPanel
  ui/         primitives · LocaleSwitch

lib/
  api/        client (typed) + hooks
  fabric/     adapter — NƠI DUY NHẤT chạm Fabric
  design/     measure.ts — bản sao thuật toán ngắt dòng của backend
  editor-store.ts  undo/redo + autosave state
  i18n/       vi · en · locale (server-safe) · index (provider)
  plans.ts    định nghĩa gói cho trang pricing
```

**Hai bất biến quan trọng:**

1. `lib/design/measure.ts` phải luôn khớp `backend/internal/fontkit`. Lệch nhau
   thì cảnh báo tràn chữ trong editor sẽ nói khác với file export (§157).
2. Chỉ `lib/fabric/adapter.ts` được import `fabric`. Mọi thứ khác làm việc với
   Design JSON ở toạ độ logic 1080×1350 (§98, §99).

---

## 3. Vấn đề đã biết

### 3.1 Bug đã sửa — giữ lại để biết vì sao code viết như vậy

**Object trùng lặp trên canvas (đã sửa).** `loadDesign` là async vì phải fetch
ảnh. Thứ tự cũ `clear() → await → add()` khiến hai lần load chồng nhau để lại
**hai object cho mỗi element**; `updateTextStyle` sửa bản thứ nhất còn bản thứ
hai vẽ đè lên, nên đổi font/màu/căn lề không thấy gì thay đổi. Đã sửa bằng cách
dựng object *trước* rồi mới `clear()`, cộng generation token để lần load cũ tự
bỏ cuộc. Test `exactly one canvas object per design element` canh chừng.

> Bài học: bug ở editor thường là bug **render**, không phải bug state. Type check
> và unit test không nhìn thấy. Test pixel thật là cách duy nhất.

### 3.2 Cần sửa

| Vấn đề | Chi tiết |
|---|---|
| **`DeepDives.tsx` hardcode tiếng Việt** | `CaptionVisual` có 2 dòng caption mẫu viết cứng tiếng Việt (`DeepDives.tsx:124` và `:126`), không qua i18n. Xem landing bằng tiếng Anh sẽ thấy tiếng Việt lẫn vào. |
| **Landing dùng ảnh picsum.photos** | 3 chỗ: `SlideStack.tsx` (3 slide hero), `DeepDives.tsx` (`CanvasVisual`), `app/page.tsx` (`UseCases`). Ảnh ngoài, seed cố định. Ổn cho dev, **phải thay bằng asset thật trước khi launch** — phụ thuộc bên thứ ba cho trang marketing là rủi ro. |
| **Không có error boundary** | Một lỗi render trong editor làm trắng cả trang, không có lối thoát. |
| **Undo/redo mất khi reload** | Lịch sử nằm trong memory Zustand. Backend đã snapshot `carousel_versions` nhưng frontend chưa dùng. |
| **Xung đột autosave xử lý thô** | 409 → tải lại bản server, **bỏ luôn sửa đổi cục bộ**. Đúng về dữ liệu nhưng user mất công. Nên báo rõ hoặc cho merge. |

### 3.3 Giới hạn đã biết

- **Editor không có layout mobile.** Cố định 3 cột (sidebar 320px · canvas ·
  rail slide 128px). Dưới ~1000px sẽ vỡ. Các trang khác đều responsive.
- **Không có unit test.** Chỉ có e2e, và e2e **cần API + worker + web đang chạy**.
  Chưa có test cho `measure.ts`, `editor-store.ts`, `plans.ts` — đều là logic
  thuần, dễ test, nên làm.
- **Chưa audit accessibility.** Có focus ring, aria-label ở vài chỗ, FAQ dùng
  `<details>` native. Chưa test bàn phím/screen reader toàn trang.
- **`/pricing` hoàn toàn client-side.** `lib/plans.ts` là nguồn sự thật; backend
  không biết gì về gói. Nút gói trả phí bị disable có chủ đích.
- **UI chỉ vi/en**, trong khi nội dung hỗ trợ 9 ngôn ngữ. Đây là chủ ý (§164):
  ngôn ngữ UI và ngôn ngữ nội dung độc lập.
- **`__tksCanvas` lộ ra ở non-production** để e2e đọc được object trên canvas.
  Có guard `NODE_ENV !== 'production'`.
- **Ảnh nền không kéo/chọn được trong editor** (`selectable: false`) — chủ ý,
  vị trí do design quyết định. Đổi ảnh qua tab Hình ảnh.

---

## 4. Việc tiếp theo, theo thứ tự ưu tiên

1. **i18n hoá `CaptionVisual`** — lỗi hiển thị thấy được ngay, sửa 5 phút.
2. **Thay ảnh picsum trên landing/pricing bằng asset thật.**
3. **Error boundary** cho editor và cho toàn app.
4. **Unit test** cho `measure.ts` (đối chiếu với ca test Go), `editor-store.ts`
   (undo/redo/history cap), `plans.ts` (giá và giảm 20%).
5. **Layout mobile/tablet cho editor**, hoặc chí ít một màn hình chặn lịch sự
   thay vì vỡ bố cục.
6. **UI cho version history** — backend đã có sẵn dữ liệu.
7. **Xử lý xung đột autosave tử tế hơn**: báo rõ đã tải lại và cho phép khôi
   phục bản nháp cục bộ.
8. **UI cho locked elements** (§46) — cần backend enforce trước (xem
   `backend_progress.md` mục 4.1).
9. **Audit accessibility** trước khi mở public.

---

## 5. Chạy và kiểm chứng

```bash
make web                     # :3000
make e2e-install             # tải Chromium, chạy 1 lần
make e2e                     # cần API + worker + web đang chạy
cd frontend && npx tsc --noEmit && npx eslint . && npm run build
```

**Lưu ý vận hành:** nếu `npm run dev` báo *"You can access the existing server…"*
mà thực ra không có gì chạy ở :3000, đó là lock cũ sót lại sau khi process bị
kill cứng. Xoá `frontend/.next/dev` rồi chạy lại.

---

## 6. Ghi chú về thiết kế

Landing và pricing dùng **chung chrome với app đã đăng nhập** (cùng palette
neutral, cùng `Button`/`Card`, cùng header) để đăng nhập xong không có cảm giác
bước sang sản phẩm khác.

**Cố ý không có testimonial hay logo khách hàng** — bịa ra là social proof giả.
Phần proof dùng số liệu kiểm chứng được của chính sản phẩm: 9 ngôn ngữ, 3 khổ
ảnh, 8 công thức, export không watermark.

Landing chỉ có **một** chuỗi animation lúc load (câu mô tả tự gõ rồi ba slide
bay vào), tôn trọng `prefers-reduced-motion`. Không thêm hiệu ứng hover hay
fade-in theo scroll ở các section.
