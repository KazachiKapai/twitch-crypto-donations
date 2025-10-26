#!/bin/bash
# scripts/fix-all-permissions.sh
# One-time setup to grant all necessary permissions to Cloud Build

set -e

export PROJECT_ID=twitch-crypto-donations-core
export PROJECT_NUMBER=612003413257
export CLOUDBUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"
export COMPUTE_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║  Fixing Cloud Build Permissions for Twitch Crypto API     ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "Project: $PROJECT_ID"
echo "Cloud Build SA: $CLOUDBUILD_SA"
echo ""

# 1. Storage permissions (for uploading source code)
echo "📦 [1/4] Granting Storage Admin permissions..."
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${CLOUDBUILD_SA}" \
    --role="roles/storage.admin" \
    --quiet 2>/dev/null || true

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${COMPUTE_SA}" \
    --role="roles/storage.admin" \
    --quiet 2>/dev/null || true

echo "   ✅ Storage permissions granted"

# 2. Artifact Registry permissions (for pushing Docker images)
echo "📦 [2/4] Granting Artifact Registry Writer permissions..."
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${CLOUDBUILD_SA}" \
    --role="roles/artifactregistry.writer" \
    --quiet 2>/dev/null || true

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${COMPUTE_SA}" \
    --role="roles/artifactregistry.writer" \
    --quiet 2>/dev/null || true

echo "   ✅ Artifact Registry permissions granted"

# 3. Cloud Build permissions
echo "🏗️  [3/4] Granting Cloud Build Builder permissions..."
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${CLOUDBUILD_SA}" \
    --role="roles/cloudbuild.builds.builder" \
    --quiet 2>/dev/null || true

echo "   ✅ Cloud Build permissions granted"

# 4. Logging permissions
echo "📝 [4/4] Granting Logging Writer permissions..."
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${CLOUDBUILD_SA}" \
    --role="roles/logging.logWriter" \
    --quiet 2>/dev/null || true

echo "   ✅ Logging permissions granted"

# Grant Cloud SQL Client role
gcloud projects add-iam-policy-binding twitch-crypto-donations-core \
    --member="serviceAccount:612003413257-compute@developer.gserviceaccount.com" \
    --role="roles/cloudsql.client"