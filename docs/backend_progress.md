# Backend — trạng thái và việc còn lại

Cập nhật: 2026-09-05 · Go 1.27 · ~9.500 dòng · `backend/`

---

## 1. Tóm tắt

Toàn bộ pipeline MVP theo [`PRODUCT_TECHNICAL_PLAN.md`](PRODUCT_TECHNICAL_PLAN.md) đã chạy end-to-end:
login → project → dynamic schema → generate (async) → design → images → autosave
→ export ZIP. `make smoke` đi hết luồng này mỗi lần chạy.

**Chạy được không cần API key.** `AI_API_KEY` rỗng → mock provider deterministic;
`UNSPLASH_ACCESS_KEY` rỗng → picsum stub. Mọi lớp validation, retry, worker,
export đều đi qua đúng code path thật.

| Hạng mục | Trạng thái |
|---|---|
| Test | 35 test, 7/7 package pass |
| Migrations | `0001_init.sql`, `0002_caption_and_brand.sql` — chạy tự động lúc boot |
| Binaries | `cmd/server` (HTTP), `cmd/worker` (queue) — chung `internal/app` |

---

## 2. Bản đồ package

```
internal/
  app         wiring + router
  auth        session server-side, Google OAuth, dev login
  project     identity, AI context có version, brand kit, schema versions
  schema      trình biên dịch an toàn cho AI-schema + validate input
  carousel    CRUD, content/design versions, optimistic concurrency, caption
  ai          provider, prompt có version, task có kiểu, retry, audit, mock
  content     hợp đồng nội dung + cổng kiểm duyệt ngữ nghĩa
  design      Design JSON + validator bố cục + brand marks
  fontkit     ENGINE đo chữ duy nhất (dùng chung validator + exporter)
  language    registry ngôn ngữ nội dung (9 ngôn ngữ hệ Latin)
  image       provider abstraction, Unsplash compliance, cache tìm kiếm
  asset       validate upload, MinIO, signed URL
  export      renderer Go native, PNG, ZIP, thumbnail
  job         state machine, checkpoint, Redis queue
  worker      pipeline generation + export
  research    abstraction search provider
  registry    fonts + formulas + languages
  ratelimit   quota theo ngày cho mỗi user
```

**Bất biến quan trọng:** `internal/fontkit` là implementation *duy nhất* của đo
chữ và ngắt dòng. Design validator, export renderer và
`frontend/lib/design/measure.ts` đều dùng cùng thuật toán. Đây là thứ giữ cho
WYSIWYG không vỡ (§155, §157). Đừng bao giờ viết đo chữ ở chỗ khác.

---

## 3. API hiện có

```
GET    /healthz
GET    /api/v1/auth/config | /auth/google | /auth/callback | /auth/me
POST   /api/v1/auth/dev-login | /auth/logout

GET    /projects                       POST   /projects
GET    /projects/{id}                  PATCH  /projects/{id}      DELETE /projects/{id}
GET    /projects/{id}/schema           POST   /projects/{id}/generate-schema
GET    /projects/{id}/carousels        POST   /projects/{id}/carousels/generate

GET    /carousels/{id}                 PATCH  /carousels/{id}     DELETE /carousels/{id}
POST   /carousels/{id}/duplicate | /archive | /regenerate | /apply-brand
GET    /carousels/{id}/content | /sources
GET    /carousels/{id}/design          PATCH  /carousels/{id}/design
GET    /carousels/{id}/caption         PATCH  /carousels/{id}/caption
POST   /carousels/{id}/caption/generate
POST   /carousels/{id}/export          GET    /exports/{id}

GET    /carousels/{id}/slides/{slideId}/images
POST   /carousels/{id}/slides/{slideId}/images/search | /images/select

POST   /assets/upload                  GET /assets/{id}   DELETE /assets/{id}
GET    /jobs/{id}                      GET /registry
```

8 AI task: `project_analysis`, `carousel_schema`, `research`,
`content_generation`, `formula_selection`, `design_generation`, `image_intent`,
`caption_generation`.

---

## 4. Vấn đề đã biết

### 4.1 Dead code / chưa nối dây — nên xử lý sớm

| Vấn đề | Chi tiết |
|---|---|
| **`job.Repo.ResetForRetry` không ai gọi** | Checkpoint (`last_completed_step`, `step_outputs`) đã ghi đầy đủ theo §158, nhưng **không có đường retry tự động**. Job fail thì nằm im; user phải bấm *Tạo lại* và chạy lại từ đầu, mất toàn bộ tiền AI đã tiêu. Đây là gap lớn nhất hiện tại. |
| **`design.Element.Locked` không được backend đọc** | Chỉ là field trong struct. Validator không kiểm tra, và `regenerate` dựng design mới hoàn toàn nên **phần tử bị khoá vẫn bị ghi đè** — trái với §46. Frontend có tôn trọng `locked` (không cho chọn/kéo), nên hiện trạng là "khoá ở UI, không khoá ở AI". |
| **Session hết hạn không được dọn** | Chỉ xoá khi logout. Bảng `sessions` sẽ phình. Cần một job dọn định kỳ. |
| **`image_searches` không bao giờ prune** | Cache đọc trong 7 ngày nhưng row cũ nằm lại vĩnh viễn. |

### 4.2 Chưa có, cần trước khi mở bán

| Vấn đề | Chi tiết |
|---|---|
| **Không có khái niệm gói (plan)** | UI `/pricing` hiển thị Free/Pro/Max nhưng **backend hoàn toàn không biết**: không có cột `plan` trên `users`, không enforce giới hạn theo gói. `ratelimit` chỉ có quota cố định theo ngày cho mọi user (§163): generation 20, image search 30, export 20, upload 100. Muốn bán thật phải: thêm `plan` + `subscriptions`, đổi `ratelimit` thành theo gói, và đếm theo tháng chứ không theo ngày. |
| **Chưa apply Unsplash Production access** | §156 cảnh báo: demo app chỉ 50 request/giờ, production ~5000. Duyệt mất thời gian và **có thể chặn launch**. Download-trigger và attribution đã implement đúng. |
| **Không có theo dõi chi phí** | `ai_generations` lưu token usage nhưng chưa quy ra tiền, chưa tổng hợp theo user (§88). `usage_events` đã ghi ở 5 chỗ nhưng chưa ai đọc. |
| **Không có metric / trace** | Chỉ có structured log + request_id. Đủ để debug thủ công, không đủ để biết tỉ lệ generation fail hay p95 latency. |

### 4.3 Giới hạn đã biết, chấp nhận được ở MVP

- **Research chỉ hỗ trợ Tavily.** `research.SearchProvider` là interface nhưng
  mới có một implementation. Không có `SEARCH_API_KEY` thì không có search thật,
  và prompt yêu cầu model trả `sources` rỗng thay vì bịa URL — đúng nhưng nghĩa
  là mục "Research sources" luôn trống.
- **Mock provider là fixture, không phải model.** Kết quả offline không phản ánh
  chất lượng thật. Đừng đánh giá chất lượng nội dung bằng mock.
- **`design_generation` là prompt mong manh nhất** (toạ độ tuyệt đối + ràng buộc
  safe area). Model yếu sẽ bị reject nhiều lần rồi fail. Xem log worker để biết
  lý do reject.
- **Export chỉ PNG + ZIP.** Chưa có JPG/PDF (§112).
- **Chưa có regenerate từng slide** (§45 V1), chưa có API restore version dù
  `carousel_versions` đã snapshot mỗi lần generate.
- **Ngôn ngữ nội dung giới hạn ở 9 ngôn ngữ hệ Latin** vì font trong `fontkit`.
  Thêm tiếng Thái/Nhật/Hàn/Trung **phải thêm font trước** — test
  `TestEveryRegisteredFontCoversEveryOfferedLanguage` sẽ fail nếu quên.
- **Brand marks chỉ áp lúc generate.** Đổi brand kit không tự cập nhật carousel
  cũ; có `POST /apply-brand` để áp thủ công.
- **CSRF dựa vào SameSite=Lax + CORS credentialed**, chưa có CSRF token riêng.
  Đủ cho MVP nhưng nên rà lại trước khi mở public.

---

## 5. Việc tiếp theo, theo thứ tự ưu tiên

1. **Nối dây retry cho generation job.** `ResetForRetry` + đọc checkpoint trong
   `worker.Pipeline.Run` để resume từ `last_completed_step`. Toàn bộ hạ tầng đã
   sẵn, chỉ thiếu đoạn nối. Tiết kiệm trực tiếp chi phí AI.
2. **Enforce locked elements ở backend.** Khi regenerate, giữ nguyên element có
   `locked: true` từ design cũ thay vì dựng lại toàn bộ (§46).
3. **Apply Unsplash Production access.** Việc hành chính, làm càng sớm càng tốt.
4. **Thêm plan + subscription.** Cột `plan` trên `users`, `ratelimit` theo gói và
   theo tháng, endpoint trả gói hiện tại cho frontend.
5. **Job dọn dẹp định kỳ:** session hết hạn, `image_searches` cũ, export cũ trong
   MinIO.
6. **Prompt eval harness (§161).** Đã có prompt versioned + audit đầy đủ trong
   `ai_generations` — đây là tiền đề. Không có nó thì mỗi lần sửa prompt là một
   canh bạc không đo được.
7. **Metrics:** tỉ lệ generation thành công, latency từng step, token/chi phí
   theo user.
8. **Regenerate từng slide** và **restore version** — hai thứ user sẽ hỏi sớm.

---

## 6. Chạy và kiểm chứng

```bash
make infra     # postgres, redis, minio
make api       # :8080, migrations chạy lúc boot
make worker
make test      # 35 test Go
make smoke     # đi hết luồng MVP trên stack đang chạy
make golden    # dựng lại ảnh golden của export renderer
```

**Lưu ý vận hành:** `go run ./cmd/worker` chạy binary trong build cache với argv
khác hẳn, nên `pkill -f "backend/bin/worker"` **không** bắt được nó. Đã có một
worker cũ sống sót 4 tiếng và tranh job. Cách kiểm tra chắc chắn:

```bash
docker exec tks_redis redis-cli CLIENT LIST | grep -c brpop   # 0 = không còn worker
```
