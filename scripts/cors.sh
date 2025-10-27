#!/bin/bash
set -e

# ==========================================
# Configure CORS for Cloud Run Service
# Allows all origins, methods, and headers
# ==========================================

# Load .env
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

export GOOGLE_CLOUD_PROJECT=twitch-crypto-donations-core
export REGION=us-central1
export SERVICE_NAME=twitch-crypto-api

echo "🔧 Configuring CORS for Cloud Run Service"
echo "=========================================="
echo "Project: $GOOGLE_CLOUD_PROJECT"
echo "Region: $REGION"
echo "Service: $SERVICE_NAME"
echo ""

# Update the service with permissive CORS settings
echo "📝 Updating service with CORS annotations..."
gcloud run services update $SERVICE_NAME \
    --region=$REGION \
    --project=$GOOGLE_CLOUD_PROJECT \
    --update-annotations="run.googleapis.com/ingress=all" \
    --quiet

# Note: Cloud Run doesn't have native CORS configuration
# CORS must be handled at the application level
echo ""
echo "⚠️  IMPORTANT: Cloud Run CORS Configuration"
echo "=========================================="
echo ""
echo "Cloud Run does not have native CORS headers configuration."
echo "CORS must be handled in your application code."
echo ""
echo "=========================================="
echo ""
echo "✅ Service annotations updated"
echo "🔓 Ingress: Allow all traffic"
echo ""
echo "📋 Next steps:"
echo "   1. Add the CORS middleware to your Go application"
echo "   2. Rebuild and redeploy: ./scripts/deploy.sh"
echo "   3. Test CORS: curl -H 'Origin: https://example.com' -H 'Access-Control-Request-Method: POST' -H 'Access-Control-Request-Headers: Content-Type' -X OPTIONS $SERVICE_URL/api/register -v"
echo ""
echo "🌐 Service URL:"
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME \
    --region=$REGION \
    --project=$GOOGLE_CLOUD_PROJECT \
    --format='value(status.url)' 2>/dev/null || echo "Not deployed")
echo "   $SERVICE_URL"
echo ""