#!/usr/bin/env bash
# Drives the MVP acceptance flow (§110, §146) against a running API + worker.
# Usage: ./scripts/smoke.sh
set -euo pipefail

API="${API:-http://localhost:8080}/api/v1"
JAR=$(mktemp)
WORKDIR=$(mktemp -d)
trap 'rm -rf "$JAR" "$WORKDIR"' EXIT

step() { printf "\n\033[1m%s\033[0m\n" "$1"; }
jqp()  { python3 -c "import json,sys; d=json.load(sys.stdin); print($1)"; }

step "1. Health"
curl -sf "${API%/api/v1}/healthz" | jqp "d['status']"

step "2. Dev login"
curl -sf -c "$JAR" -X POST "$API/auth/dev-login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"smoke@example.com","name":"Smoke"}' | jqp "d['email']"

step "3. Create project from a niche (AI onboarding)"
PROJECT=$(curl -sf -b "$JAR" -X POST "$API/projects" -H 'Content-Type: application/json' \
  -d '{"niche":"Tôi muốn xây kênh TikTok chia sẻ kiến thức chăm sóc trẻ sơ sinh 0-3 tuổi","language":"vi"}')
PID=$(echo "$PROJECT" | jqp "d['id']")
echo "$PROJECT" | jqp "d['name'] + ' — ' + d['niche']"

step "4. Dynamic form generated for this niche"
curl -sf -b "$JAR" "$API/projects/$PID/schema" | python3 -c "
import json,sys
d=json.load(sys.stdin)
print(f\"schema v{d['version']}\")
for p in d['form']['properties']:
    print(f\"  {p['name']:<14} {p['component']:<12} {'required' if p['required'] else ''}\")"

step "5. Generate a carousel"
GEN=$(curl -sf -b "$JAR" -X POST "$API/projects/$PID/carousels/generate" \
  -H 'Content-Type: application/json' -H "Idempotency-Key: smoke-$(date +%s)" \
  -d '{"inputs":{"topic":"5 dấu hiệu trẻ sơ sinh đang thiếu ngủ","angle":"mistakes","slide_count":5,"cta":"save"},"ratio":"4:5"}')
CID=$(echo "$GEN" | jqp "d['carousel_id']")
JOB=$(echo "$GEN" | jqp "d['job_id']")

for _ in $(seq 1 60); do
  sleep 2
  STATUS=$(curl -sf -b "$JAR" "$API/jobs/$JOB" | jqp "d['status'] + ' ' + str(d['progress']) + '% ' + d['current_step']")
  echo "  $STATUS"
  case "$STATUS" in
    completed*) break ;;
    failed*) echo "generation failed"; exit 1 ;;
  esac
done

step "6. Design JSON"
curl -sf -b "$JAR" "$API/carousels/$CID/design" | python3 -c "
import json,sys
d=json.load(sys.stdin)
print(f\"version {d['version']} · {d['design']['canvas']['width']}x{d['design']['canvas']['height']} · {len(d['design']['slides'])} slides\")
for s in d['design']['slides']:
    texts=[e for e in s['elements'] if e['type']=='text']
    img=[e for e in s['elements'] if e['type']=='image']
    print(f\"  {s['id']:<9} {s['role']:<9} {len(texts)} text, image={'yes' if img and img[0].get('asset_id') else 'no'}\")"

step "7. Image candidates for slide 2"
curl -sf -b "$JAR" "$API/carousels/$CID/slides/slide_2/images" | jqp "str(len(d['candidates'])) + ' candidates for ' + repr(d['query'])"

step "8. Edit + autosave (optimistic concurrency)"
curl -sf -b "$JAR" "$API/carousels/$CID/design" > "$WORKDIR/design.json"
python3 - "$WORKDIR" <<'PY'
import json,sys
w=sys.argv[1]
d=json.load(open(f"{w}/design.json"))
for e in d['design']['slides'][0]['elements']:
    if e['type']=='text':
        e['content']='Chỉnh sửa từ smoke test'
        break
json.dump({'version':d['version'],'design':d['design']}, open(f"{w}/patch.json",'w'))
PY
curl -sf -b "$JAR" -X PATCH "$API/carousels/$CID/design" -H 'Content-Type: application/json' \
  -d @"$WORKDIR/patch.json" | jqp "'saved as version ' + str(d['version'])"
CODE=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR" -X PATCH "$API/carousels/$CID/design" \
  -H 'Content-Type: application/json' -d @"$WORKDIR/patch.json")
echo "  stale write rejected with HTTP $CODE (expected 409)"
[ "$CODE" = "409" ] || { echo "optimistic concurrency is not working"; exit 1; }

step "9. Export ZIP"
EXID=$(curl -sf -b "$JAR" -X POST "$API/carousels/$CID/export" | jqp "d['id']")
for _ in $(seq 1 40); do
  sleep 2
  RES=$(curl -sf -b "$JAR" "$API/exports/$EXID")
  ST=$(echo "$RES" | jqp "d['status']")
  [ "$ST" = "completed" ] && break
  [ "$ST" = "failed" ] && { echo "export failed"; exit 1; }
done
URL=$(echo "$RES" | jqp "d['download_url']")
curl -sf -o "$WORKDIR/carousel.zip" "$URL"
unzip -l "$WORKDIR/carousel.zip" | awk 'NR>3 && $NF ~ /\.png$/ {printf "  %s (%s bytes)\n", $NF, $1}'

printf "\n\033[32mMVP flow passed end to end.\033[0m\n"
