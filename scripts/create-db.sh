#!/bin/bash
set -e

# ==========================================
# Cloud SQL (PostgreSQL) setup for Twitch Crypto Donations
# Private IP with VPC Connector
# ==========================================

# Load .env file
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
else
    echo "❌ .env file not found. Please create one with DB vars."
    exit 1
fi

# Configuration
export GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT:-twitch-crypto-donations-core}
export REGION=${REGION:-us-central1}
export NETWORK_NAME=${NETWORK_NAME:-default}
export SQL_INSTANCE_NAME=${SQL_INSTANCE_NAME:-twitch-crypto-postgres}
export VPC_CONNECTOR_NAME=${VPC_CONNECTOR_NAME:-private-sql-connector}

echo "🧱 Setting up Cloud SQL (PostgreSQL)..."
echo "=========================================="
echo "Project: $GOOGLE_CLOUD_PROJECT"
echo "Region:  $REGION"
echo "DB Name: $POSTGRES_DB"
echo "User:    $POSTGRES_USER"
echo ""

# Ensure gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo "❌ gcloud CLI not found. Please install it:"
    echo "   https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Enable required APIs
echo "📦 Enabling required APIs..."
gcloud services enable \
    sqladmin.googleapis.com \
    servicenetworking.googleapis.com \
    compute.googleapis.com \
    vpcaccess.googleapis.com \
    --project=$GOOGLE_CLOUD_PROJECT

# 🔍 Detect or create VPC network
echo "🌐 Checking for VPC network..."
if ! gcloud compute networks describe $NETWORK_NAME --project=$GOOGLE_CLOUD_PROJECT &>/dev/null; then
    echo "⚠️  Network '$NETWORK_NAME' not found. Creating a new one..."
    gcloud compute networks create $NETWORK_NAME \
        --subnet-mode=auto \
        --project=$GOOGLE_CLOUD_PROJECT
else
    echo "ℹ️  Network '$NETWORK_NAME' already exists."
fi

# 🔌 Ensure Service Networking peering exists
echo "🔌 Checking or creating Service Networking peering..."
if ! gcloud services vpc-peerings list \
    --network=$NETWORK_NAME \
    --project=$GOOGLE_CLOUD_PROJECT \
    --format="value(name)" | grep -q "servicenetworking"; then

    echo "🧩 Creating Service Networking peering..."

    # Create address range (ignore if already exists)
    if ! gcloud compute addresses describe google-managed-services-$NETWORK_NAME \
        --global \
        --project=$GOOGLE_CLOUD_PROJECT &>/dev/null; then
        gcloud compute addresses create google-managed-services-$NETWORK_NAME \
            --global \
            --purpose=VPC_PEERING \
            --prefix-length=16 \
            --network=$NETWORK_NAME \
            --project=$GOOGLE_CLOUD_PROJECT
    else
        echo "ℹ️  Address range already exists."
    fi

    gcloud services vpc-peerings connect \
        --service=servicenetworking.googleapis.com \
        --network=$NETWORK_NAME \
        --ranges=google-managed-services-$NETWORK_NAME \
        --project=$GOOGLE_CLOUD_PROJECT
else
    echo "ℹ️  Service Networking peering already exists."
fi

# 🔗 Create or check VPC connector
echo "🔗 Checking or creating VPC connector..."
if ! gcloud compute networks vpc-access connectors describe $VPC_CONNECTOR_NAME \
    --region=$REGION \
    --project=$GOOGLE_CLOUD_PROJECT &>/dev/null; then

    echo "🧩 Creating VPC Access Connector '$VPC_CONNECTOR_NAME'..."
    gcloud compute networks vpc-access connectors create $VPC_CONNECTOR_NAME \
        --region=$REGION \
        --network=$NETWORK_NAME \
        --range=10.9.0.0/28 \
        --project=$GOOGLE_CLOUD_PROJECT

    # Wait for connector to be READY
    echo "⏳ Waiting for VPC connector to be ready..."
    for i in {1..60}; do
        STATE=$(gcloud compute networks vpc-access connectors describe $VPC_CONNECTOR_NAME \
            --region=$REGION \
            --project=$GOOGLE_CLOUD_PROJECT \
            --format="value(state)" 2>/dev/null || echo "ERROR")

        if [ "$STATE" = "READY" ]; then
            echo "✅ VPC connector is ready!"
            break
        elif [ "$STATE" = "ERROR" ]; then
            echo "❌ VPC connector creation failed"
            exit 1
        else
            echo "  [$i/60] Connector status: $STATE (waiting...)"
            sleep 5
        fi
    done
else
    STATE=$(gcloud compute networks vpc-access connectors describe $VPC_CONNECTOR_NAME \
        --region=$REGION \
        --project=$GOOGLE_CLOUD_PROJECT \
        --format="value(state)")

    if [ "$STATE" = "READY" ]; then
        echo "ℹ️  VPC Connector '$VPC_CONNECTOR_NAME' already exists and is READY."
    else
        echo "⚠️  VPC Connector exists but is in state: $STATE"
    fi
fi

# 🧰 Create Cloud SQL instance with PRIVATE IP ONLY
echo "🧰 Creating Cloud SQL instance (private IP only)..."
if ! gcloud sql instances describe $SQL_INSTANCE_NAME --project=$GOOGLE_CLOUD_PROJECT &>/dev/null; then
    gcloud sql instances create $SQL_INSTANCE_NAME \
        --database-version=POSTGRES_15 \
        --tier=db-custom-2-3840 \
        --region=$REGION \
        --root-password=$POSTGRES_PASSWORD \
        --network=projects/$GOOGLE_CLOUD_PROJECT/global/networks/$NETWORK_NAME \
        --no-assign-ip \
        --project=$GOOGLE_CLOUD_PROJECT
else
    echo "ℹ️  Instance '$SQL_INSTANCE_NAME' already exists."
fi

# 📗 Create database if missing
echo "📗 Creating database '$POSTGRES_DB' (if missing)..."
if ! gcloud sql databases describe $POSTGRES_DB \
    --instance=$SQL_INSTANCE_NAME \
    --project=$GOOGLE_CLOUD_PROJECT &>/dev/null; then
    gcloud sql databases create $POSTGRES_DB \
        --instance=$SQL_INSTANCE_NAME \
        --project=$GOOGLE_CLOUD_PROJECT
else
    echo "ℹ️  Database '$POSTGRES_DB' already exists."
fi

# 📡 Fetch private IP
echo ""
echo "📡 Fetching Cloud SQL private IP..."
CLOUD_SQL_IP=$(gcloud sql instances describe $SQL_INSTANCE_NAME \
    --project=$GOOGLE_CLOUD_PROJECT \
    --format="value(ipAddresses.ipAddress)")

if [ -z "$CLOUD_SQL_IP" ]; then
    echo "❌ Failed to fetch private IP. Check instance networking settings."
    exit 1
fi

# Get connection name for Cloud SQL Proxy
CONNECTION_NAME="$GOOGLE_CLOUD_PROJECT:$REGION:$SQL_INSTANCE_NAME"

gcloud sql users set-password twitch \
    --host=% \
    --instance=twitch-crypto-postgres \
    --password=$POSTGRES_PASSWORD \
    --project=twitch-crypto-donations-core

# ✅ Summary
echo ""
echo "=========================================="
echo "✅ Cloud SQL setup complete!"
echo "=========================================="
echo "📦 Instance name:    $SQL_INSTANCE_NAME"
echo "🔒 Private IP:       $CLOUD_SQL_IP"
echo "📗 Database:         $POSTGRES_DB"
echo "👤 User:             $POSTGRES_USER"
echo "🔑 Password:         $POSTGRES_PASSWORD"
echo "🔗 Connection Name:  $CONNECTION_NAME"
echo "🌐 VPC Connector:    $VPC_CONNECTOR_NAME"
echo ""
echo "💡 Next step: Run ./scripts/deploy.sh"
echo ""