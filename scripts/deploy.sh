#!/bin/bash
set -e

# ==========================================
# Twitch Crypto Donations API Deployment
# Uses VPC Connector + Cloud SQL Proxy
# ==========================================

# Load .env
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

export GOOGLE_CLOUD_PROJECT=twitch-crypto-donations-core
export REPOSITORY=twitch-crypto-donations-core
export REGION=us-central1
export SERVICE_NAME=twitch-crypto-api
export SQL_INSTANCE_NAME=twitch-crypto-postgres
export VPC_CONNECTOR_NAME=private-sql-connector
export VERSION=${1:-latest}

# Cloud SQL connection name
export CONNECTION_NAME="$GOOGLE_CLOUD_PROJECT:$REGION:$SQL_INSTANCE_NAME"

echo "🚀 Deploying Twitch Crypto Donations API"
echo "=========================================="
echo "Project: $GOOGLE_CLOUD_PROJECT"
echo "Region: $REGION"
echo "Version: $VERSION"
echo ""

# Enable services
echo "📦 Enabling services..."
gcloud services enable \
    artifactregistry.googleapis.com \
    run.googleapis.com \
    cloudbuild.googleapis.com \
    sqladmin.googleapis.com \
    vpcaccess.googleapis.com \
    --project=$GOOGLE_CLOUD_PROJECT \
    --quiet

# Build image
echo "🏗️  Building Docker image..."
gcloud builds submit \
    --tag=$REGION-docker.pkg.dev/$GOOGLE_CLOUD_PROJECT/$REPOSITORY/api:$VERSION \
    --project=$GOOGLE_CLOUD_PROJECT

echo "✅ Image built successfully"

# Check VPC connector status
echo "🔍 Checking VPC connector status..."
VPC_STATE=$(gcloud compute networks vpc-access connectors describe $VPC_CONNECTOR_NAME \
    --region=$REGION \
    --project=$GOOGLE_CLOUD_PROJECT \
    --format="value(state)" 2>/dev/null || echo "NOT_FOUND")

if [ "$VPC_STATE" != "READY" ]; then
    echo "❌ VPC connector is not ready (state: $VPC_STATE)"
    echo "   Please run ./scripts/create-db.sh first to create the VPC connector"
    exit 1
fi

echo "✅ VPC connector is ready"

# Deploy using VPC Connector + Cloud SQL Proxy
echo "🚀 Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
    --image=$REGION-docker.pkg.dev/$GOOGLE_CLOUD_PROJECT/$REPOSITORY/api:$VERSION \
    --region=$REGION \
    --platform=managed \
    --allow-unauthenticated \
    --port=8080 \
    --memory=512Mi \
    --cpu=1 \
    --timeout=300 \
    --max-instances=10 \
    --min-instances=0 \
    --vpc-connector=$VPC_CONNECTOR_NAME \
    --add-cloudsql-instances=$CONNECTION_NAME \
    --set-env-vars="\
APP_ENV=$APP_ENV,\
DB_HOST=/cloudsql/$CONNECTION_NAME,\
DB_PORT=$DB_PORT,\
POSTGRES_USER=$POSTGRES_USER,\
POSTGRES_PASSWORD=$POSTGRES_PASSWORD,\
POSTGRES_DB=$POSTGRES_DB,\
DB_SSLMODE=$DB_SSLMODE,\
POSTGRES_MIGRATIONS_DIR=$POSTGRES_MIGRATIONS_DIR,\
SWAGGER_PATH=$SWAGGER_PATH,\
OBS_SERVICE_DOMAIN=$OBS_SERVICE_DOMAIN,\
HTTP_LISTEN_PORT=$HTTP_LISTEN_PORT,\
ROUTE_PREFIX=$ROUTE_PREFIX, \
JWT_SECRET=$JWT_SECRET,
JWT_TOKEN_EXPIRATION_HOURS=$JWT_TOKEN_EXPIRATION_HOURS" \
    --project=$GOOGLE_CLOUD_PROJECT

# Get service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME \
    --region=$REGION \
    --project=$GOOGLE_CLOUD_PROJECT \
    --format='value(status.url)')

echo ""
echo "=========================================="
echo "✅ Deployment complete!"
echo "=========================================="
echo ""
echo "🌐 Service URL: $SERVICE_URL"
echo ""
echo "📝 Test endpoints:"
echo "   Health:    curl $SERVICE_URL/health"
echo "   Swagger:   $SERVICE_URL/swagger/index.html"
echo "   Register:  curl -X POST $SERVICE_URL/api/register -H 'Content-Type: application/json' -d '{\"wallet\":\"test-wallet-123\"}'"
echo ""
echo "📊 View logs:"
echo "   gcloud run services logs tail $SERVICE_NAME --region=$REGION --project=$GOOGLE_CLOUD_PROJECT"
echo ""
echo "💡 Using VPC Connector + Cloud SQL Proxy:"
echo "   VPC Connector: $VPC_CONNECTOR_NAME"
echo "   Cloud SQL:     $CONNECTION_NAME"
echo "   Socket:        /cloudsql/$CONNECTION_NAME"
echo ""